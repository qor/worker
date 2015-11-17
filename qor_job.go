package worker

import "github.com/jinzhu/gorm"

type QorJob struct {
	gorm.Model
	Name     string
	Status   string
	Argument interface{}
	*Job
}
