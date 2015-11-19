package worker

import "github.com/qor/qor/admin"

type workerController struct {
	*Worker
}

func (wc workerController) Index(context *admin.Context) {
	context.Execute("index", wc.Worker)
}

func (wc workerController) Show(context *admin.Context) {
	context.Execute("show", wc.Worker)
}

func (wc workerController) New(context *admin.Context) {
	context.Execute("new", wc.Worker)
}

func (workerController) AddJob(context *admin.Context) {
}

func (workerController) RunJob(context *admin.Context) {
}

func (workerController) KillJob(context *admin.Context) {
}
