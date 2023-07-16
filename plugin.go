package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ekimeel/sabal-plugin/pkg/metric_utils"
	"github.com/ekimeel/sabal-plugin/pkg/plugin"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	tq "sabal-time-quality/pkg"
	"time"
)

const (
	PluginName     = "timeQuality"
	PluginVersion  = "v1.0"
	configFilePath = "./ext/tq.json"
	envSqlDb       = "sql.DB"
)

var (
	logger *log.Logger
)

func Install(env *plugin.Environment) error {

	if val, ok := env.Get("logger"); ok {
		logger, _ = val.(*log.Logger)
	} else {
		logger = log.New()
	}

	logger.Infof("installing plugin: %s@%s", PluginName, PluginVersion)

	logger.Infof("configuring environment")
	if err := setupDatabase(env); err != nil {
		return err
	}

	var config plugin.Config
	configFile, err := os.ReadFile(configFilePath)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return err
	}

	return nil
}

// Process handles the plugin event. It logs the processing time and any errors that occur.
func Process(ctx context.Context, event plugin.Event) {
	if len(event.Metrics) == 0 {
		log.Infof("no metrics to process")
		return
	}

	ref := metric_utils.ConvertToPointerSlice(event.Metrics)

	log.Infof("running: %s", Name())
	start := time.Now()

	tq.GetService().Run(ctx, ref)

	log.Infof("%s, processed [%d] in [%s]", Name(), len(event.Metrics), time.Since(start))
}

func Name() string {
	return fmt.Sprintf("%s@%s", PluginName, PluginVersion)
}

func setupDatabase(env *plugin.Environment) error {
	if val, ok := env.Get(envSqlDb); ok {
		tq.DB = val.(*sql.DB)
		log.Infof("sucessfully found %s", envSqlDb)
	} else {
		return fmt.Errorf("plugin %s requireds a valid %s value", PluginName, envSqlDb)
	}
	return nil
}
