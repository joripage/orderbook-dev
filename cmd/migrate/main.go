package main

import (
	"encoding/json"
	"flag"

	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joripage/orderbook-dev/config"
	"github.com/joripage/orderbook-dev/pkg/infra"
	"go.uber.org/zap"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config-file", "", "Specify config file path")
	flag.Parse()

	cfg, err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	configBytes, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		zap.S().Warnf("could not convert config to JSON: %v", err)
	} else {
		zap.S().Debugf("load config %s", string(configBytes))
	}

	mgTool := infra.GetMigrateTool()
	mgTool.Migrate("file://migration/sql", cfg.OmsDB.MigrationConnURL)
}
