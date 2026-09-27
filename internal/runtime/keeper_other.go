//go:build !windows

package runtime

// RunJobKeeper is only implemented for the Windows runtime's independent Job
// handle keeper. Hidden CLI dispatch must exit before opening daemon services.
func RunJobKeeper([]string) error { return ErrCapability }
