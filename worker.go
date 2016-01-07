package worker

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

const (
	JobStatusNew       = "new"
	JobStatusRunning   = "running"
	JobStatusDone      = "done"
	JobStatusException = "exception"
	JobStatusKilled    = "killed"
)

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

type Config struct {
	Queue Queue
	Job   QorJobInterface
	Admin *admin.Admin
}

type Worker struct {
	*Config
	JobResource *admin.Resource
	Jobs        []*Job
}

func (worker *Worker) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/worker/views"))
		}
		res.UseTheme("worker")

		worker.Admin = res.GetAdmin()
		worker.JobResource = worker.Admin.NewResource(worker.Config.Job)
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

		// Auto Migration
		worker.Admin.Config.DB.AutoMigrate(worker.Config.Job)

		// Configure jobs
		for _, job := range worker.Jobs {
			if job.Resource == nil {
				job.Resource = worker.Admin.NewResource(worker.JobResource.Value)
			}
		}

		// Parse job
		var qorJobID = flag.String("qor-job", "", "Qor Job ID")
		var runAnother = flag.Bool("run-another", false, "Run another qor job")
		flag.Parse()
		if *qorJobID != "" {
			if *runAnother == true {
				if newJob := worker.SaveAnotherJob(*qorJobID); newJob != nil {
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

func (worker *Worker) SetQueue(queue Queue) {
	worker.Queue = queue
}

func (worker *Worker) RegisterJob(job Job) error {
	job.Worker = worker
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

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
	}
	return nil, fmt.Errorf("failed to find job: %v", jobID)
}

func (worker *Worker) AddJob(qorJob QorJobInterface) error {
	return worker.Queue.Add(qorJob)
}

func (worker *Worker) RunJob(jobID string) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		defer func() {
			if r := recover(); r != nil {
				qorJob.SetProgressText(fmt.Sprint(r))
				qorJob.SetStatus(JobStatusException)
			}
		}()

		if qorJob.GetStatus() != JobStatusNew {
			return errors.New("invalid job status, current status: " + qorJob.GetStatus())
		}

		if err := qorJob.SetStatus(JobStatusRunning); err == nil {
			if err := qorJob.GetJob().GetQueue().Run(qorJob); err == nil {
				return qorJob.SetStatus(JobStatusDone)
			} else {
				qorJob.SetProgressText(err.Error())
				qorJob.SetStatus(JobStatusException)
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func (worker *Worker) SaveAnotherJob(jobID string) QorJobInterface {
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

func (worker *Worker) KillJob(jobID string) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		if qorJob.GetStatus() == JobStatusRunning {
			if err := qorJob.GetJob().GetQueue().Kill(qorJob); err == nil {
				qorJob.SetStatus(JobStatusKilled)
				return nil
			} else {
				return err
			}
		} else if qorJob.GetStatus() == JobStatusNew {
			return worker.RemoveJob(jobID)
		} else {
			return errors.New("invalid job status")
		}
	} else {
		return err
	}
}

func (worker *Worker) RemoveJob(jobID string) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetJob().GetQueue().Remove(qorJob)
	} else {
		return err
	}
}
