package computation

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/satori/go.uuid"
)

type Computation struct {
	Prime   primes.Prime
	Divisor *big.Int
	Hash    uuid.UUID
}

func GetJSON(c Computation) ([]byte, error) {
	json, err := json.Marshal(c)
	return json, err
}

func GenerateUUID() uuid.UUID {
	u := uuid.NewV4()
	return u
}

func GetNextComputation() Computation {
	return Computation{primes.Prime{1, big.NewInt(101), 1 * time.Second}, big.NewInt(3), GenerateUUID()}
}
