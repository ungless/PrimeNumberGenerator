package server

import (
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"

	app "github.com/urfave/cli"
)

func readFromComputationsChannel() computation.Computation {
	return computation.Computation{primes.Prime{1, big.NewInt(101), 1 * time.Second}, big.NewInt(3), false, 0 * time.Second, computation.GenerateUUID()}
}

func handler(w http.ResponseWriter, r *http.Request) {
	_, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	fmt.Printf("%s: Sending computation\n", port)

	c := readFromComputationsChannel()
	json, err := computation.GetJSONFromComputation(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", json)
}

func LaunchServer(c *app.Context) {
	go fmt.Println("Launching server on port 8080...")
	numbersToCheck := make(chan *big.Int, 100)
	validPrimes := make(chan primes.Prime, 100)
	invalidPrimes := make(chan primes.Prime, 100)
	var primeBuffer storage.BigIntSlice

	go func() {
		for i := config.LastPrimeGenerated; true; i.Add(i, big.NewInt(2)) {
			numberToTest := big.NewInt(0).Set(i)
			numbersToCheck <- numberToTest
		}
	}()

	go func() {
		for elem := range validPrimes {
			if len(primeBuffer) == 1 {
				storage.FlushBufferToFile(primeBuffer)
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

	go func() {
		for i := range numbersToCheck {
			computationsToPerform := make(chan computation.Computation)
			currentSolvingPrime := primes.Prime{
				TimeTaken: 0 * time.Second,
				Value:     i,
			}
			go computation.GetComputationsToPerform(currentSolvingPrime, computationsToPerform)
		}
	}()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
