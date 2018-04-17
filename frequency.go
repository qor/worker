package worker

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/qor/admin"
	"github.com/qor/qor/resource"
)

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

// Scan scan frequency value
func (frequency *Frequency) Scan(value interface{}) error {
	switch data := value.(type) {
	case []byte:
		return json.Unmarshal(data, frequency)
	case string:
		return frequency.Scan([]byte(data))
	case []string:
		for _, str := range data {
			if err := frequency.Scan([]byte(str)); err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported data")
	}
	return nil
}

// Value get value of frequency
func (frequency Frequency) Value() (driver.Value, error) {
	return json.Marshal(frequency)
}

// GetFrequency get frequency
func (frequency Frequency) GetFrequency() *Frequency {
	return &frequency
}

// ConfigureQorMeta configure qor meta
func (frequency Frequency) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*admin.Meta); ok {
		meta.Type = "frequency"
	}
}
