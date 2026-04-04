package main

import (
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
	fmt.Printf("Validators: %d, Byzantine: %d, Rounds: %d, Round time: %dms\n", n, f, *totalRounds, *roundTimeMs)

	validators := make([]*simulation.Validator, n)
	for i := range validators {
		byzantine := i >= n-f // last f validators are faulty
		validators[i] = simulation.NewValidator(i, f, byzantine)
	}

	// start the consensus simulation
	runSimulation(validators, *totalRounds, *roundTimeMs)
}

func runSimulation(validators []*simulation.Validator, totalRounds, roundTimeMs int) {
	for round := 1; round <= totalRounds; round++ {
		fmt.Printf("--- Round %d ---\n", round)

		for _, v := range validators {
			v.Propose(round)
		}

		time.Sleep(time.Duration(roundTimeMs) * time.Millisecond)
	}

	fmt.Println("Simulation complete")
}
