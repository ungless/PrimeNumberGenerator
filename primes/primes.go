package primes

import (
	"fmt"
	"math/big"
	"time"
)

type Prime struct {
	Id        uint64
	Value     *big.Int
	TimeTaken time.Duration
	IsValid   bool
}

// ChecknPrimality checks whether number is a prime.
func CheckPrimality(number *big.Int) bool {
	return number.ProbablyPrime(0)
}

// DisplayPrimePretty displays successful prime generations nicely.
func DisplayPrimePretty(number *big.Int, timeTaken time.Duration) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[4;30;42mSuccess\x1b[0m\t%s\x1b[0m\n",
		number,
		timeTaken,
	)
}

// DisplayFailPretty displays failed prime generations nicely.
func DisplayFailPretty(number *big.Int, timeTaken time.Duration) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[2;1;41mFail\x1b[0m\t%s\t\x1b[0m\n",
		number,
		timeTaken,
	)
}
