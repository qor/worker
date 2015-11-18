package worker

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
)

func New(config Config) *Worker {
	if config.Job == nil {
		config.Job = &QorJob{}
	}

	// Auto Migration
	config.DB.AutoMigrate(config.Job)

	return &Worker{Config: &config}
}

type Config struct {
	DB    *gorm.DB
	Queue Queue
	Job   QorJobInterface
}

type Worker struct {
	*Config
	JobResource *admin.Resource
	Jobs        []*Job
}

func (worker *Worker) ConfigureQorResource(res *admin.Resource) {
	worker.JobResource = res.GetAdmin().NewResource(worker.Config.Job)

	router := res.GetAdmin().GetRouter()
	controller := workerController{Worker: worker}

	router.Get("/"+res.ToParam(), controller.index)
	router.Get("/"+res.ToParam()+"/.*$", controller.show)
	router.Post("/"+res.ToParam()+"/.*/run$", controller.runJob)
	router.Post("/"+res.ToParam()+"/.*/kill$", controller.killJob)
}

func (worker *Worker) SetQueue(queue Queue) {
	worker.Queue = queue
}

func (worker *Worker) RegisterJob(job Job) error {
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

func (worker *Worker) GetJob(jobID uint) (QorJobInterface, error) {
	var qorJob QorJobInterface

	if err := worker.DB.First(&qorJob, jobID).Error; err == nil {
		for _, job := range worker.Jobs {
			if job.Name == qorJob.GetJobName() {
				// qorJob.Job = job
				return qorJob, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find job: %v", jobID)
}

func (worker *Worker) AddJob(QorJobInterface) error {
	return nil
}

func (worker *Worker) RunJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetJob().Run(qorJob.GetArgument())
	} else {
		return err
	}
}

func (worker *Worker) KillJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetJob().GetQueue().Kill(qorJob)
	} else {
		return err
	}
}

func (worker *Worker) DeleteJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetJob().GetQueue().Delete(qorJob)
	} else {
		return err
	}
}
