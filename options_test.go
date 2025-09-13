// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"testing"
)

func TestOption(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected Option
	}{
		{
			name:     "basic option",
			key:      "content-type",
			value:    "application/json",
			expected: Option{Key: "content-type", Value: "application/json"},
		},
		{
			name:     "empty key and value",
			key:      "",
			value:    "",
			expected: Option{Key: "", Value: ""},
		},
		{
			name:     "key with empty value",
			key:      "priority",
			value:    "",
			expected: Option{Key: "priority", Value: ""},
		},
		{
			name:     "empty key with value",
			key:      "",
			value:    "high",
			expected: Option{Key: "", Value: "high"},
		},
		{
			name:     "special characters in key and value",
			key:      "x-custom-header",
			value:    "value-with-dashes",
			expected: Option{Key: "x-custom-header", Value: "value-with-dashes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := Option{
				Key:   tt.key,
				Value: tt.value,
			}

			if option.Key != tt.expected.Key {
				t.Errorf("Option.Key = %v, want %v", option.Key, tt.expected.Key)
			}
			if option.Value != tt.expected.Value {
				t.Errorf("Option.Value = %v, want %v", option.Value, tt.expected.Value)
			}
		})
	}
}

func TestOption_StructFields(t *testing.T) {
	option := Option{
		Key:   "test-key",
		Value: "test-value",
	}

	// Test field accessibility
	if option.Key != "test-key" {
		t.Errorf("Option.Key field not accessible or incorrect value")
	}
	if option.Value != "test-value" {
		t.Errorf("Option.Value field not accessible or incorrect value")
	}

	// Test field modification
	option.Key = "modified-key"
	option.Value = "modified-value"

	if option.Key != "modified-key" {
		t.Errorf("Option.Key field not modifiable")
	}
	if option.Value != "modified-value" {
		t.Errorf("Option.Value field not modifiable")
	}
}

func TestOption_ZeroValue(t *testing.T) {
	var option Option

	// Test zero values
	if option.Key != "" {
		t.Errorf("Option zero value Key = %v, want empty string", option.Key)
	}
	if option.Value != "" {
		t.Errorf("Option zero value Value = %v, want empty string", option.Value)
	}
}