package worker

type Queue interface {
	Kill(*QorJob) error
	Add(i *QorJob) error
	Delete(*QorJob) error
}
