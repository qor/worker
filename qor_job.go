package worker

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/audited"
)

type QorJobInterface interface {
	GetJobName() string
	SetJobName(string)
	GetStatus() string
	SetStatus(string)
	SetJob(*Job)
	GetJob() *Job
	admin.SerializeArgumentInterface
}

type QorJob struct {
	gorm.Model
	Status string
	audited.AuditedModel
	admin.SerializeArgument
	Job *Job `sql:"-"`
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

func (job *QorJob) SetJob(j *Job) {
	job.Job = j
}

func (job *QorJob) GetJob() *Job {
	return job.Job
}

func (job *QorJob) GetSerializeArgumentResource() *admin.Resource {
	if job.Job != nil {
		return job.Job.Worker.JobResource
	}
	return nil
}
