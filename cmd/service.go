package main

import (
	"context"
	"errors"
	"github.com/ekimeel/sabal-pb/pb"
	"github.com/ekimeel/sabal-plugin/pkg/plugin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
	"time"
)

var (
	singletonService *service
	onceService      sync.Once
)

type service struct {
	dao *dao
}

func getService() *service {

	onceService.Do(func() {
		singletonService = &service{}
		singletonService.dao = getDao()
	})

	return singletonService
}

func (s *service) Run(offset *plugin.Offset) error {
	status = plugin.Running

	log.Infof("running: %s with offset of: %v", Name(), offset.Value)

	req := &pb.ListRequest{Limit: 10000, Offset: 0}

	res, err := pointServiceClient.GetAll(context.Background(), req)

	if err != nil {
		log.Warn("failed to get points")
		return errors.New("no points")
	}
	log.Infof("found %v points", len(res.Points))

	for _, point := range res.Points {
		log.Tracef("updating point: %v", point.Id)
		cur, err := s.dao.selectByPointId(point.Id)
		log.Tracef("found point with id: %v", point.Id)

		if cur == nil {
			log.Tracef("no existing tqp entry, creating new")
			cur = &TimeQuality{
				PointId: point.Id,
				Start:   time.Unix(0, 0),
				End:     time.Unix(0, 0),
			}
			_, err := s.dao.insert(cur)
			if err != nil {
				log.Errorf("failed to create new dayOfWeek entry: %s", err)
				return errors.New("failed to create entry")
			}
		}

		r := &pb.MetricRequest{
			PointId: point.Id,
			From:    timestamppb.New(cur.End),
			To:      timestamppb.Now(),
		}

		data, err := metricServiceClient.Select(context.Background(), r)
		if err != nil {
			log.Errorf("failed to read data for point id:%d, from:%s, to:%s err: %s", point.Id, r.From, r.To, err)
			continue
		}

		doQualityCalc(data.Metrics, cur)
		_, err = s.dao.update(cur)
		if err != nil {
			log.Errorf("failed to update: %s", err)
		}
	}

	return nil
}
