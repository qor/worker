package worker

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

func New(config Config) *Worker {
	if config.Job == nil {
		config.Job = &QorJob{}
	}

	// Auto Migration
	config.DB.AutoMigrate(config.Job)

	return &Worker{Config: &config}
}

type Config struct {
	DB    *gorm.DB
	Queue Queue
	Job   QorJobInterface
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

	Admin := res.GetAdmin()
	worker.JobResource = Admin.NewResource(worker.Config.Job)
	worker.JobResource.Meta(&admin.Meta{Name: "Name", Valuer: func(record interface{}, context *qor.Context) interface{} {
		return record.(QorJobInterface).GetJobName()
	}})
	worker.JobResource.IndexAttrs("ID", "Name", "Status")

	// configure jobs
	for _, job := range worker.Jobs {
		if job.Resource == nil {
			job.Resource = Admin.NewResource(worker.JobResource.Value)
		}

		job.Resource.Meta(&admin.Meta{
			Name: "_job_name",
			Type: "hidden",
			Valuer: func(interface{}, *qor.Context) interface{} {
				return job.Name
			},
			Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
				resource.(QorJobInterface).SetJobName(utils.ToString(metaValue.Value))
			},
		})
	}

	// configure routes
	router := Admin.GetRouter()
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
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

func (worker *Worker) GetJob(jobID uint) (QorJobInterface, error) {
	var qorJob QorJobInterface

	if err := worker.DB.First(&qorJob, jobID).Error; err == nil {
		for _, job := range worker.Jobs {
			if job.Name == qorJob.GetJobName() {
				qorJob.SetJob(job)
				return qorJob, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find job: %v", jobID)
}

func (worker *Worker) AddJob(QorJobInterface) error {
	return nil
}

func (worker *Worker) RunJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetJob().Run(qorJob.GetSerializeArgument(qorJob))
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
