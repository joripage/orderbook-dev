package postgres_wrapper

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	_ "github.com/lib/pq" // nolint
	"go.uber.org/zap"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type PostgresConfig struct {
	DriverName                 string `yaml:"driver_name"`
	DataSource                 string `yaml:"data_source"`
	Username                   string `yaml:"username"`
	Password                   string `yaml:"password"`
	DatabaseName               string `yaml:"database_name"`
	MaxOpenConns               int    `yaml:"max_open_conns"`
	MaxIdleConns               int    `yaml:"max_idle_conns"`
	ConnMaxLifeTimeMiliseconds int64  `yaml:"conn_max_life_time_ms"`
	MigrationConnURL           string `yaml:"migration_conn_url"`
	IsDevMode                  bool   `yaml:"is_dev_mode"`
	UseCache                   bool   `yaml:"use_cache"`
	//EnableTracing              bool     `yaml:"enable_tracing"`
	SlaveSources []string        `yaml:"slave_sources"`
	LogLevel     logger.LogLevel `yaml:"log_level"`
	Location     string          `yaml:"location"`
}

// InitPostgres set up postgres
func InitPostgres(cfg *PostgresConfig) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,  // Slow SQL threshold
			LogLevel:      cfg.LogLevel, // Log level
			Colorful:      true,         // Disable color
		},
	)

	db, err := gorm.Open(pg.Open(cfg.DataSource), &gorm.Config{
		Logger: newLogger,
		NowFunc: func() time.Time {
			loc, _ := time.LoadLocation(cfg.Location)
			return time.Now().In(loc)
		},
	})
	if err != nil {
		zap.S().Debugf("open portgres fail: %+v", err)
		return nil, err
	}

	var repl []gorm.Dialector
	for _, s := range cfg.SlaveSources {
		repl = append(repl, pg.Open(s))
	}

	if len(repl) > 0 {
		zap.S().Debugf("register replicas portgres")
		err := db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: repl,
			Policy:   dbresolver.RandomPolicy{},
		}))
		if err != nil {
			zap.S().Debugf("init portgres replicas fail: %+v", err)
			return nil, err
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Debugf("get DB instance failed %v", err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTimeMiliseconds) * time.Millisecond)

	return db, nil
}

// InitPostgresWithBackoff init postgre db use backoff
func InitPostgresWithBackoff(cfg *PostgresConfig) *gorm.DB {
	var db *gorm.DB
	boff := backoff.NewExponentialBackOff()
	err := backoff.Retry(func() error {
		var err error
		db, err = InitPostgres(cfg)
		if err != nil {
			fmt.Printf("Connect postgres error %s \n", err.Error())
		}
		return err
	}, boff)
	if err != nil {
		panic(err)
	}

	return db
}
