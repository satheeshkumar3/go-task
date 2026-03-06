package usecase

import (
    "context"
    "errors"
    "time"

    "task-manager/internal/domain"
    "task-manager/internal/repository/interfaces"
    "task-manager/internal/pkg/validator"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
    userRepo      interfaces.UserRepository
    jwtSecret     string
    tokenDuration time.Duration
}

type RegisterInput struct {
    Email    string `validate:"required,email"`
    Username string `validate:"required,min=3,max=50"`
    Password string `validate:"required,min=8"`
}

type LoginInput struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required"`
}

type UpdateProfileInput struct {
    Email    string `json:"email" validate:"omitempty,email"`
    Username string `json:"username" validate:"omitempty,min=3,max=50"`
}

type AuthResult struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    User         *domain.User
}

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func NewAuthUseCase(userRepo interfaces.UserRepository, jwtSecret string, tokenDuration time.Duration) *AuthUseCase {
    return &AuthUseCase{
        userRepo:      userRepo,
        jwtSecret:     jwtSecret,
        tokenDuration: tokenDuration,
    }
}

func (uc *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
    // Validate input
    if err := validator.Validate(input); err != nil {
        return nil, err
    }

    // Check if user exists
    existingUser, _ := uc.userRepo.GetByEmail(ctx, input.Email)
    if existingUser != nil {
        return nil, errors.New("email already registered")
    }

    existingUser, _ = uc.userRepo.GetByUsername(ctx, input.Username)
    if existingUser != nil {
        return nil, errors.New("username already taken")
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // Create user
    user := &domain.User{
        Email:    input.Email,
        Username: input.Username,
        Password: string(hashedPassword),
    }

    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    // Generate tokens
    return uc.generateTokens(user)
}

func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
    // Validate input
    if err := validator.Validate(input); err != nil {
        return nil, err
    }

    // Get user
    user, err := uc.userRepo.GetByEmail(ctx, input.Email)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("invalid credentials")
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        return nil, errors.New("invalid credentials")
    }

    // Generate tokens
    return uc.generateTokens(user)
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
    // Parse and validate refresh token
    token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(uc.jwtSecret), nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid refresh token")
    }

    // Get user
    user, err := uc.userRepo.GetByID(ctx, claims.UserID)
    if err != nil || user == nil {
        return nil, errors.New("user not found")
    }

    // Generate new tokens
    return uc.generateTokens(user)
}

func (uc *AuthUseCase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
    user, err := uc.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not found")
    }
    return user, nil
}

func (uc *AuthUseCase) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*domain.User, error) {
    // Validate input
    if err := validator.Validate(input); err != nil {
        return nil, err
    }

    // Get user
    user, err := uc.userRepo.GetByID(ctx, userID)
    if err != nil || user == nil {
        return nil, errors.New("user not found")
    }

    // Update fields
    if input.Email != "" && input.Email != user.Email {
        // Check if email is taken
        existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
        if existing != nil {
            return nil, errors.New("email already taken")
        }
        user.Email = input.Email
    }

    if input.Username != "" && input.Username != user.Username {
        // Check if username is taken
        existing, _ := uc.userRepo.GetByUsername(ctx, input.Username)
        if existing != nil {
            return nil, errors.New("username already taken")
        }
        user.Username = input.Username
    }

    // Save
    if err := uc.userRepo.Update(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (uc *AuthUseCase) generateTokens(user *domain.User) (*AuthResult, error) {
    // Create access token
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
        UserID: user.ID,
        Email:  user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.tokenDuration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    })

    accessTokenString, err := accessToken.SignedString([]byte(uc.jwtSecret))
    if err != nil {
        return nil, err
    }

    // Create refresh token (longer expiration)
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
        UserID: user.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.tokenDuration * 7)), // 7 days
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    })

    refreshTokenString, err := refreshToken.SignedString([]byte(uc.jwtSecret))
    if err != nil {
        return nil, err
    }

    return &AuthResult{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresAt:    time.Now().Add(uc.tokenDuration),
        User:         user,
    }, nil
}