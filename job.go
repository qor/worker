package worker

import (
	"github.com/qor/qor/admin"
	"github.com/qor/roles"
)

// Job defined qor job struct
type Job struct {
	Name       string
	Group      string
	Handler    func(interface{}, QorJobInterface) error
	Permission roles.Permission
	Queue      Queue
	Resource   *admin.Resource
	Worker     *Worker
}

func (job *Job) NewStruct() interface{} {
	qorJobInterface := job.Worker.JobResource.NewStruct().(QorJobInterface)
	qorJobInterface.SetJob(job)
	return qorJobInterface
}

func (job *Job) GetQueue() Queue {
	if job.Queue != nil {
		return job.Queue
	}
	return job.Worker.Queue
}
