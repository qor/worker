# Worker

Worker runs a single Job in the background, it can do so immediately or at a scheduled time.

Once registered with QOR Admin, Worker will provide a `Workers` section in the navigation tree, containing pages for listing and managing the following aspects of Workers:

	- All Jobs.
	- Running: Jobs that are currently running.
	- Scheduled: Jobs which have been scheduled to run at a time in the future.
	- Done: finished Jobs.
	- Errors: any errors output from any Workers that have been run.

The admin interface for a schedulable Job will have an additional `Schedule Time` input, with which administrators can set the scheduled date and time.

[![GoDoc](https://godoc.org/github.com/qor/worker?status.svg)](https://godoc.org/github.com/qor/worker)

## Documentation

<https://doc.getqor.com/plugins/worker.html>

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
