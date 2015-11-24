package worker

type Queue interface {
	Kill(QorJobInterface) error
	Add(QorJobInterface) error
	Remove(QorJobInterface) error
}
