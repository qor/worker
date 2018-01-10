package kubernetes

import (
	"errors"

	"github.com/qor/worker"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ worker.Queue = &Kubernetes{}

// Kubernetes implemented a worker Queue based on kubernetes jobs
type Kubernetes struct {
	Clientset *kubernetes.Clientset
	Config    *Config
}

// Config kubernetes config
type Config struct {
	Namespace     string
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
func (k8s *Kubernetes) Add(qorJob worker.QorJobInterface) error {
	// TODO CronJob
	k8sJob := &v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "job id",
			Labels: map[string]string{},
		},
		Spec: v1.JobSpec{
			Selector: &metav1.LabelSelector{
			// from config?
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "job id",
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
				// TODO
				},
			},
		},
	}
	_, err := k8s.Clientset.Batch().Jobs(k8s.Config.Namespace).Create(k8sJob)
	return err
}

// Run a job from k8s queue
func (k8s *Kubernetes) Run(qorJob worker.QorJobInterface) error {
	job := qorJob.GetJob()

	if job.Handler != nil {
		return job.Handler(qorJob.GetSerializableArgument(qorJob), qorJob)
	}

	return errors.New("no handler found for job " + job.Name)
}

// Kill a job from k8s queue
func (k8s *Kubernetes) Kill(job worker.QorJobInterface) error {
	return k8s.Clientset.Core().Pods(k8s.Config.Namespace).Delete("job name", &metav1.DeleteOptions{})
}

// Remove a job from k8s queue
func (k8s *Kubernetes) Remove(job worker.QorJobInterface) error {
	// Don't remove if it is already running
	return k8s.Clientset.Core().Pods(k8s.Config.Namespace).Delete("job name", &metav1.DeleteOptions{})
}
