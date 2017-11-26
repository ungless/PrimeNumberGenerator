package main

import (
	"fmt"
)

func ShowCurrentCount() {
	fmt.Printf("Total (to the nearest hundred) prime numbers calculated and stored: #%d\n", GetMaximumId())
}
