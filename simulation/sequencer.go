package simulation

import "sort"

// Sequencer tracks which rounds/blocks have been started and finalized
type Sequencer struct {
	nextRound int
	quorum    int // 2f+1
	OnBlock   func(id BlockID)
}

func NewSequencer(f int, onBlock func(id BlockID)) *Sequencer {
	return &Sequencer{
		nextRound: 1,
		quorum:    2*f + 1,
		OnBlock:   onBlock,
	}
}

// TryAdvance checks whether the next expected round is complete in the DAG
func (s *Sequencer) TryAdvance(dag *DAG) {
	// keeps advancing until it finds an incomplete round.
	for {
		certs := dag.GetCertifiedAtRound(s.nextRound)
		if len(certs) < s.quorum {
			return // round not complete yet
		}

		// sorts so it is deterministci
		sort.Slice(certs, func(i, j int) bool {
			return certs[i].Block.Author < certs[j].Block.Author
		})

		for _, cert := range certs {
			s.OnBlock(cert.Block.GetID())
		}
		s.nextRound++
	}
}

// OrderUpToRound returns all certified blocks from round 1 through maxRound sorted by validator ID.
func OrderUpToRound(dag *DAG, maxRound int) []BlockID {
	dag.mu.RLock()
	defer dag.mu.RUnlock()

	var ordered []BlockID
	for round := 1; round <= maxRound; round++ {
		certs := make([]*Certificate, 0, len(dag.blocks[round]))
		for _, cert := range dag.blocks[round] {
			certs = append(certs, cert)
		}
		sort.Slice(certs, func(i, j int) bool {
			return certs[i].Block.Author < certs[j].Block.Author
		})
		for _, cert := range certs {
			ordered = append(ordered, cert.Block.GetID())
		}
	}
	return ordered
}
