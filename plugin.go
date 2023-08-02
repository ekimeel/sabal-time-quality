package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ekimeel/sabal-pb/pb"
	"github.com/ekimeel/sabal-plugin/pkg/metric_utils"
	"github.com/ekimeel/sabal-plugin/pkg/plugin"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	tq "sabal-time-quality/pkg"
	"time"
)

const (
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

	logger.WithField("plugin", Name()).Info("installing plugin")
	logger.WithField("plugin", Name()).Info("configuring environment")
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

	if metrics, ok := event.Data.([]pb.Metric); ok {

		if len(metrics) == 0 {
			logger.WithField("plugin", Name()).Info("no metrics to process")
			return
		}

		ref := metric_utils.ConvertToPointerSlice(metrics)

		logger.WithField("plugin", Name()).Infof("no metrics to process")
		log.Infof("running: %s", Name())
		start := time.Now()

		tq.GetService().Run(ctx, ref)
		logger.WithField("plugin", Name()).
			Infof("%s, processed [%d] in [%s]", Name(), len(metrics), time.Since(start))
	} else {
		// Handle case where Data does not hold a []pb.Metric
		fmt.Println("Event Data is not of type []pb.Metric")
	}
}

func Name() string {
	return fmt.Sprintf("%s@%s", tq.PluginName, tq.PluginVersion)
}

func setupDatabase(env *plugin.Environment) error {
	if val, ok := env.Get(envSqlDb); ok {
		tq.DB = val.(*sql.DB)
		log.Infof("sucessfully found %s", envSqlDb)
	} else {
		return fmt.Errorf("plugin %s requireds a valid %s value", tq.PluginName, envSqlDb)
	}
	return nil
}
