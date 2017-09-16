package main

import (
	"fmt"
)

func ShowCurrentCount() {
	fmt.Printf("Total (to the nearest hundred) prime numbers calculated: #%d\n", GetMaximumId())
}
