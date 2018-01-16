package bahamut

import "context"

// Job is the type of function that can be run as a Job.
type Job func() error

// RunJob runs a Job can than be canceled at any time according to the context.
func RunJob(ctx context.Context, job Job) (bool, error) {

	out := make(chan error)

	go func() { out <- job() }()

	select {
	case <-ctx.Done():
		return true, nil
	case err := <-out:
		return false, err
	}
}
