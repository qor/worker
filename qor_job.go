package worker

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/audited"
)

type QorJobInterface interface {
	GetJobID() string
	GetJobName() string
	GetStatus() string
	SetStatus(string) error
	GetJob() *Job
	SetJob(*Job)

	SetProgress(uint)
	SetProgressText(string)
	AddLog(string)
	AddErrorRow([]TableCell)

	admin.SerializeArgumentInterface
}

type TableCell struct {
	Value interface{}
	Error error
}

type QorJob struct {
	gorm.Model
	Status       string `sql:"default:'new'"`
	Progress     uint
	ProgressText string
	Log          string `sql:"size:65532"`
	audited.AuditedModel
	admin.SerializeArgument
	Job *Job `sql:"-"`
}

func (job QorJob) GetJobID() string {
	return fmt.Sprint(job.ID)
}

func (job *QorJob) GetJobName() string {
	return job.Kind
}

func (job *QorJob) GetStatus() string {
	return job.Status
}

func (job *QorJob) SetStatus(status string) error {
	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.Status = status
	return worker.JobResource.CallSave(job, context)
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

func (job *QorJob) SetProgress(uint) {
}

func (job *QorJob) SetProgressText(string) {
}

func (job *QorJob) AddLog(string) {
}

func (job *QorJob) AddErrorRow([]TableCell) {
}
