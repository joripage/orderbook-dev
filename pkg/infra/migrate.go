package infra

import (
	"fmt"
	"sync"

	"github.com/cenkalti/backoff"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	postgres_wrapper "github.com/joripage/orderbook-dev/pkg/infra/postgres"
	"gorm.io/gorm"
)

// IMigrateTool tool to migrate schema and data.
type IMigrateTool interface {
	// Create test db for unit test
	CreateDBAndMigrate(cfg *postgres_wrapper.PostgresConfig, migrationFile string) *gorm.DB

	// Migrate from current version to latest verion.
	Migrate(source string, connStr string)
}

type migrateTool struct{}

var once sync.Once         // nolint
var mutex = &sync.Mutex{}  // nolint
var singleton IMigrateTool // nolint

// GetMigrateTool get singleton instance for migrate tool
func GetMigrateTool() IMigrateTool { // nolint
	once.Do(func() {
		singleton = &migrateTool{}
	})
	return singleton
}

// Migrate execute migration in serialize.
func (mt *migrateTool) Migrate(source string, connStr string) {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println("Migrating....")
	// fmt.Printf("Source=%+v Connection=%+v\n", source, connStr)

	mg, err := migrate.New(source, connStr)
	if err != nil {
		fmt.Printf("create new migration fail with err: %v", err)
		panic(err)
	}
	defer mg.Close()

	version, dirty, err := mg.Version()
	if err != nil && err != migrate.ErrNilVersion {
		panic(err)
	}

	if dirty {
		mg.Force(int(version) - 1) // nolint
	}

	err = mg.Up()

	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	fmt.Println("Migration done...")
}

// CreateDBAndMigrate create test store DB and operator DB to execute unit test.
func (mt *migrateTool) CreateDBAndMigrate(cfg *postgres_wrapper.PostgresConfig, migrationFile string) *gorm.DB {
	var db *gorm.DB
	// Wait to create store DB first
	boff := backoff.NewExponentialBackOff()

	// Wait to create operator DB
	err := backoff.Retry(func() error {
		var errNested error
		db, errNested = postgres_wrapper.InitPostgres(cfg)
		if errNested != nil {
			fmt.Printf("Connect postgres error %s \n", errNested.Error())
		} else {
			fmt.Println("Connect postgres successful.")
		}
		return errNested
	}, boff)
	if err != nil {
		panic(err)
	}

	mt.Migrate(migrationFile, cfg.MigrationConnURL)
	return db
}
