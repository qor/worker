package kubernetes

import (
	"github.com/qor/worker"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Kubernetes implemented a worker Queue based on kubernetes jobs
type Kubernetes struct {
	Clientset *kubernetes.Clientset
	Config    *Config
}

// Config kubernetes config
type Config struct {
	ClusterConfig *rest.Config
}

// New initialize Kubernetes
func New(config *Config) (*Kubernetes, error) {
	var err error

	if config == nil {
		config = &Config{}
	}

	if config.ClusterConfig == nil {
		config.ClusterConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config.ClusterConfig)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{Clientset: clientset, Config: config}, nil
}

// Add a job to k8s queue
func (k8s *Kubernetes) Add(worker.QorJobInterface) error {
	return nil
}

// Run a job from k8s queue
func (k8s *Kubernetes) Run(worker.QorJobInterface) error {
	return nil
}

// Kill a job from k8s queue
func (k8s *Kubernetes) Kill(worker.QorJobInterface) error {
	return nil
}

// Remove a job from k8s queue
func (k8s *Kubernetes) Remove(worker.QorJobInterface) error {
	return nil
}
