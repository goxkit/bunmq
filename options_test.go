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
				Key:   OptionKey(tt.key),
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
	if option.Value != nil {
		t.Errorf("Option zero value Value = %v, want nil", option.Value)
	}
}

func TestDeliveryModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		mode     DeliveryMode
		expected uint8
	}{
		{
			name:     "transient delivery mode",
			mode:     DeliveryModeTransient,
			expected: 1,
		},
		{
			name:     "persistent delivery mode",
			mode:     DeliveryModePersistent,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint8(tt.mode) != tt.expected {
				t.Errorf("DeliveryMode = %v, want %v", uint8(tt.mode), tt.expected)
			}
		})
	}
}

func TestOptionKeyConstants(t *testing.T) {
	tests := []struct {
		name     string
		key      OptionKey
		expected string
	}{
		{
			name:     "delivery mode key",
			key:      OptionDeliveryModeKey,
			expected: "DeliveryMode",
		},
		{
			name:     "headers key",
			key:      OptionHeadersKey,
			expected: "Headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.key) != tt.expected {
				t.Errorf("OptionKey = %v, want %v", string(tt.key), tt.expected)
			}
		})
	}
}

func TestOptionPersistentDeliveryMode(t *testing.T) {
	options := OptionPersistentDeliveryMode()

	if len(options) != 1 {
		t.Errorf("OptionPersistentDeliveryMode() returned %d options, want 1", len(options))
		return
	}

	option := options[0]
	if option.Key != OptionDeliveryModeKey {
		t.Errorf("Option.Key = %v, want %v", option.Key, OptionDeliveryModeKey)
	}
	if option.Value != DeliveryModePersistent {
		t.Errorf("Option.Value = %v, want %v", option.Value, DeliveryModePersistent)
	}
}

func TestOptionTransientDeliveryMode(t *testing.T) {
	options := OptionTransientDeliveryMode()

	if len(options) != 1 {
		t.Errorf("OptionTransientDeliveryMode() returned %d options, want 1", len(options))
		return
	}

	option := options[0]
	if option.Key != OptionDeliveryModeKey {
		t.Errorf("Option.Key = %v, want %v", option.Key, OptionDeliveryModeKey)
	}
	if option.Value != DeliveryModeTransient {
		t.Errorf("Option.Value = %v, want %v", option.Value, DeliveryModeTransient)
	}
}

func TestOptionHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]any
	}{
		{
			name:    "empty headers",
			headers: map[string]any{},
		},
		{
			name: "single header",
			headers: map[string]any{
				"content-type": "application/json",
			},
		},
		{
			name: "multiple headers",
			headers: map[string]any{
				"content-type":   "application/json",
				"priority":       10,
				"x-custom-bool":  true,
				"x-custom-float": 3.14,
			},
		},
		{
			name: "headers with nil values",
			headers: map[string]any{
				"nil-header":   nil,
				"valid-header": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := OptionHeaders(tt.headers)

			if len(options) != 1 {
				t.Errorf("OptionHeaders() returned %d options, want 1", len(options))
				return
			}

			option := options[0]
			if option.Key != OptionHeadersKey {
				t.Errorf("Option.Key = %v, want %v", option.Key, OptionHeadersKey)
			}

			// Verify the headers map is properly set
			headers, ok := option.Value.(map[string]any)
			if !ok {
				t.Errorf("Option.Value is not of type map[string]any, got %T", option.Value)
				return
			}

			if len(headers) != len(tt.headers) {
				t.Errorf("Headers length = %d, want %d", len(headers), len(tt.headers))
				return
			}

			for key, expectedValue := range tt.headers {
				actualValue, exists := headers[key]
				if !exists {
					t.Errorf("Header key %v not found in result", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Header[%v] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestPublisherOptions(t *testing.T) {
	tests := []struct {
		name         string
		deliveryMode DeliveryMode
		headers      map[string]any
	}{
		{
			name:         "transient mode with empty headers",
			deliveryMode: DeliveryModeTransient,
			headers:      map[string]any{},
		},
		{
			name:         "persistent mode with headers",
			deliveryMode: DeliveryModePersistent,
			headers: map[string]any{
				"content-type": "application/json",
				"priority":     5,
			},
		},
		{
			name:         "transient mode with nil headers",
			deliveryMode: DeliveryModeTransient,
			headers:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := PublisherOptions(tt.deliveryMode, tt.headers)

			if len(options) != 2 {
				t.Errorf("PublisherOptions() returned %d options, want 2", len(options))
				return
			}

			// Check delivery mode option
			deliveryModeOption := options[0]
			if deliveryModeOption.Key != OptionDeliveryModeKey {
				t.Errorf("First option Key = %v, want %v", deliveryModeOption.Key, OptionDeliveryModeKey)
			}
			if deliveryModeOption.Value != tt.deliveryMode {
				t.Errorf("First option Value = %v, want %v", deliveryModeOption.Value, tt.deliveryMode)
			}

			// Check headers option
			headersOption := options[1]
			if headersOption.Key != OptionHeadersKey {
				t.Errorf("Second option Key = %v, want %v", headersOption.Key, OptionHeadersKey)
			}

			headers, ok := headersOption.Value.(map[string]any)
			if !ok {
				t.Errorf("Second option Value is not of type map[string]any, got %T", headersOption.Value)
				return
			}

			if tt.headers == nil {
				if headers != nil {
					t.Errorf("Expected nil headers, got %v", headers)
				}
			} else {
				if len(headers) != len(tt.headers) {
					t.Errorf("Headers length = %d, want %d", len(headers), len(tt.headers))
					return
				}

				for key, expectedValue := range tt.headers {
					actualValue, exists := headers[key]
					if !exists {
						t.Errorf("Header key %v not found in result", key)
						continue
					}
					if actualValue != expectedValue {
						t.Errorf("Header[%v] = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestOptionKey_String(t *testing.T) {
	tests := []struct {
		name     string
		key      OptionKey
		expected string
	}{
		{
			name:     "delivery mode key string conversion",
			key:      OptionDeliveryModeKey,
			expected: "DeliveryMode",
		},
		{
			name:     "headers key string conversion",
			key:      OptionHeadersKey,
			expected: "Headers",
		},
		{
			name:     "custom key string conversion",
			key:      OptionKey("custom-key"),
			expected: "custom-key",
		},
		{
			name:     "empty key string conversion",
			key:      OptionKey(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.key)
			if result != tt.expected {
				t.Errorf("string(OptionKey) = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDeliveryMode_Types(t *testing.T) {
	// Test that DeliveryMode is properly typed as uint8
	transient := DeliveryModeTransient
	persistent := DeliveryModePersistent

	// These should compile without type conversion
	u1 := uint8(transient)
	u2 := uint8(persistent)

	if u1 != 1 {
		t.Errorf("DeliveryModeTransient as uint8 = %d, want 1", u1)
	}
	if u2 != 2 {
		t.Errorf("DeliveryModePersistent as uint8 = %d, want 2", u2)
	}
}

func TestOption_WithDifferentValueTypes(t *testing.T) {
	tests := []struct {
		name       string
		key        OptionKey
		value      any
		comparable bool
	}{
		{
			name:       "string value",
			key:        "string-key",
			value:      "string-value",
			comparable: true,
		},
		{
			name:       "int value",
			key:        "int-key",
			value:      42,
			comparable: true,
		},
		{
			name:       "bool value",
			key:        "bool-key",
			value:      true,
			comparable: true,
		},
		{
			name:       "float value",
			key:        "float-key",
			value:      3.14,
			comparable: true,
		},
		{
			name:       "map value",
			key:        "map-key",
			value:      map[string]string{"nested": "value"},
			comparable: false,
		},
		{
			name:       "slice value",
			key:        "slice-key",
			value:      []string{"item1", "item2"},
			comparable: false,
		},
		{
			name:       "nil value",
			key:        "nil-key",
			value:      nil,
			comparable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := Option{
				Key:   tt.key,
				Value: tt.value,
			}

			if option.Key != tt.key {
				t.Errorf("Option.Key = %v, want %v", option.Key, tt.key)
			}

			// For comparable types, use direct comparison
			if tt.comparable {
				if option.Value != tt.value {
					t.Errorf("Option.Value = %v, want %v", option.Value, tt.value)
				}
			} else {
				// For non-comparable types, just check they were set (not nil)
				if option.Value == nil {
					t.Errorf("Option.Value should not be nil for %v", tt.name)
				}
				// Additional type-specific checks
				switch v := tt.value.(type) {
				case map[string]string:
					optMap, ok := option.Value.(map[string]string)
					if !ok {
						t.Errorf("Option.Value is not a map[string]string")
					} else if len(optMap) != len(v) {
						t.Errorf("Option.Value map length = %d, want %d", len(optMap), len(v))
					}
				case []string:
					optSlice, ok := option.Value.([]string)
					if !ok {
						t.Errorf("Option.Value is not a []string")
					} else if len(optSlice) != len(v) {
						t.Errorf("Option.Value slice length = %d, want %d", len(optSlice), len(v))
					}
				}
			}
		})
	}
}
