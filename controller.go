package worker

import (
	"errors"
	"net/http"

	"github.com/qor/qor/admin"
	"github.com/qor/responder"
)

type workerController struct {
	*Worker
}

func (wc workerController) Index(context *admin.Context) {
	context = context.NewResourceContext(wc.JobResource)
	result, err := context.FindMany()
	context.AddError(err)

	if context.HasError() {
		http.NotFound(context.Writer, context.Request)
	} else {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			context.JSON("index", result)
		}).Respond(context.Request)
	}
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
		context.AddError(wc.Worker.AddJob(result))
	}

	if !context.HasError() {
		context.Flash(string(context.Admin.T(context.Context, "resource_successfully_created", "{{.Name}} was successfully created", jobResource)), "success")
	}

	http.Redirect(context.Writer, context.Request, context.Request.URL.Path, http.StatusFound)
}

func (wc workerController) RunJob(context *admin.Context) {
	if newJob := wc.Worker.saveAnotherJob(context.ResourceID); newJob != nil {
		wc.Worker.AddJob(newJob)
	} else {
		context.AddError(errors.New("failed to clone job " + context.ResourceID))
	}

	http.Redirect(context.Writer, context.Request, context.URLFor(wc.Worker.JobResource), http.StatusFound)
}

func (wc workerController) KillJob(context *admin.Context) {
	context.AddError(wc.Worker.KillJob(context.ResourceID))
	http.Redirect(context.Writer, context.Request, context.Request.URL.Path, http.StatusFound)
}
