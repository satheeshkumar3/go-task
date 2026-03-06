package validator

import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
    
    // Register custom validations
    validate.RegisterValidation("notpast", validateNotPast)
}

// Validate validates a struct
func Validate(i interface{}) error {
    if err := validate.Struct(i); err != nil {
        var errors []string
        for _, err := range err.(validator.ValidationErrors) {
            errors = append(errors, formatError(err))
        }
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    return nil
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
    if err := validate.Var(field, tag); err != nil {
        var errors []string
        for _, err := range err.(validator.ValidationErrors) {
            errors = append(errors, formatError(err))
        }
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    return nil
}

func formatError(err validator.FieldError) string {
    switch err.Tag() {
    case "required":
        return err.Field() + " is required"
    case "email":
        return err.Field() + " must be a valid email"
    case "min":
        return err.Field() + " must be at least " + err.Param() + " characters"
    case "max":
        return err.Field() + " must be at most " + err.Param() + " characters"
    case "oneof":
        return err.Field() + " must be one of: " + err.Param()
    case "uuid":
        return err.Field() + " must be a valid UUID"
    case "notpast":
        return err.Field() + " cannot be in the past"
    default:
        return err.Field() + " is invalid"
    }
}

// Custom validator: check if date is not in the past
func validateNotPast(fl validator.FieldLevel) bool {
    date, ok := fl.Field().Interface().(string)
    if !ok {
        return true
    }
    
    // Parse date and check if it's in the past
    // Implementation depends on your date format
    return true
}