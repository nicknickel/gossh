package config

import "testing"

func TestNormalizeString(t *testing.T) {
	testCases := []string{
		"examplestring",
		"example_string",
		"example-string",
		"exampl_e-string",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			result := NormalizeString(testCase)
			if result != "examplestring" {
				t.Errorf("expected NormalizeString(%v) to be examplestring but got %v", testCase, result)
			}
		})
	}

	result := NormalizeString("")
	if result != "" {
		t.Errorf("expected NormalizeString(nil) to be nil but got %v", result)
	}
}
