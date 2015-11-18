package worker

type Queue interface {
	Kill(QorJob) error
	Add(QorJob) error
	Delete(QorJob) error
}
