package utils

import (
	"fmt"
	"testing"
)

func TestCalcFactorial(t *testing.T) {
	tests := []struct {
		name     string
		n        int64
		expected string
	}{
		{"0!", -1, "0"},
		{"0!", 0, "1"},
		{"1!", 1, "1"},
		{"5!", 5, "120"},
		{"10!", 10, "3628800"},
		{"40!", 40, "815915283247897734345611269596115894272000000000"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%s", i, test.name), func(t *testing.T) {
			result := CalcFactorial(test.n)
			if result.String() != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result.String())
			}
		})
	}
}
