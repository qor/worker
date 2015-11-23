package worker

import (
	"os"
	"os/exec"
)

type Cron struct {
}

func NewCronQueue() *Cron {
	return &Cron{}
}

func (Cron) Add(job QorJobInterface) error {
	binaryFile := os.Args[0]
	cmd := exec.Command(binaryFile, "--qor-job", job.GetJobID())
	return cmd.Start()
}

func (Cron) Kill(job QorJobInterface) error {
	return nil
}

func (Cron) Delete(job QorJobInterface) error {
	return nil
}
