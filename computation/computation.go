package computation

import (
	"math/big"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"
)

// ComputePrimes computes primes concurrently until KeyboardInterrupt
func ComputePrimes(lastPrime *big.Int, writeToFile bool, toInfinity bool, maxNumber *big.Int) {
	numbersToCheck := make(chan *big.Int, 100)
	validPrimes := make(chan primes.Prime, 100)
	invalidPrimes := make(chan primes.Prime, 100)
	var primeBuffer storage.BigIntSlice

	go func() {
		if toInfinity {
			for i := lastPrime; true; i.Add(i, big.NewInt(2)) {
				numberToTest := big.NewInt(0).Set(i)
				numbersToCheck <- numberToTest
			}
		} else {
			for i := lastPrime; i.Cmp(maxNumber) == -1; i.Add(i, big.NewInt(2)) {
				numberToTest := big.NewInt(0).Set(i)
				numbersToCheck <- numberToTest
			}
		}
	}()

	go func() {
		for elem := range validPrimes {
			primeBuffer = append(primeBuffer, elem.Value)
			if len(primeBuffer) == config.MaxBufferSize {
				if writeToFile {
					storage.FlushBufferToFile(primeBuffer)
				}
				primeBuffer = nil
			}
			primes.DisplayPrimePretty(elem.Value, elem.TimeTaken)
		}
	}()

	go func() {
		for elem := range invalidPrimes {
			if config.ShowFails == true {
				primes.DisplayFailPretty(elem.Value, elem.TimeTaken)
			}
		}
	}()

	for i := range numbersToCheck {
		go func(i *big.Int) {
			start := time.Now()
			isPrime := primes.CheckPrimality(i)
			if isPrime == true {
				validPrimes <- primes.Prime{
					TimeTaken: time.Now().Sub(start),
					Value:     i,
					Id:        config.Id,
				}
			} else {
				invalidPrimes <- primes.Prime{
					TimeTaken: time.Now().Sub(start),
					Value:     i,
				}
			}
		}(i)
	}
}
