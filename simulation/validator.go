package simulation

import (
	"context"
	"dag-based-consensus/export"
)

type Validator struct {
	ID         int
	F          int
	Byzantine  bool
	Inbox      chan Message
	dag        *DAG
	net        *Network
	votes      map[BlockID]int   // tracks vote count per block for certification
	blockCache map[BlockID]Block // cache proposals so parents are preserved
	sequencer  *Sequencer
}

func NewValidator(id, f int, isByzantine bool, net *Network) *Validator {
	v := &Validator{
		ID:         id,
		F:          f,
		Byzantine:  isByzantine,
		Inbox:      make(chan Message, 100),
		dag:        NewDAG(),
		net:        net,
		votes:      make(map[BlockID]int),
		blockCache: make(map[BlockID]Block),
	}
	v.sequencer = NewSequencer(f, func(id BlockID) {
		_ = id
	})

	return v
}

// GetDAG returns the validator's local DAG for external inspection
func (v *Validator) GetDAG() *DAG {
	return v.dag
}

// PrintDAG prints the validator's local DAG state
func (v *Validator) PrintDAG() {
	v.dag.Print(v.ID)
}

// ExportDAG converts the validator's DAG to a JSON serializable format
func (v *Validator) ExportDAG() []export.ExportBlock {
	blocks := []export.ExportBlock{}
	for round := 1; round <= v.dag.CountRounds(); round++ {
		for _, cert := range v.dag.GetCertifiedAtRound(round) {
			parents := make([]string, len(cert.Block.Parents))
			for i, p := range cert.Block.Parents {
				parents[i] = string(p)
			}
			blocks = append(blocks, export.ExportBlock{
				ID:      string(cert.Block.GetID()),
				Round:   cert.Block.Round,
				Author:  cert.Block.Author,
				Parents: parents,
				Votes:   cert.Votes,
			})
		}
	}
	return blocks
}

// Listen runs the validator's message loop, handling incoming messages
// until the context is cancelled
func (v *Validator) Listen(ctx context.Context) {
	for {
		select {
		case msg := <-v.Inbox:
			v.Handle(msg)
		case <-ctx.Done():
			return
		}
	}
}

// Propose creates a block for the given round and broadcasts it to all peers.
// Byzantine validators are silent and do not propose.
func (v *Validator) Propose(round int) {
	if v.Byzantine {
		return
	}

	block := Block{
		Round:   round,
		Author:  v.ID,
		TxCount: 10,
		Parents: v.collectParents(round),
	}

	v.net.Broadcast(v.ID, Message{
		Type:    MsgProposal,
		From:    v.ID,
		Payload: block,
	})
}

// Handle processes an incoming message based on its type
func (v *Validator) Handle(msg Message) {
	switch msg.Type {
	case MsgProposal:
		block := msg.Payload.(Block)
		v.blockCache[block.GetID()] = block

		// send vote back to proposer
		v.net.Send(v.ID, block.Author, Message{
			Type:    MsgVote,
			From:    v.ID,
			Payload: block.GetID(),
		})

	case MsgVote:
		blockID := msg.Payload.(BlockID)
		v.votes[blockID]++

		// certify the block once 2f+1 votes are received
		if v.votes[blockID] == 2*v.F+1 {
			v.certify(blockID)
		}

	case MsgCertificate:
		// add the certified block to the local DAG
		cert := msg.Payload.(Certificate)
		v.dag.Add(&cert)
	}
}

// certify creates a certificate for a block that has received 2f+1 votes,
// adds it to the local DAG, and broadcasts it to all peers
func (v *Validator) certify(id BlockID) {
	block, ok := v.blockCache[id]
	if !ok {
		// fallback if we certified our own block (we never received it as a proposal)
		round, author := parseBlockID(id)
		block = Block{Round: round, Author: author, TxCount: 10, Parents: v.collectParents(round)}
	}

	cert := Certificate{
		Block: block,
		Votes: 2*v.F + 1,
	}

	v.dag.Add(&cert)
	v.sequencer.TryAdvance(v.dag)

	v.net.Broadcast(v.ID, Message{
		Type:    MsgCertificate,
		From:    v.ID,
		Payload: cert,
	})
}

// collectParents returns the BlockIDs of all certified blocks from the previous
// round to be used as parent references in a new block proposal
func (v *Validator) collectParents(round int) []BlockID {
	if round <= 1 {
		return []BlockID{}
	}
	certs := v.dag.GetCertifiedAtRound(round - 1)
	parents := make([]BlockID, len(certs))
	for i, cert := range certs {
		parents[i] = cert.Block.GetID()
	}
	return parents
}
