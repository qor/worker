package worker

import "time"

type Scheduler interface {
	GetScheduleTime() *time.Time
}

type Schedule struct {
	ScheduleTime *time.Time
}

func (schedule Schedule) GetScheduleTime() *time.Time {
	return schedule.ScheduleTime
}
