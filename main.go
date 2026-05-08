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
	numValidators := flag.Int("n", 4, "number of validators")
	totalRounds := flag.Int("rounds", 10, "number of rounds to simulate")
	roundTimeMs := flag.Int("delay", 100, "round duration in ms")
	proposeProb := flag.Float64("p", 1.0, "probability a validator proposes in a round")
	flag.Parse()

	// compute number of validators and initialize them
	f = *numByzantine
	n = *numValidators
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
	printThroughputComparison(validators, *totalRounds, *roundTimeMs)
}

func randomFaultySet() map[int]bool {
	faulty := make(map[int]bool, f)
	for _, i := range rand.Perm(n)[:f] {
		faulty[i] = true
	}
	return faulty
}

func runSimulation(validators []*simulation.Validator, totalRounds, roundTimeMs int, proposeProb float64) int {
    proposed := 0
    for round := 1; round <= totalRounds; round++ {
        faultySet := randomFaultySet()
        for i, v := range validators {
            v.Byzantine = faultySet[i]
            v.ByzantineHistory[round] = faultySet[i]

            if rand.Float64() < proposeProb {
                v.Propose(round)
                proposed++
            }
        }
        time.Sleep(time.Duration(roundTimeMs) * time.Millisecond)
    }
    return proposed
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

	// Collect rejected blocks from all validators (union across validators)
	rejectedSet := make(map[string]bool)
	for _, v := range validators {
		for id := range v.RejectedBlocks {
			rejectedSet[string(id)] = true
		}
	}
	rejectedList := make([]string, 0, len(rejectedSet))
	for id := range rejectedSet {
		rejectedList = append(rejectedList, id)
	}

	fmt.Println("\n--- Total Order (V0) ---")
	order := simulation.TotalOrder(validators[0].GetDAG())
	orderStrings := make([]string, len(order))
	for i, id := range order {
		orderStrings[i] = string(id)
	}

	// export to CSV
	blocks := validators[0].ExportDAG()
	export.WriteEdgesCSV(blocks, "csvs/edges.csv")
	export.WriteOrderCSV(orderStrings, "csvs/order.csv")
	export.WriteByzantineCSV(records, "csvs/byzantine.csv")
	export.WriteRejectedCSV(rejectedList, "csvs/rejected.csv")
	fmt.Println("\nExported to edges.csv, order.csv, byzantine.csv, and rejected.csv")

}


func printThroughputComparison(validators []*simulation.Validator, totalRounds, roundTimeMs int) {
    totalTimeS := float64(totalRounds*roundTimeMs) / 1000.0

    // count actually certified blocks from V0's DAG
    dag := validators[0].GetDAG()
    certified := 0
    for round := 1; round <= totalRounds; round++ {
        certified += dag.CountAtRound(round)
    }

    // pBFT: one leader per round, so at most totalRounds blocks committed
    // scale by the same certification rate to be fair
    certRate := float64(certified) / float64(totalRounds*len(validators))
    pbftCommitted := int(float64(totalRounds) * certRate)

    dagThroughput  := float64(certified) / totalTimeS
    pbftThroughput := float64(pbftCommitted) / totalTimeS

    fmt.Println("\n--- Throughput Comparison ---")
    fmt.Printf("Elapsed time:     %.2fs (%d rounds x %d ms)\n", totalTimeS, totalRounds, roundTimeMs)
    fmt.Printf("DAG throughput:   %.0f blocks/s (%d blocks certified)\n", dagThroughput, certified)
    fmt.Printf("pBFT throughput:  %.0f blocks/s (%d blocks, one leader/round)\n", pbftThroughput, pbftCommitted)
    fmt.Printf("DAG speedup:      %.1fx\n", dagThroughput/pbftThroughput)
}
