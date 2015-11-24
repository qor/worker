package worker

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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

	SetProgress(uint) error
	SetProgressText(string) error
	AddLog(string) error
	AddTableRow([]TableCell) error

	GetArgument() interface{}
	admin.SerializeArgumentInterface
}

type ErrorTable struct {
	TableCells [][]TableCell
}

func (errorTable *ErrorTable) Scan(data interface{}) error {
	switch values := data.(type) {
	case []byte:
		return json.Unmarshal(values, errorTable)
	case string:
		return errorTable.Scan([]byte(values))
	default:
		return errors.New("unsupported data type for Qor Job error table")
	}
}

func (errorTable ErrorTable) Value() (driver.Value, error) {
	result, err := json.Marshal(errorTable)
	return string(result), err
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
	Log          string     `sql:"size:65532"`
	ErrorTable   ErrorTable `sql:"size:65532"`
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

func (job *QorJob) GetArgument() interface{} {
	return job.GetSerializeArgument(job)
}

func (job *QorJob) GetSerializeArgumentResource() *admin.Resource {
	if j := job.GetJob(); j != nil {
		return j.Resource
	}
	return nil
}

func (job *QorJob) SetProgress(progress uint) error {
	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.Progress = progress
	return worker.JobResource.CallSave(job, context)
}

func (job *QorJob) SetProgressText(str string) error {
	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.ProgressText = str
	return worker.JobResource.CallSave(job, context)
}

func (job *QorJob) AddLog(log string) error {
	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.Log += "\n" + log
	return worker.JobResource.CallSave(job, context)
}

func (job *QorJob) AddTableRow(cells []TableCell) error {
	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.ErrorTable.TableCells = append(job.ErrorTable.TableCells, cells)
	return worker.JobResource.CallSave(job, context)
}
