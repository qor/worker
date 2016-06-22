package worker

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/audited"
	"github.com/qor/serializable_meta"
)

// QorJobInterface is a interface, defined methods that needs for a qor job
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
	GetResultsTable() ResultsTable
	AddResultsRow(...TableCell) error

	GetArgument() interface{}
	serializable_meta.SerializableMetaInterface
}

// ResultsTable is a struct, including importing/exporting results
type ResultsTable struct {
	Name       string `json:"-"` // only used for generate string column in database
	TableCells [][]TableCell
}

// Scan used to scan value from database into itself
func (resultsTable *ResultsTable) Scan(data interface{}) error {
	switch values := data.(type) {
	case []byte:
		return json.Unmarshal(values, resultsTable)
	case string:
		return resultsTable.Scan([]byte(values))
	default:
		return errors.New("unsupported data type for Qor Job error table")
	}
}

// Value used to read value from itself and save it into databae
func (resultsTable ResultsTable) Value() (driver.Value, error) {
	result, err := json.Marshal(resultsTable)
	return string(result), err
}

// TableCell including Value, Error for a data cell
type TableCell struct {
	Value string
	Error string
}

// QorJob predefined qor job struct, which will be used for Worker, if it doesn't include a job resource
type QorJob struct {
	gorm.Model
	Status       string `sql:"default:'new'"`
	Progress     uint
	ProgressText string
	Log          string       `sql:"size:65532"`
	ResultsTable ResultsTable `sql:"size:65532"`

	mutex sync.Mutex `sql:"-"`
	Job   *Job       `sql:"-"`
	audited.AuditedModel
	serializable_meta.SerializableMeta
}

// GetJobID get job's ID from a qor job
func (job *QorJob) GetJobID() string {
	return fmt.Sprint(job.ID)
}

// GetJobName get job's name from a qor job
func (job *QorJob) GetJobName() string {
	return job.Kind
}

// GetStatus get job's status from a qor job
func (job *QorJob) GetStatus() string {
	return job.Status
}

// SetStatus set job's status to a qor job instance
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

// SetJob set `Job` for a qor job instance
func (job *QorJob) SetJob(j *Job) {
	job.Kind = j.Name
	job.Job = j
}

// GetJob get predefined job for a qor job instance
func (job *QorJob) GetJob() *Job {
	if job.Job != nil {
		return job.Job
	}
	return nil
}

// GetArgument get job's argument
func (job *QorJob) GetArgument() interface{} {
	return job.GetSerializableArgument(job)
}

// GetSerializableArgumentResource get job's argument's resource
func (job *QorJob) GetSerializableArgumentResource() *admin.Resource {
	if j := job.GetJob(); j != nil {
		return j.Resource
	}
	return nil
}

// GetProgress get qor job's progress
func (job *QorJob) GetProgress() uint {
	return job.Progress
}

// SetProgress set qor job's progress
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

// GetProgressText get qor job's progress text
func (job *QorJob) GetProgressText() string {
	return job.ProgressText
}

// SetProgressText set qor job's progress text
func (job *QorJob) SetProgressText(str string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.ProgressText = str
	return worker.JobResource.CallSave(job, context)
}

// GetLogs get qor job's logs
func (job *QorJob) GetLogs() []string {
	return strings.Split(job.Log, "\n")
}

// AddLog add a log to qor job
func (job *QorJob) AddLog(log string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	fmt.Println(log)
	job.Log += "\n" + log
	return worker.JobResource.CallSave(job, context)
}

// GetResultsTable get the job's process logs
func (job *QorJob) GetResultsTable() ResultsTable {
	return job.ResultsTable
}

// AddResultsRow add a row of process results to a job
func (job *QorJob) AddResultsRow(cells ...TableCell) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	worker := job.GetJob().Worker
	context := worker.Admin.NewContext(nil, nil).Context
	job.ResultsTable.TableCells = append(job.ResultsTable.TableCells, cells)
	return worker.JobResource.CallSave(job, context)
}
