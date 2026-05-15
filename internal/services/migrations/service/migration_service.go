package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/migrations/seed"
	"gorm.io/gorm"
)

type MigrationStatus struct {
	Version uint `json:"version"`
	Dirty   bool `json:"dirty"`
}

type MigrationService struct {
	cfg    *config.Config
	db     *gorm.DB
	logger *slog.Logger
}

func New(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *MigrationService {
	return &MigrationService{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

func (s *MigrationService) Up(ctx context.Context, steps int) (MigrationStatus, error) {
	return s.runMigration(ctx, func(m *migrate.Migrate) error {
		if steps > 0 {
			return m.Steps(steps)
		}
		return m.Up()
	})
}

func (s *MigrationService) Down(ctx context.Context, steps int) (MigrationStatus, error) {
	if steps <= 0 {
		return MigrationStatus{}, errors.New("steps must be positive")
	}

	return s.runMigration(ctx, func(m *migrate.Migrate) error {
		return m.Steps(-steps)
	})
}

func (s *MigrationService) Status(ctx context.Context) (MigrationStatus, error) {
	m, err := s.newMigrator()
	if err != nil {
		return MigrationStatus{}, err
	}
	defer s.closeMigrator(m)

	return s.currentStatus(m)
}

func (s *MigrationService) Seed(ctx context.Context) error {
	return seed.IAM(ctx, s.db, s.cfg, s.logger)
}

func (s *MigrationService) runMigration(ctx context.Context, run func(*migrate.Migrate) error) (MigrationStatus, error) {
	m, err := s.newMigrator()
	if err != nil {
		return MigrationStatus{}, err
	}
	defer s.closeMigrator(m)

	if err := run(m); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return MigrationStatus{}, err
	}

	return s.currentStatus(m)
}

func (s *MigrationService) newMigrator() (*migrate.Migrate, error) {
	sqlDB, err := s.db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres driver: %w", err)
	}

	sourceURL, err := s.migrationSourceURL()
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	return m, nil
}

func (s *MigrationService) migrationSourceURL() (string, error) {
	path := s.cfg.MigrationsPath
	if path == "" {
		path = "migrations"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve migrations path: %w", err)
	}

	return "file://" + filepath.ToSlash(absPath), nil
}

func (s *MigrationService) currentStatus(m *migrate.Migrate) (MigrationStatus, error) {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return MigrationStatus{Version: 0, Dirty: false}, nil
		}
		return MigrationStatus{}, err
	}

	return MigrationStatus{Version: version, Dirty: dirty}, nil
}

func (s *MigrationService) closeMigrator(m *migrate.Migrate) {
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		s.logger.Error("failed to close migration source", "error", sourceErr)
	}
	if dbErr != nil {
		s.logger.Error("failed to close migration db", "error", dbErr)
	}
}

