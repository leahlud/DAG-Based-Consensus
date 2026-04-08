package simulation

import (
	"fmt"
	"sync"
)

// DAG stores certified blocks organized by round and validator.
// Each validator owns its own DAG instance, so two validators
// may have different views at any given moment.
type DAG struct {
	mu     sync.RWMutex
	blocks map[int]map[int]*Certificate // round -> authorID -> cert
}

func NewDAG() *DAG {
	return &DAG{
		blocks: make(map[int]map[int]*Certificate),
	}
}

// Add inserts a certificate into the DAG
func (d *DAG) Add(cert *Certificate) {
	d.mu.Lock()
	defer d.mu.Unlock()

	r := cert.Block.Round
	if d.blocks[r] == nil {
		d.blocks[r] = make(map[int]*Certificate)
	}
	d.blocks[r][cert.Block.Author] = cert
}

// Contains returns true if a certificate for this round and author exists
func (d *DAG) Contains(round, author int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.blocks[round] != nil && d.blocks[round][author] != nil
}

// GetCertificate returns the certificate for a given round and author.
// The second return value is false if no certificate exists.
func (d *DAG) GetCertificate(round, author int) (*Certificate, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.blocks[round] == nil {
		return nil, false
	}
	cert, ok := d.blocks[round][author]
	return cert, ok
}

// CountRounds returns the total number of rounds in the DAG
func (d *DAG) CountRounds() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.blocks)
}

// CountAtRound returns how many certified blocks exist at a given round
func (d *DAG) CountAtRound(round int) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.blocks[round])
}

// GetCertifiedAtRound returns all certified blocks at a given round.
// Used by validators to collect parent references when proposing.
func (d *DAG) GetCertifiedAtRound(round int) []*Certificate {
	d.mu.RLock()
	defer d.mu.RUnlock()

	certs := make([]*Certificate, 0, len(d.blocks[round]))
	for _, cert := range d.blocks[round] {
		certs = append(certs, cert)
	}
	return certs
}

// Print displays the DAG state for a given validator, organized by round
func (d *DAG) Print(validatorID int) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	fmt.Printf("  [V%d] DAG:\n", validatorID)
	for round := 1; round <= len(d.blocks); round++ {
		fmt.Printf("    Round %d: ", round)
		for author, cert := range d.blocks[round] {
			fmt.Printf("r%d-v%d(%d votes) ", round, author, cert.Votes)
		}
		fmt.Println()
	}
}
