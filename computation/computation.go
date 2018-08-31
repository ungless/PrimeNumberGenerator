package computation

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"

	"github.com/satori/go.uuid"
)

type Computation struct {
	Prime         primes.Prime
	Divisor       *big.Int
	IsValid       bool
	TimeTaken     time.Duration
	ComputationId *big.Int
	Hash          uuid.UUID
}

// GetUnMarshalledPrime produces a computation from a JSON string
func GetUnMarshalledPrime(body string) primes.Prime {
	var c primes.Prime
	err := json.Unmarshal([]byte(body), &c)
	if err != nil {
		config.Logger.Fatal(err)
	}
	return c
}

// GenerateUUID generates a new, random UUID (v4)
func GenerateUUID() uuid.UUID {
	u, err := uuid.NewV4()
	if err != nil {
		config.Logger.Fatal(err)
	}
	return u
}

// getComputation produces a Computation struct given its values
func getComputation(prime primes.Prime, divisor *big.Int, computationId *big.Int) Computation {
	nextUUID := GenerateUUID()
	computationStruct := Computation{prime, divisor, prime.IsValid, 0 * time.Second, computationId, nextUUID}
	return computationStruct
}

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

// RunDistributedComputation calculates the modulus of a given Computation
func RunDistributedComputation(c Computation) bool {
	var computationIsValid bool
	modulus := big.NewInt(0).Mod(c.Prime.Value, c.Divisor)
	if modulus.Cmp(big.NewInt(0)) == 0 {
		computationIsValid = true
	} else {
		computationIsValid = false
	}
	return computationIsValid
}

// getDivisorsOfPrime returns a slice containing all good divisors of a prime
func getDivisorsOfPrime(i *big.Int) storage.BigIntSlice {
	var divisorsOfPrime storage.BigIntSlice
	squareRoot := big.NewInt(0).Sqrt(i)
	for n := big.NewInt(3); n.Cmp(squareRoot) == -1; n.Add(n, big.NewInt(2)) {
		divisor := new(big.Int).Set(n)
		divisorsOfPrime = append(divisorsOfPrime, divisor)
	}
	return divisorsOfPrime
}

// GetComputationsToPerform passes all computations needed to be performed to a channel
func GetComputationsToPerform(prime primes.Prime) []Computation {
	var computations []Computation
	divisors := getDivisorsOfPrime(prime.Value)
	for i, v := range divisors {
		computationId := big.NewInt(int64(i))
		nextComputation := getComputation(prime, v, computationId)
		computations = append(computations, nextComputation)
	}
	return computations
}
