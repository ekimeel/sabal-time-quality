package main

import (
	"database/sql"
	"errors"
	"github.com/ekimeel/sabal-pb/pb"
	"github.com/ekimeel/sabal-plugin/pkg/plugin"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	oncePlugin          sync.Once
	offset              plugin.Offset
	status              plugin.Status
	db                  *sql.DB
	metricServiceClient pb.MetricServiceClient
	pointServiceClient  pb.PointServiceClient
	runChannel          chan string
	ticker              *time.Ticker
	logger              *log.Logger
)

func Run(offset *plugin.Offset) error {
	log.Infof("running timeQualityPlugin: %s", Name())
	status = plugin.Running
	err := getService().Run(offset)
	if err != nil {
		status = plugin.Error
		return errors.New("failed to run timeQualityPlugin")
	}
	status = plugin.Ready
	return nil
}
