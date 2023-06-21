package main

import (
	"database/sql"
	"fmt"
	"github.com/ekimeel/sabal-pb/pb"
	"github.com/ekimeel/sabal-plugin/pkg/plugin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

func Install(env *plugin.Environment) error {

	if status == plugin.Running {
		return fmt.Errorf("plugin %s has running status, cannot be install", pluginName)
	}

	if val, ok := env.Get("logger"); ok {
		logger, _ = val.(*log.Logger)
	} else {
		logger = log.New()
	}

	logger.Infof("installing plugin: %s@%s", pluginName, pluginVersion)

	logger.Infof("configuring environment")
	if val, ok := env.Get(envSqlDb); ok {
		db = val.(*sql.DB)
		log.Infof("sucessfully found %s", envSqlDb)
	} else {
		return fmt.Errorf("plugin %s requireds a valid %s value", pluginName, envSqlDb)
	}

	if val, ok := env.Get(envPointServiceClient); ok {
		pointServiceClient = val.(pb.PointServiceClient)
		log.Infof("sucessfully found %s", envPointServiceClient)
	} else {
		return fmt.Errorf("plugin %s requireds a valid %s value", pluginName, envPointServiceClient)
	}

	if val, ok := env.Get(envMetricServiceClient); ok {
		metricServiceClient = val.(pb.MetricServiceClient)
		log.Infof("sucessfully found %s", envMetricServiceClient)
	} else {
		return fmt.Errorf("plugin %s requireds a valid %s value", pluginName, envMetricServiceClient)
	}

	if val, ok := env.Get(envRunChannel); ok {
		runChannel = val.(chan string)
		log.Infof("sucessfully found %s", envRunChannel)
	} else {
		return fmt.Errorf("plugin %s requireds a valid %s value", pluginName, envRunChannel)
	}

	var config Config
	configFile, err := os.ReadFile(configFile)

	if err != nil {
		log.Fatalf("failed to load config file: %s", err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("failed to unmarshal config file: %s", err)
	}

	err = yaml.Unmarshal(configFile, &config.Data)
	if err != nil {
		log.Fatalf("failed to unmarshal config file: %s", err)
	}

	status = plugin.None

	start()

	return nil
}

func Name() string {
	return fmt.Sprintf("%s@%s", pluginName, pluginVersion)
}

func start() {
	log.Infof("starting ticker")
	ticker = time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Infof("requesting update: %s", Name())
				runChannel <- Name()
			}
		}
	}()
}
