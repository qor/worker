package worker

type QorJobInterface interface {
	SetArgument(argument interface{})
	GetArgument() interface{}
	SetStatus(string) error
	GetStatus() string
	SetJobName(string) error
	GetJobName() string
	GetJob() *Job
}
