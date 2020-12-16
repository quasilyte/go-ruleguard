package worker

// NewWorker creates and returns a new worker with the specified name.
func NewWorker(typ string, name string) TaskExecuter {
	_id++
	switch typ {
	case "job":
		return &Job{Name: name, id: _id}
	case "task":
		return &Task{Name: name, id: _id}
	default:
		return nil
	}
}

var _id int

// Job performs work.
type Job struct {
	Name string
	id   int
}

// Run executes the work
func (j *Job) Run() error {
	return nil
}

// Task performs work.
type Task struct {
	Name string
	id   int
}

// Run executes the work
func (j *Task) Run() error {
	return nil
}

// TaskExecuter is implemented by any type that can execute a task.
type TaskExecuter interface {
	Run() error
}

// Other is a test struct that intentionally does not implement TaskExecutor.
type Other struct {
}

func (o Other) String() string {
	return "other"
}
