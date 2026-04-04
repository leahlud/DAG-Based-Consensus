package main

import (
	"context"
	"dag-based-consensus/simulation"
	"flag"
	"fmt"
	"time"
)

func main() {
	// parse user arguments
	numByzantine := flag.Int("f", 1, "max byzantine faults tolerated")
	totalRounds := flag.Int("rounds", 10, "number of rounds to simulate")
	roundTimeMs := flag.Int("delay", 100, "round duration in ms")
	flag.Parse()

	// compute number of validators and initialize them
	f := *numByzantine
	n := 3*f + 1
	net := simulation.NewNetwork()
	validators := createValidators(n, f, net)
	net.Register(validators)

	// for testing
	fmt.Println("=== Setup ===")
	for _, v := range validators {
		if v.Byzantine {
			fmt.Printf("V%d (byzantine)\n", v.ID)
		} else {
			fmt.Printf("V%d (honest)\n", v.ID)
		}
	}

	// start each validator's listener as a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startValidators(ctx, validators)

	// start the consensus simulation
	runSimulation(validators, *totalRounds, *roundTimeMs)

	fmt.Println("\n=== DAG State ===")
	for _, v := range validators {
		v.PrintDAG()
	}
}

func runSimulation(validators []*simulation.Validator, totalRounds, roundTimeMs int) {
	for round := 1; round <= totalRounds; round++ {
		for _, v := range validators {
			v.Propose(round)
		}

		time.Sleep(time.Duration(roundTimeMs) * time.Millisecond)
	}
}

func createValidators(n, f int, net *simulation.Network) []*simulation.Validator {
	validators := make([]*simulation.Validator, n)
	for i := range validators {
		validators[i] = simulation.NewValidator(i, f, i >= n-f, net)
	}
	return validators
}

func startValidators(ctx context.Context, validators []*simulation.Validator) {
	for _, v := range validators {
		go v.Listen(ctx)
	}
}
