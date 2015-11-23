package worker

import "github.com/qor/worker"

type Cron struct {
}

func New() *Cron {
	return &Cron{}
}

func (Cron) Add(job worker.QorJobInterface) error {
	return nil
}

func (Cron) Kill(job worker.QorJobInterface) error {
	return nil
}

func (Cron) Delete(job worker.QorJobInterface) error {
	return nil
}
