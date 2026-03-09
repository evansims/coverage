package coverage

import (
	"strings"
	"testing"
)

func TestGetParser(t *testing.T) {
	formats := []string{"lcov", "gocover", "cobertura", "clover", "jacoco"}
	for _, f := range formats {
		t.Run(f, func(t *testing.T) {
			p, err := getParser(f)
			if err != nil {
				t.Fatalf("getParser(%q) returned error: %v", f, err)
			}
			if p == nil {
				t.Fatalf("getParser(%q) returned nil", f)
			}
		})
	}

	t.Run("unknown format", func(t *testing.T) {
		_, err := getParser("unknown")
		if err == nil {
			t.Fatal("expected error for unknown format")
		}
	})
}

func TestRejectXMLEntities(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "clean XML",
			data:    `<?xml version="1.0"?><coverage></coverage>`,
			wantErr: false,
		},
		{
			name:    "DOCTYPE without entities is allowed",
			data:    `<?xml version="1.0"?><!DOCTYPE report PUBLIC "-//JACOCO//DTD" "report.dtd"><report></report>`,
			wantErr: false,
		},
		{
			name:    "DOCTYPE with ENTITY declaration",
			data:    `<?xml version="1.0"?><!DOCTYPE coverage [<!ENTITY a "x">]><coverage></coverage>`,
			wantErr: true,
		},
		{
			name:    "standalone ENTITY declaration",
			data:    `<?xml version="1.0"?><!ENTITY a "x"><coverage></coverage>`,
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rejectXMLEntities([]byte(tt.data))
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRejectXMLEntitiesBeyondPrefix(t *testing.T) {
	// Verify that entity detection works even when the ENTITY declaration
	// appears far into the file (beyond any prefix-based scan limit).
	prefix := make([]byte, 8192)
	for i := range prefix {
		prefix[i] = ' '
	}
	data := append(prefix, []byte("<!ENTITY bomb \"boom\">")...)
	err := rejectXMLEntities(data)
	if err == nil {
		t.Fatal("expected error for ENTITY declaration beyond 4096 bytes")
	}
}

func TestXMLParsersRejectEntities(t *testing.T) {
	bomb := []byte(`<?xml version="1.0"?><!DOCTYPE coverage [<!ENTITY a "x">]><coverage></coverage>`)

	for _, name := range []string{"cobertura", "clover", "jacoco"} {
		t.Run(name, func(t *testing.T) {
			parser, _ := getParser(name)
			_, err := parser(bomb)
			if err == nil {
				t.Fatal("expected error for XML with ENTITY")
			}
			if !strings.Contains(err.Error(), "ENTITY") {
				t.Errorf("error should mention ENTITY: %v", err)
			}
		})
	}
}
