package worker

type Queue interface {
	Kill(QorJobInterface) error
	Add(QorJobInterface) error
	Delete(QorJobInterface) error
}
