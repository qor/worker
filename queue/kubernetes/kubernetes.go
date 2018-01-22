package kubernetes

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

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
	Image         string
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

// GetCurrentPod get current pod
func (k8s *Kubernetes) GetCurrentPod() *corev1.Pod {
	var (
		podlist, err = k8s.Clientset.Core().Pods("").List(metav1.ListOptions{})
		localeIP     = GetLocalIP()
	)

	for _, item := range podlist.Items {
		if item.Status.PodIP == localeIP {
			return &item
		}
	}
	return nil
}

// Add a job to k8s queue
func (k8s *Kubernetes) Add(qorJob worker.QorJobInterface) error {
	jobName := fmt.Sprintf("qor-job-%v", qorJob.GetJobID())

	currentPath, _ := os.Getwd()
	binaryFile, err := filepath.Abs(os.Args[0])

	// TODO K8s CronJob
	k8sJob := &v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: map[string]string{},
		},
		Spec: v1.JobSpec{
			Selector: &metav1.LabelSelector{
			// from config?
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   jobName,
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:       jobName,
						Image:      k8s.Config.Image,
						Command:    []string{binaryFile, "--qor-job", qorJob.GetJobID()},
						WorkingDir: currentPath,
					}},
					RestartPolicy: "Never",
				},
			},
		},
	}
	_, err = k8s.Clientset.Batch().Jobs(k8s.Config.Namespace).Create(k8sJob)
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

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
