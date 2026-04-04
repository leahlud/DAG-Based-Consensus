package simulation

import "sort"

// TotalOrder derives a deterministic total order from a completed DAG.
// It sweeps through rounds in order and sorts blocks within each round
// by author ID as a tiebreaker, producing a consistent linear sequence.
func TotalOrder(dag *DAG) []BlockID {
	dag.mu.RLock()
	defer dag.mu.RUnlock()

	ordered := []BlockID{}

	for round := 1; round <= len(dag.blocks); round++ {
		// collect all certified blocks at this round
		certs := []*Certificate{}
		for _, cert := range dag.blocks[round] {
			certs = append(certs, cert)
		}

		// sort by author ID for determinism
		sort.Slice(certs, func(i, j int) bool {
			return certs[i].Block.Author < certs[j].Block.Author
		})

		for _, cert := range certs {
			ordered = append(ordered, cert.Block.GetID())
		}
	}

	return ordered
}
