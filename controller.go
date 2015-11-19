package worker

import (
	"net/http"
	"path"

	"github.com/qor/qor/admin"
)

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

func (wc workerController) AddJob(context *admin.Context) {
	jobResource := wc.Worker.JobResource
	jobResourceult := jobResource.NewStruct()
	if context.AddError(jobResource.Decode(context.Context, jobResourceult)); !context.HasError() {
		context.AddError(jobResource.CallSave(jobResourceult, context.Context))
	}

	http.Redirect(context.Writer, context.Request, path.Join(context.Request.URL.Path), http.StatusFound)
}

func (workerController) RunJob(context *admin.Context) {
}

func (workerController) KillJob(context *admin.Context) {
}
