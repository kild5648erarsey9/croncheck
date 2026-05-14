package job

import (
	"fmt"
	"sync"
)

// Registry holds all monitored jobs, keyed by name.
type Registry struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		jobs: make(map[string]*Job),
	}
}

// Register adds a job to the registry. Returns an error if the name is already taken.
func (r *Registry) Register(j *Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.jobs[j.Name]; exists {
		return fmt.Errorf("job %q is already registered", j.Name)
	}
	r.jobs[j.Name] = j
	return nil
}

// Get returns the job with the given name, or an error if not found.
func (r *Registry) Get(name string) (*Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.jobs[name]
	if !ok {
		return nil, fmt.Errorf("job %q not found", name)
	}
	return j, nil
}

// All returns snapshots of every registered job.
func (r *Registry) All() []Job {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snaps := make([]Job, 0, len(r.jobs))
	for _, j := range r.jobs {
		snaps = append(snaps, j.Snapshot())
	}
	return snaps
}

// Len returns the number of registered jobs.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.jobs)
}
