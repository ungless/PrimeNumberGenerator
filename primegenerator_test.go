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
		isPrime := checkPrimality(value)
		if isPrime != expected {
			t.Errorf("Expected checkPrime(%d) to be %t, instead got %t", value, expected, isPrime)
		}
	}
}

func TestFormatFilename(t *testing.T) {
	value, expected := "0-1000000", "/home/max/.primes/0-1000000.txt"
	formatFilePath := formatFilePath(value)
	if formatFilePath != expected {
		t.Errorf("Expected %s, got %s, with %s.", expected, formatFilePath, value)
	}
}

func BenchmarkPrimeAssertion(b *testing.B) {
	ComputePrimes(big.NewInt(1), false, false, big.NewInt(int64(b.N)))
}
