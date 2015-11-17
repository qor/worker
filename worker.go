package worker

func New() *Worker {
	return &Worker{}
}

type Worker struct {
}

func (worker *Worker) AddJob(job Job) {
}
