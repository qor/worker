package worker

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/audited"
	"github.com/qor/serializable_meta"
)

type QorJobInterface interface {
	GetJobID() string
	GetJobName() string
	GetStatus() string
	SetStatus(string) error
	GetJob() *Job
	SetJob(*Job)

	GetProgress() uint
	SetProgress(uint) error
	GetProgressText() string
	SetProgressText(string) error
	GetLogs() []string
	AddLog(string) error
	GetErrorTable() ErrorTable
	AddTableRow(...TableCell) error

	GetArgument() interface{}
	serializable_meta.SerializableMetaInterface
}

type ErrorTable struct {
	Name       string `json:"-"` // only used for generate string column in database
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
	Value string
	Error string
}

type QorJob struct {
	gorm.Model
	Status       string `sql:"default:'new'"`
	Progress     uint
	ProgressText string
	Log          string     `sql:"size:65532"`
	ErrorTable   ErrorTable `sql:"size:65532"`

	mutex sync.Mutex `sql:"-"`
	Job   *Job       `sql:"-"`
	audited.AuditedModel
	serializable_meta.SerializableMeta
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
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.Status = status
	if status == JobStatusDone {
		job.Progress = 100
	}
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
	return job.GetSerializableArgument(job)
}

func (job *QorJob) GetSerializableArgumentResource() *admin.Resource {
	if j := job.GetJob(); j != nil {
		return j.Resource
	}
	return nil
}

func (job *QorJob) GetProgress() uint {
	return job.Progress
}

func (job *QorJob) SetProgress(progress uint) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	if progress > 100 {
		progress = 100
	}
	job.Progress = progress
	return worker.JobResource.CallSave(job, context)
}

func (job *QorJob) GetProgressText() string {
	return job.ProgressText
}

func (job *QorJob) SetProgressText(str string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.ProgressText = str
	return worker.JobResource.CallSave(job, context)
}

func (job *QorJob) GetLogs() []string {
	return strings.Split(job.Log, "\n")
}

func (job *QorJob) AddLog(log string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	fmt.Println(log)
	job.Log += "\n" + log
	return worker.JobResource.CallSave(job, context)
}

func (job *QorJob) GetErrorTable() ErrorTable {
	return job.ErrorTable
}

func (job *QorJob) AddTableRow(cells ...TableCell) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.ErrorTable.TableCells = append(job.ErrorTable.TableCells, cells)
	return worker.JobResource.CallSave(job, context)
}
