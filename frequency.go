package worker

import "time"

// Frequencier frequencier interface
type Frequencier interface {
	GetFrequency() *Frequency
}

// Frequency frequency struct
type Frequency struct {
	ScheduledStartAt *time.Time
	ScheduledEndAt   *time.Time
	Interval         time.Duration
}

// GetFrequency get frequency
func (frequency Frequency) GetFrequency() *Frequency {
	return &frequency
}
