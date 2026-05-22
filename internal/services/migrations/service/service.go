package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	coredb "github.com/imama2/Genzite-Backend/internal/core/db"
	"github.com/imama2/Genzite-Backend/internal/services/migrations/seed"
	_ "github.com/lib/pq"
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

var migrationFolders = []string{"schemas", "tables", "seeders"}

func New(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *MigrationService {
	return &MigrationService{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

func (s *MigrationService) Up(ctx context.Context, steps int) (MigrationStatus, error) {
	if steps > 0 {
		return s.runStepsUp(ctx, steps)
	}

	var lastStatus MigrationStatus
	for _, folder := range migrationFolders {
		status, err := s.runMigrationInFolder(ctx, folder, func(m *migrate.Migrate) error {
			return m.Up()
		})
		if err != nil {
			return MigrationStatus{}, err
		}
		lastStatus = status
	}

	return lastStatus, nil
}

func (s *MigrationService) Down(ctx context.Context, steps int) (MigrationStatus, error) {
	if steps <= 0 {
		return MigrationStatus{}, errors.New("steps must be positive")
	}

	current, err := s.statusFromFolder(migrationFolders[0])
	if err != nil {
		return MigrationStatus{}, err
	}

	folder, err := s.folderForVersion(current.Version)
	if err != nil {
		return MigrationStatus{}, err
	}

	return s.runMigrationInFolder(ctx, folder, func(m *migrate.Migrate) error {
		return m.Steps(-steps)
	})
}

func (s *MigrationService) Status(ctx context.Context) (MigrationStatus, error) {
	_ = ctx
	return s.statusFromFolder(migrationFolders[0])
}

func (s *MigrationService) Seed(ctx context.Context) error {
	return seed.IAM(ctx, s.db, s.cfg, s.logger)
}

func (s *MigrationService) runStepsUp(ctx context.Context, steps int) (MigrationStatus, error) {
	_ = ctx
	current, err := s.statusFromFolder(migrationFolders[0])
	if err != nil {
		return MigrationStatus{}, err
	}

	startFolder, err := s.folderForVersion(current.Version)
	if err != nil {
		return MigrationStatus{}, err
	}

	startIndex := 0
	for i, folder := range migrationFolders {
		if folder == startFolder {
			startIndex = i
			break
		}
	}

	var lastStatus MigrationStatus
	for i := startIndex; i < len(migrationFolders); i++ {
		before, err := s.statusFromFolder(migrationFolders[i])
		if err != nil {
			return MigrationStatus{}, err
		}
		status, err := s.runMigrationInFolder(ctx, migrationFolders[i], func(m *migrate.Migrate) error {
			return m.Steps(steps)
		})
		if err != nil {
			return MigrationStatus{}, err
		}
		lastStatus = status
		if status.Version != before.Version || status.Dirty != before.Dirty {
			return status, nil
		}
	}

	return lastStatus, nil
}

func (s *MigrationService) runMigrationInFolder(ctx context.Context, folder string, run func(*migrate.Migrate) error) (MigrationStatus, error) {
	_ = ctx
	m, err := s.newMigrator(folder)
	if err != nil {
		return MigrationStatus{}, err
	}
	defer s.closeMigrator(m)

	if err := run(m); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return MigrationStatus{}, err
	}

	return s.currentStatus(m)
}

func (s *MigrationService) newMigrator(folder string) (*migrate.Migrate, error) {
	dsn, err := coredb.BuildDatabaseDSN(s.cfg)
	if err != nil {
		return nil, err
	}
	fmt.Println(dsn)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open migration database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping migration database: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("create postgres driver: %w", err)
	}

	sourceURL, err := s.migrationSourceURL(folder)
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	return m, nil
}

func (s *MigrationService) migrationSourceURL(folder string) (string, error) {
	absPath, err := s.migrationSourcePath(folder)
	if err != nil {
		return "", err
	}

	return "file://" + filepath.ToSlash(absPath), nil
}

func (s *MigrationService) migrationSourcePath(folder string) (string, error) {
	path := s.cfg.MigrationsPath
	if path == "" {
		path = "migrations"
	}

	absPath, err := filepath.Abs(filepath.Join(path, folder))
	if err != nil {
		return "", fmt.Errorf("resolve migrations path: %w", err)
	}

	return absPath, nil
}

func (s *MigrationService) statusFromFolder(folder string) (MigrationStatus, error) {
	m, err := s.newMigrator(folder)
	if err != nil {
		return MigrationStatus{}, err
	}
	defer s.closeMigrator(m)

	return s.currentStatus(m)
}

func (s *MigrationService) folderForVersion(version uint) (string, error) {
	if version == 0 {
		return migrationFolders[0], nil
	}

	for _, folder := range migrationFolders {
		versions, err := s.sourceVersions(folder)
		if err != nil {
			return "", err
		}
		if _, ok := versions[version]; ok {
			return folder, nil
		}
	}

	return "", fmt.Errorf("migration version %d not found in schemas/tables/seeders", version)
}

func (s *MigrationService) sourceVersions(folder string) (map[uint]struct{}, error) {
	dirPath, err := s.migrationSourcePath(folder)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("read migrations folder: %w", err)
	}

	versions := make(map[uint]struct{})
	re := regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		parsed, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			continue
		}
		versions[uint(parsed)] = struct{}{}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no migrations found in %s", dirPath)
	}

	return versions, nil
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
