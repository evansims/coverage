package coverage

import (
	"errors"
	"fmt"
	"testing"
)

func TestThresholdErrorImplementsError(t *testing.T) {
	var err error = &ThresholdError{Message: "coverage below threshold"}
	if err.Error() != "coverage below threshold" {
		t.Errorf("Error() = %q, want %q", err.Error(), "coverage below threshold")
	}
}

func TestThresholdErrorAs(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &ThresholdError{Message: "coverage below threshold"})

	var target *ThresholdError
	if !errors.As(err, &target) {
		t.Fatal("errors.As did not match *ThresholdError")
	}
	if target.Message != "coverage below threshold" {
		t.Errorf("Message = %q, want %q", target.Message, "coverage below threshold")
	}
}

func TestConfigErrorImplementsError(t *testing.T) {
	var err error = &ConfigError{Message: "bad config"}
	if err.Error() != "bad config" {
		t.Errorf("Error() = %q, want %q", err.Error(), "bad config")
	}
}

func TestConfigErrorAs(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &ConfigError{Message: "bad config"})

	var target *ConfigError
	if !errors.As(err, &target) {
		t.Fatal("errors.As did not match *ConfigError")
	}
	if target.Message != "bad config" {
		t.Errorf("Message = %q, want %q", target.Message, "bad config")
	}
}

func TestConfigErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("underlying cause")
	err := &ConfigError{Message: "bad config", Cause: cause}

	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestConfigErrorUnwrapNil(t *testing.T) {
	err := &ConfigError{Message: "no cause"}

	if err.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
	}
}
