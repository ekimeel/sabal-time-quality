package tq

import (
	"database/sql"
	"encoding/json"
	"github.com/ekimeel/sabal-pb/pb"
	"math"
	"sort"
	"time"
)

type TimeQuality struct {
	Id                         uint32          `json:"id"`
	PointId                    uint32          `json:"point-id"`
	LastUpdated                time.Time       `json:"last-updated"`
	Start                      time.Time       `json:"start"`
	End                        time.Time       `json:"end"`
	Count                      int64           `json:"count"`
	MeanWait                   sql.NullFloat64 `json:"mean-wait"`
	MaxWait                    sql.NullInt64   `json:"max-wait"`
	MinWait                    sql.NullInt64   `json:"min-wait"`
	WaitCoefficientOfVariation sql.NullFloat64 `json:"wait-coefficient-of-variation"`
	WaitStandardDeviation      sql.NullFloat64 `json:"wait-standard-deviation"`
	FillFactor                 sql.NullFloat64 `json:"fill-factor"`
	Score                      sql.NullFloat64 `json:"score"`
}

func computeScore(existing *TimeQuality) {
	// compute fill factor
	span := float64(existing.End.Unix() - existing.Start.Unix())
	mean := existing.MeanWait.Float64
	count := float64(existing.Count)

	fillFactor := count / (span / mean) * 100
	cv := 100 - existing.WaitCoefficientOfVariation.Float64

	score := (fillFactor + cv) / 2

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	existing.Score.Scan(score)

}

func doQualityCalc(data []*pb.Metric, existing *TimeQuality) {

	if len(data) < 4 || existing == nil {
		return
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Timestamp.Seconds < data[j].Timestamp.Seconds
	})

	intervals := make([]int64, len(data)-1)

	// calculate time intervals and mean
	var sum int64
	for i := 1; i < len(data); i++ {
		intervals[i-1] = data[i].Timestamp.Seconds - data[i-1].Timestamp.Seconds

		sum += intervals[i-1]
		if (intervals[i-1] > existing.MaxWait.Int64) || existing.MaxWait.Valid == false {
			existing.MaxWait.Scan(intervals[i-1])
		}
		if (intervals[i-1] < existing.MinWait.Int64) || existing.MinWait.Valid == false {
			existing.MinWait.Scan(intervals[i-1])
		}
	}

	mean := float64(sum) / float64(len(intervals))

	// compute standard deviation and coefficient of variation
	var sd, cv float64
	for _, interval := range intervals {
		diff := float64(interval) - mean
		sd += diff * diff
	}
	sd = math.Sqrt(sd / float64(len(intervals)-1))
	cv = sd / mean * 100.0

	// compute fill factor
	span := data[len(data)-1].Timestamp.Seconds - data[0].Timestamp.Seconds
	fillFactor := float64(len(data)) / (float64(span) / mean) * 100

	if existing.Count > 0 {
		//weighted average
		curWeight := float64(len(data)) / float64(existing.Count)
		totalWeight := 1.0 - curWeight

		existing.MeanWait.Scan(weightedAverage(existing.MeanWait.Float64, totalWeight, mean, curWeight))
		existing.WaitStandardDeviation.Scan(weightedAverage(existing.WaitStandardDeviation.Float64, totalWeight, sd, curWeight))
		existing.WaitCoefficientOfVariation.Scan(weightedAverage(existing.WaitCoefficientOfVariation.Float64, totalWeight, cv, curWeight))
		existing.FillFactor.Scan(weightedAverage(existing.FillFactor.Float64, totalWeight, fillFactor, curWeight))

		if existing.Start.Unix() > data[0].Timestamp.Seconds {
			existing.Start = data[0].Timestamp.AsTime()
		}

		if existing.End.Unix() < data[len(data)-1].Timestamp.Seconds {
			existing.End = data[len(data)-1].Timestamp.AsTime()
		}

	} else {
		existing.MeanWait.Scan(mean)
		existing.WaitStandardDeviation.Scan(sd)
		existing.WaitCoefficientOfVariation.Scan(cv)
		existing.Start = data[0].Timestamp.AsTime()
		existing.End = data[len(data)-1].Timestamp.AsTime()
		existing.FillFactor.Scan(fillFactor)
	}

	existing.Count += int64(len(data))
	computeScore(existing)

}

func weightedAverage(value1, weight1, value2, weight2 float64) float64 {
	return (value1*weight1 + value2*weight2) / (weight1 + weight2)
}

func (d *TimeQuality) String() string {
	s, err := json.Marshal(d)
	if err != nil {
		return "error"
	}
	return string(s)
}
