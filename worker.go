package worker

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/roles"
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

func (worker *Worker) ConfigureQorResource(res *admin.Resource) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/worker/views"))
	}
	res.UseTheme("worker")

	worker.Admin = res.GetAdmin()
	worker.JobResource = worker.Admin.NewResource(worker.Config.Job)
	worker.JobResource.Meta(&admin.Meta{Name: "Name", Valuer: func(record interface{}, context *qor.Context) interface{} {
		return record.(QorJobInterface).GetJobName()
	}})
	worker.JobResource.IndexAttrs("ID", "Name", "Status")
	worker.JobResource.Permission = roles.Allow(roles.Update, "no_body").Allow(roles.Delete, "no_body")

	for _, status := range []string{"new", "running", "done", "exception"} {
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
	var qorJobID = flag.Int("qor-job", 0, "Qor Job ID")
	flag.Parse()
	if qorJobID != nil && *qorJobID != 0 {
		if err := worker.RunJob(uint(*qorJobID)); err == nil {
			os.Exit(0)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// configure routes
	router := worker.Admin.GetRouter()
	controller := workerController{Worker: worker}

	router.Get("/"+res.ToParam()+"/new", controller.New)
	router.Get("/"+res.ToParam()+"/.*$", controller.Show)
	router.Get("/"+res.ToParam(), controller.Index)
	router.Post("/"+res.ToParam()+"/.*/run$", controller.RunJob)
	router.Post("/"+res.ToParam()+"/.*/kill$", controller.KillJob)
	router.Post("/"+res.ToParam()+"$", controller.AddJob)
}

func (worker *Worker) SetQueue(queue Queue) {
	worker.Queue = queue
}

func (worker *Worker) RegisterJob(job Job) error {
	job.Worker = worker
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

func (worker *Worker) GetJob(jobID uint) (QorJobInterface, error) {
	qorJob := worker.JobResource.NewStruct().(QorJobInterface)

	context := worker.Admin.NewContext(nil, nil)
	context.ResourceID = fmt.Sprint(jobID)
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

func (worker *Worker) UpdateJobStatus(qorJob QorJobInterface, status string) error {
	context := worker.Admin.NewContext(nil, nil).Context
	qorJob.SetStatus(status)
	return worker.JobResource.CallSave(qorJob, context)
}

func (worker *Worker) RunJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		if err := worker.UpdateJobStatus(qorJob, "running"); err == nil {
			if err := qorJob.GetJob().Run(qorJob.GetSerializeArgument(qorJob)); err == nil {
				return worker.UpdateJobStatus(qorJob, "done")
			} else {
				worker.UpdateJobStatus(qorJob, "exception")
				return err
			}
		} else {
			return err
		}
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
