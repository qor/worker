package kubernetes

// Kubernetes implemented a worker Queue based on kubernetes jobs
type Kubernetes struct {
	Config *Config
}

// Config kubernetes config
type Config struct {
}

// New initialize Kubernetes
func New(config *Config) *Kubernetes {
	return &Kubernetes{Config: config}
}
