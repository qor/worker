package worker

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/audited"
)

type QorJobInterface interface {
	GetJobName() string
	GetStatus() string
	SetStatus(string)
	GetJob() *Job
	SetJob(*Job)
	admin.SerializeArgumentInterface
}

type QorJob struct {
	gorm.Model
	Status string `sql:"default:'new'"`
	audited.AuditedModel
	admin.SerializeArgument
	Job *Job `sql:"-"`
}

func (job *QorJob) GetJobName() string {
	return job.Kind
}

func (job *QorJob) GetStatus() string {
	return job.Status
}

func (job *QorJob) SetStatus(status string) {
	job.Status = status
}

func (job *QorJob) SetJob(j *Job) {
	job.Kind = j.Name
	job.Job = j
}

func (job *QorJob) GetJob() *Job {
	if job.Job != nil {
		return job.Job
	}
	return nil
}

func (job *QorJob) GetSerializeArgumentResource() *admin.Resource {
	if j := job.GetJob(); j != nil {
		return j.Resource
	}
	return nil
}
