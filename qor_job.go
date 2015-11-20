package worker

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/audited"
)

type QorJobInterface interface {
	SetWorker(*Worker)
	GetJobName() string
	SetJobName(string)
	GetStatus() string
	SetStatus(string)
	GetJob() *Job
	admin.SerializeArgumentInterface
}

type QorJob struct {
	gorm.Model
	Status string
	audited.AuditedModel
	admin.SerializeArgument
	Worker *Worker `sql:"-"`
}

func (job *QorJob) GetJobName() string {
	return job.Kind
}

func (job *QorJob) SetJobName(name string) {
	job.Kind = name
}

func (job *QorJob) GetStatus() string {
	return job.Status
}

func (job *QorJob) SetStatus(status string) {
	job.Status = status
}

func (job *QorJob) SetWorker(worker *Worker) {
	job.Worker = worker
}

func (job *QorJob) GetJob() *Job {
	if job.Worker != nil {
		for _, j := range job.Worker.Jobs {
			if j.Name == job.GetJobName() {
				return j
			}
		}
	}
	return nil
}

func (job *QorJob) GetSerializeArgumentResource() *admin.Resource {
	if j := job.GetJob(); j != nil {
		return j.Resource
	}
	return nil
}
