package simulation

import (
	"fmt"
	"strconv"
	"strings"
)

// A BlockID is a human readable unique identifier for a block
// (i.e. "r2-v1" --> round 2, validator 1)
type BlockID string

// A Block is a proposal made by a validator for a given round
type Block struct {
	Round   int
	Author  int
	TxCount int
	Parents []BlockID // references to certified blocks from round-1
}

// A Certificate represents a block that has received 2f+1 votes
// and is therefore guaranteed to be available across the network
type Certificate struct {
	Block Block
	Votes int
}

// GetID returns the deterministic BlockID for a block (i.e. "r2-v1")
func (b Block) GetID() BlockID {
	return BlockID(fmt.Sprintf("r%d-v%d", b.Round, b.Author))
}

func parseBlockID(id BlockID) (round int, author int) {
	parts := strings.Split(string(id), "-")
	if len(parts) != 2 {
		return 0, 0
	}
	round, _ = strconv.Atoi(strings.TrimPrefix(parts[0], "r"))
	author, _ = strconv.Atoi(strings.TrimPrefix(parts[1], "v"))
	return round, author
}
