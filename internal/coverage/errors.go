package coverage

// ThresholdError indicates coverage was below a configured threshold.
// This maps to exit code 1.
type ThresholdError struct {
	Message string
}

func (e *ThresholdError) Error() string { return e.Message }

// ConfigError indicates a configuration or parsing problem.
// This maps to exit code 2.
type ConfigError struct {
	Message string
	Cause   error
}

func (e *ConfigError) Error() string { return e.Message }
func (e *ConfigError) Unwrap() error { return e.Cause }
