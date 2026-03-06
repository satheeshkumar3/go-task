package database

import (
    "context"
    "log"
    "time"

    "task-manager/internal/domain"
    "task-manager/configs"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/plugin/prometheus"
)

type Database struct {
    *gorm.DB
}

func NewPostgresConnection(config *configs.DatabaseConfig) (*Database, error) {
    // Configure GORM logger
    gormLogger := logger.New(
        log.Default(),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logger.Info,
            IgnoreRecordNotFoundError: true,
            Colorful:                  false,
        },
    )

    // Open connection
    db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
        Logger: gormLogger,
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    })
    if err != nil {
        return nil, err
    }

    // Get generic database object
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // Configure connection pool
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(config.MaxLifetime)

    // Add Prometheus metrics
    if err := db.Use(prometheus.New(prometheus.Config{
        DBName:          "taskmanager",
        RefreshInterval: 15,
        MetricsCollector: []prometheus.MetricsCollector{
            &prometheus.Postgres{VariableNames: []string{"threads_connected", "max_connections"}},
        },
    })); err != nil {
        log.Printf("Failed to setup Prometheus: %v", err)
    }

    return &Database{db}, nil
}

// AutoMigrate runs database migrations
func (db *Database) AutoMigrate() error {
    return db.DB.AutoMigrate(
        &domain.User{},
        &domain.Task{},
        &domain.Tag{},
    )
}

// WithTransaction executes a function within a database transaction
func (db *Database) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    return db.DB.Transaction(func(tx *gorm.DB) error {
        return fn(tx)
    })
}

// Health checks database connection
func (db *Database) Health(ctx context.Context) error {
    sqlDB, err := db.DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.PingContext(ctx)
}