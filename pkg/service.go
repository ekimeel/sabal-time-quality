package tq

import (
	"context"
	"github.com/ekimeel/sabal-pb/pb"
	"github.com/ekimeel/sabal-plugin/pkg/metric_utils"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	singletonService *Service
	onceService      sync.Once
)

type Service struct {
	dao *dao
}

func GetService() *Service {

	onceService.Do(func() {
		singletonService = &Service{}
		singletonService.dao = getDao()
	})

	return singletonService
}

func (s *Service) dowork(pointId uint32, metrics []*pb.Metric) {
	cur, err := s.dao.selectByPointId(pointId)

	if cur == nil {
		log.Tracef("no existing tqp entry, creating new")
		cur = &TimeQuality{
			Id:      0,
			PointId: pointId,
			Start:   time.Unix(0, 0),
			End:     time.Unix(0, 0),
		}
	}

	doQualityCalc(metrics, cur)

	if cur.Id == 0 {
		_, err = s.dao.insert(cur)
	} else {
		_, err = s.dao.update(cur)
	}

	if err != nil {
		log.Errorf("failed to update: %s", err)
	}
}

func (s *Service) Run(ctx context.Context, metrics []*pb.Metric) {

	unitOfWork := metric_utils.GroupMetricsByPointId(metrics)
	var wg sync.WaitGroup

	for pointId, items := range unitOfWork {
		wg.Add(1)
		go func(pointId uint32, items []*pb.Metric) {
			defer wg.Done()
			s.dowork(pointId, items)
		}(pointId, items)
	}

	wg.Wait()
}
