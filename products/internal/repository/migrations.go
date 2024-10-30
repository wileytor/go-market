package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
)

func Migrations(dbAddr, migrationsPath string, zlog *zerolog.Logger) error {
	// Проверка существования директории миграций
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations path does not exist: %s", migrationsPath)
	}

	absolutePath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}
	migratePath := fmt.Sprintf("file://%s", absolutePath)

	m, err := migrate.New(migratePath,  dbAddr)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			zlog.Debug().Msg("No migrations apply")
			return nil
		}
		return err
	}

	zlog.Debug().Msg("Migrate complete")
	return nil
}
