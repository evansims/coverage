package math

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("expected 5")
	}
}

func TestIsPositive(t *testing.T) {
	if !IsPositive(1) {
		t.Error("expected true")
	}
	if IsPositive(-1) {
		t.Error("expected false")
	}
	if IsPositive(0) {
		t.Error("expected false")
	}
}
