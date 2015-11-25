package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type cronJob struct {
	JobID   string
	Pid     int
	Command string
	Delete  bool `json:"-"`
}

func (job cronJob) ToString() string {
	marshal, _ := json.Marshal(job)
	return fmt.Sprintf("## BEGIN QOR JOB %v # %v\n%v\n## END QOR JOB\n", job.JobID, string(marshal), job.Command)
}

type Cron struct {
	Jobs  []*cronJob
	mutex sync.Mutex `sql:"-"`
}

func NewCronQueue() *Cron {
	return &Cron{}
}

func (cron *Cron) ParseJobs() []*cronJob {
	cron.mutex.Lock()

	cron.Jobs = []*cronJob{}
	if out, err := exec.Command("crontab", "-l").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "## BEGIN QOR JOB") {
				if idx := strings.Index(line, "{"); idx > 1 {
					var job cronJob
					if json.Unmarshal([]byte(line[idx-1:]), &job) == nil {
						cron.Jobs = append(cron.Jobs, &job)
					}
				}
			}
		}
	}
	return cron.Jobs
}

func (cron *Cron) WriteCronJob() error {
	defer cron.mutex.Unlock()

	cmd := exec.Command("crontab", "-")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	for _, job := range cron.Jobs {
		if !job.Delete {
			stdin.Write([]byte(job.ToString()))
		}
	}
	stdin.Close()
	return cmd.Run()
}

func (cron *Cron) Add(job QorJobInterface) error {
	cron.ParseJobs()
	defer cron.WriteCronJob()

	binaryFile := os.Args[0]
	cmd := exec.Command(binaryFile, "--qor-job", job.GetJobID())
	if err := cmd.Start(); err == nil {
		cron.Jobs = append(cron.Jobs, &cronJob{
			JobID:   job.GetJobID(),
			Command: "", // FIXME cronjob scheduler
			Pid:     cmd.Process.Pid,
		})
		cmd.Process.Release()
		return nil
	} else {
		return err
	}
}

func (cron *Cron) Run(qorJob QorJobInterface) error {
	cron.ParseJobs()
	defer cron.WriteCronJob()

	for _, cronJob := range cron.Jobs {
		if cronJob.JobID == qorJob.GetJobID() {
			if job := qorJob.GetJob(); job.Handler != nil {
				if err := job.Handler(qorJob.GetSerializeArgument(qorJob), qorJob); err == nil {
					cronJob.Delete = true
					return nil
				} else {
					return err
				}
			} else {
				return errors.New("no handler found for job " + job.Name)
			}
		}
	}

	return nil
}

func (cron *Cron) Kill(job QorJobInterface) error {
	cron.ParseJobs()
	defer cron.WriteCronJob()

	for _, cronJob := range cron.Jobs {
		if cronJob.JobID == job.GetJobID() {
			if process, err := os.FindProcess(cronJob.Pid); err == nil {
				if err := process.Kill(); err == nil {
					cronJob.Delete = true
					return nil
				} else {
					return err
				}
			} else {
				return err
			}
		}
	}
	return errors.New("failed to find job")
}

func (cron *Cron) Remove(job QorJobInterface) error {
	cron.ParseJobs()
	defer cron.WriteCronJob()

	for _, cronJob := range cron.Jobs {
		if cronJob.JobID == job.GetJobID() {
			if cronJob.Pid == 0 {
				cronJob.Delete = true
			} else {
				return errors.New("failed to remove current job as it is running")
			}
		}
	}
	return errors.New("failed to find job")
}
