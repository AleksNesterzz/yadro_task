package models

import "time"

type Result struct {
	Started        bool
	Finished       bool
	SetTime        time.Time
	StartTime      time.Time
	StartLapTime   time.Time
	StartPenalty   time.Time
	AllPenaltyTime time.Time
	TotalTime      time.Time
	LTAS           []LapTimeAvgSpeed
	PenaltyLapTime LapTimeAvgSpeed
	HitNumber      int
	ShotNumber     int
}

type LapTimeAvgSpeed struct {
	LapTime  time.Time
	AvgSpeed float64
}

func (ltas *LapTimeAvgSpeed) IsEmpty() bool {
	zero, _ := time.Parse("15:04:05.999", "00:00:00.000")
	zeroStr := zero.Format("15:04:05.999")
	temp := ltas.LapTime.Format("15:04:05.999")
	if temp == zeroStr {
		return true
	} else {
		return false
	}
}
