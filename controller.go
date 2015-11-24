package worker

import (
	"net/http"

	"github.com/qor/qor/admin"
)

type workerController struct {
	*Worker
}

func (wc workerController) Index(context *admin.Context) {
	context.Execute("index", wc.Worker)
}

func (wc workerController) Show(context *admin.Context) {
	job, err := wc.GetJob(context.ResourceID)
	context.AddError(err)
	context.Execute("show", job)
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

	http.Redirect(context.Writer, context.Request, context.Request.URL.Path, http.StatusFound)
}

func (wc workerController) RunJob(context *admin.Context) {
	jobResource := wc.Worker.JobResource
	result := jobResource.NewStruct().(QorJobInterface)

	if job, err := wc.Worker.GetJob(context.ResourceID); context.AddError(err) {
		result.SetJob(job.GetJob())
		result.SetSerializeArgumentValue(result.GetArgument())
		context.AddError(jobResource.CallSave(result, context.Context))
		if context.HasError() {
			wc.Worker.AddJob(result)
		}
	}

	http.Redirect(context.Writer, context.Request, context.UrlFor(jobResource), http.StatusFound)
}

func (wc workerController) KillJob(context *admin.Context) {
	wc.Worker.KillJob(context.ResourceID)
	http.Redirect(context.Writer, context.Request, context.Request.URL.Path, http.StatusFound)
}
