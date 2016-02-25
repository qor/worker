package worker

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

const (
	// JobStatusNew job status new
	JobStatusNew = "new"
	// JobStatusRunning job status running
	JobStatusRunning = "running"
	// JobStatusDone job status done
	JobStatusDone = "done"
	// JobStatusException job status exception
	JobStatusException = "exception"
	// JobStatusKilled job status killed
	JobStatusKilled = "killed"
)

// New create Worker with Config
func New(config ...Config) *Worker {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Job == nil {
		cfg.Job = &QorJob{}
	}

	if cfg.Queue == nil {
		cfg.Queue = NewCronQueue()
	}

	return &Worker{Config: &cfg}
}

// Config worker config
type Config struct {
	Queue Queue
	Job   QorJobInterface
	Admin *admin.Admin
}

// Worker worker definition
type Worker struct {
	*Config
	JobResource *admin.Resource
	Jobs        []*Job
}

// ConfigureQorResourceBeforeInitialize a method used to config Worker for qor admin
func (worker *Worker) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		admin.RegisterViewPath("github.com/qor/worker/views")
		res.UseTheme("worker")

		worker.Admin = res.GetAdmin()
		worker.JobResource = worker.Admin.NewResource(worker.Config.Job)
		worker.JobResource.UseTheme("worker")
		worker.JobResource.Meta(&admin.Meta{Name: "Name", Valuer: func(record interface{}, context *qor.Context) interface{} {
			return record.(QorJobInterface).GetJobName()
		}})
		worker.JobResource.IndexAttrs("ID", "Name", "Status", "CreatedAt")
		worker.JobResource.Name = res.Name

		for _, status := range []string{JobStatusNew, JobStatusRunning, JobStatusDone, JobStatusException} {
			var status = status
			worker.JobResource.Scope(&admin.Scope{Name: status, Handle: func(db *gorm.DB, ctx *qor.Context) *gorm.DB {
				return db.Where("status = ?", status)
			}})
		}

		// default scope
		worker.JobResource.Scope(&admin.Scope{
			Handle: func(db *gorm.DB, ctx *qor.Context) *gorm.DB {
				if jobName := ctx.Request.URL.Query().Get("job"); jobName != "" {
					return db.Where("kind = ?", jobName)
				}

				if groupName := ctx.Request.URL.Query().Get("group"); groupName != "" {
					var jobNames []string
					for _, job := range worker.Jobs {
						if groupName == job.Group {
							jobNames = append(jobNames, job.Name)
						}
					}
					if len(jobNames) > 0 {
						return db.Where("kind IN (?)", jobNames)
					}
					return db.Where("kind IS NULL")
				}

				return db
			},
			Default: true,
		})

		// Auto Migration
		worker.Admin.Config.DB.AutoMigrate(worker.Config.Job)

		// Configure jobs
		for _, job := range worker.Jobs {
			if job.Resource == nil {
				job.Resource = worker.Admin.NewResource(worker.JobResource.Value)
			}
		}
	}
}

// ConfigureQorResource a method used to config Worker for qor admin
func (worker *Worker) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		// Parse job
		var qorJobID = flag.String("qor-job", "", "Qor Job ID")
		var runAnother = flag.Bool("run-another", false, "Run another qor job")
		flag.Parse()
		if *qorJobID != "" {
			if *runAnother == true {
				if newJob := worker.saveAnotherJob(*qorJobID); newJob != nil {
					newJobID := newJob.GetJobID()
					qorJobID = &newJobID
				} else {
					fmt.Println("failed to clone job " + *qorJobID)
					os.Exit(1)
				}
			}

			if err := worker.RunJob(*qorJobID); err == nil {
				os.Exit(0)
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		// register view funcmaps
		worker.Admin.RegisterFuncMap("get_grouped_jobs", func(context *admin.Context) map[string][]*Job {
			var groupedJobs = map[string][]*Job{}
			var groupName = context.Request.URL.Query().Get("group")
			var jobName = context.Request.URL.Query().Get("job")
			for _, job := range worker.Jobs {
				if (groupName == "" || groupName == job.Group) && (jobName == "" || jobName == job.Name) {
					groupedJobs[job.Group] = append(groupedJobs[job.Group], job)
				}
			}
			return groupedJobs
		})

		// configure routes
		router := worker.Admin.GetRouter()
		controller := workerController{Worker: worker}
		jobParamIDName := worker.JobResource.ParamIDName()
		router.Get(res.ToParam(), controller.Index)
		router.Get(res.ToParam()+"/new", controller.New)
		router.Get(fmt.Sprintf("%v/%v", res.ToParam(), jobParamIDName), controller.Show)
		router.Get(fmt.Sprintf("%v/%v/edit", res.ToParam(), jobParamIDName), controller.Show)
		router.Post(fmt.Sprintf("%v/%v/run", res.ToParam(), jobParamIDName), controller.RunJob)
		router.Post(res.ToParam(), controller.AddJob)
		router.Delete(fmt.Sprintf("%v/%v", res.ToParam(), jobParamIDName), controller.KillJob)
	}
}

// SetQueue set worker's queue
func (worker *Worker) SetQueue(queue Queue) {
	worker.Queue = queue
}

// RegisterJob register a job into Worker
func (worker *Worker) RegisterJob(job *Job) error {
	job.Worker = worker
	worker.Jobs = append(worker.Jobs, job)
	return nil
}

// GetJob get job with id
func (worker *Worker) GetJob(jobID string) (QorJobInterface, error) {
	qorJob := worker.JobResource.NewStruct().(QorJobInterface)

	context := worker.Admin.NewContext(nil, nil)
	context.ResourceID = jobID
	context.Resource = worker.JobResource

	if err := worker.JobResource.FindOneHandler(qorJob, nil, context.Context); err == nil {
		for _, job := range worker.Jobs {
			if job.Name == qorJob.GetJobName() {
				qorJob.SetJob(job)
				return qorJob, nil
			}
		}
		return nil, fmt.Errorf("failed to load job: %v, unregistered job type: %v", jobID, qorJob.GetJobName())
	}
	return nil, fmt.Errorf("failed to find job: %v", jobID)
}

// AddJob add job to worker
func (worker *Worker) AddJob(qorJob QorJobInterface) error {
	return worker.Queue.Add(qorJob)
}

// RunJob run job with job id
func (worker *Worker) RunJob(jobID string) error {
	qorJob, err := worker.GetJob(jobID)

	if err == nil {
		defer func() {
			if r := recover(); r != nil {
				qorJob.SetProgressText(fmt.Sprint(r))
				qorJob.SetStatus(JobStatusException)
			}
		}()

		if qorJob.GetStatus() != JobStatusNew {
			return errors.New("invalid job status, current status: " + qorJob.GetStatus())
		}

		if err = qorJob.SetStatus(JobStatusRunning); err == nil {
			if err = qorJob.GetJob().GetQueue().Run(qorJob); err == nil {
				return qorJob.SetStatus(JobStatusDone)
			}

			qorJob.SetProgressText(err.Error())
			qorJob.SetStatus(JobStatusException)
		}
	}

	return err
}

func (worker *Worker) saveAnotherJob(jobID string) QorJobInterface {
	jobResource := worker.JobResource
	newJob := jobResource.NewStruct().(QorJobInterface)

	job, err := worker.GetJob(jobID)
	if err == nil {
		newJob.SetJob(job.GetJob())
		newJob.SetSerializableArgumentValue(job.GetArgument())
		context := worker.Admin.NewContext(nil, nil)
		if err := jobResource.CallSave(newJob, context.Context); err == nil {
			return newJob
		}
	}
	return nil
}

// KillJob kill job with job id
func (worker *Worker) KillJob(jobID string) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		if qorJob.GetStatus() == JobStatusRunning {
			if err = qorJob.GetJob().GetQueue().Kill(qorJob); err == nil {
				qorJob.SetStatus(JobStatusKilled)
				return nil
			}
			return err
		} else if qorJob.GetStatus() == JobStatusNew {
			return worker.RemoveJob(jobID)
		} else {
			return errors.New("invalid job status")
		}
	} else {
		return err
	}
}

// RemoveJob remove job with job id
func (worker *Worker) RemoveJob(jobID string) error {
	qorJob, err := worker.GetJob(jobID)
	if err == nil {
		return qorJob.GetJob().GetQueue().Remove(qorJob)
	}
	return err
}
