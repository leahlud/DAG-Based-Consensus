package main

import (
	"context"
	"dag-based-consensus/export"
	"dag-based-consensus/simulation"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

var f = 1       // num byzantine
var n = 3*f + 1 // total number of validators

func main() {
	// parse user arguments
	numByzantine := flag.Int("f", 1, "max byzantine faults tolerated")
	totalRounds := flag.Int("rounds", 10, "number of rounds to simulate")
	roundTimeMs := flag.Int("delay", 100, "round duration in ms")
	proposeProb := flag.Float64("p", 1.0, "probability a validator proposes in a round")
	flag.Parse()

	// compute number of validators and initialize them
	f = *numByzantine
	n = 3*f + 1
	net := simulation.NewNetwork()
	validators := createValidators(net)
	net.Register(validators)

	// for testing
	fmt.Println("--- Setup ---")
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
	runSimulation(validators, *totalRounds, *roundTimeMs, *proposeProb)

	exportTotalOrdering(validators, *totalRounds)
}

func randomFaultySet() map[int]bool {
	faulty := make(map[int]bool, f)
	for _, i := range rand.Perm(n)[:f] {
		faulty[i] = true
	}
	return faulty
}

func runSimulation(validators []*simulation.Validator, totalRounds, roundTimeMs int, proposeProb float64) {
	for round := 1; round <= totalRounds; round++ {
		faultySet := randomFaultySet()
		for i, v := range validators {
			v.Byzantine = faultySet[i]
			v.ByzantineHistory[round] = faultySet[i]

			if rand.Float64() < proposeProb {
				v.Propose(round)
			}
		}

		time.Sleep(time.Duration(roundTimeMs) * time.Millisecond)
	}
}

func createValidators(net *simulation.Network) []*simulation.Validator {
	validators := make([]*simulation.Validator, n)
	for i := range validators {
		validators[i] = simulation.NewValidator(i, f, false, net)
	}
	return validators
}

func startValidators(ctx context.Context, validators []*simulation.Validator) {
	for _, v := range validators {
		go v.Listen(ctx)
	}
}

func exportTotalOrdering(validators []*simulation.Validator, totalRounds int) {
	fmt.Println("\n--- DAG States ---")
	for _, v := range validators {
		v.PrintDAG()
	}

	// takes validators' Byzantine history and makes it into a data type that exporter uses to create its CSV
	var records []export.ByzantineRecord
	for _, v := range validators {
		for round := 1; round <= totalRounds; round++ {
			records = append(records, export.ByzantineRecord{
				Round:     round,
				Validator: v.ID,
				Byzantine: v.ByzantineHistory[round],
			})
		}
	}

	fmt.Println("\n--- Total Order (V0) ---")
	order := simulation.TotalOrder(validators[0].GetDAG())
	orderStrings := make([]string, len(order))
	for i, id := range order {
		orderStrings[i] = string(id)
	}

	// export to CSV
	blocks := validators[0].ExportDAG()
	export.WriteEdgesCSV(blocks, "edges.csv")
	export.WriteOrderCSV(orderStrings, "order.csv")
	export.WriteByzantineCSV(records, "byzantine.csv")
	fmt.Println("\nExported to edges.csv, order.csv, and byzantine.csv")

}
