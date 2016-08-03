package worker

import (
	"time"
)

// Scheduler is a interface, for job used to `GetScheduleTime`
type Scheduler interface {
	GetScheduleTime() *time.Time
	GetRepeatTime() *time.Duration
}

// Schedule could be embedded as job argument, then the job will get run as scheduled feature
type Schedule struct {
	ScheduleTime *time.Time
	RepeatTime *string `meta:"label:22"`
}

// GetScheduleTime get scheduled time
func (schedule Schedule) GetScheduleTime() *time.Time {
	if scheduleTime := schedule.ScheduleTime; scheduleTime != nil {
		if scheduleTime.After(time.Now().Add(time.Minute)) {
			return scheduleTime
		}
	}
	return nil
}

// GetScheduleTime get scheduled time
func (schedule Schedule) GetRepeatTime() *time.Duration {
	if repeatTime := schedule.RepeatTime; repeatTime != nil {
		repeatTimeDuration, _ := time.ParseDuration(*repeatTime)
		if repeatTimeDuration > time.Second {
			return &repeatTimeDuration
		}
	}
	return nil
}
