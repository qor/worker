package worker

import "fmt"

func New() *Worker {
	return &Worker{}
}

type Worker struct {
	Jobs []*Job
}

func (worker *Worker) AddJob(job Job) error {
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

func (worker *Worker) RunJob(name string) error {
	for _, job := range worker.Jobs {
		if job.Name == name {
			return nil
		}
	}
	return fmt.Errorf("failed to find job: %v", name)
}
