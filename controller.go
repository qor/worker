package worker

import "github.com/qor/qor/admin"

type workerController struct {
	*Worker
}

func (workerController) index(context *admin.Context) {
	context.Execute("index", nil)
}

func (workerController) show(context *admin.Context) {
	context.Execute("show", nil)
}

func (workerController) addJob(context *admin.Context) {
}

func (workerController) runJob(context *admin.Context) {
}

func (workerController) killJob(context *admin.Context) {
}
