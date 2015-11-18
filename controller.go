package worker

import "github.com/qor/qor/admin"

type workerController struct {
	*Worker
}

func (workerController) Index(context *admin.Context) {
	context.Execute("index", nil)
}

func (workerController) Show(context *admin.Context) {
	context.Execute("show", nil)
}

func (workerController) New(context *admin.Context) {
	context.Execute("new", nil)
}

func (workerController) AddJob(context *admin.Context) {
}

func (workerController) RunJob(context *admin.Context) {
}

func (workerController) KillJob(context *admin.Context) {
}
