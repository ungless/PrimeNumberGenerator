package main

import (
	"math/big"
	"testing"
)

type test struct {
	test      int64
	expecting bool
}

var tests = []test{
	{3, true},
	{7, true},
	{15, false},
	{21, false},
}

func TestCorrectPrimeNumberOutput(t *testing.T) {
	for _, test := range tests {
		value, expected := big.NewInt(test.test), test.expecting
		is_prime := check_prime(value)
		if is_prime != expected {
			t.Errorf("Expected check_prime(%d) to be %t, instead got %t", value, expected, is_prime)
		}
	}
}

func TestFormatFilename(t *testing.T) {
	value, expected := "0-1000000", "primes/0-1000000.txt"
	formatted_filename := format_filename(value)
	if formatted_filename != expected {
		t.Errorf("Expected %s, got %s, with %s.", expected, formatted_filename, value)
	}
}
