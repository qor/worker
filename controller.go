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
	result := jobResource.NewStruct().(QorJobInterface)

	var job *Job
	for _, j := range wc.Worker.Jobs {
		if j.Name == context.Request.Form.Get("job_name") {
			job = j
		}
	}
	result.SetJob(job)

	if context.AddError(jobResource.Decode(context.Context, result)); !context.HasError() {
		// ensure job name is correct
		result.SetJob(job)
		context.AddError(jobResource.CallSave(result, context.Context))
		wc.Worker.AddJob(result)
	}

	http.Redirect(context.Writer, context.Request, path.Join(context.Request.URL.Path), http.StatusFound)
}

func (workerController) RunJob(context *admin.Context) {
}

func (workerController) KillJob(context *admin.Context) {
}
