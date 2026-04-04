package simulation

import (
	"context"
	"fmt"
)

type Validator struct {
	ID        int
	F         int
	Byzantine bool
	Inbox     chan Message
	dag       *DAG
	net       *Network
	votes     map[BlockID]int
}

func NewValidator(id, f int, isByzantine bool, net *Network) *Validator {
	return &Validator{
		ID:        id,
		F:         f,
		Byzantine: isByzantine,
		Inbox:     make(chan Message, 100),
		dag:       NewDAG(),
		net:       net,
		votes:     make(map[BlockID]int),
	}
}

func (v *Validator) PrintDAG() {
	v.dag.Print(v.ID)
}

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

func (v *Validator) Propose(round int) {
	if v.Byzantine {
		fmt.Printf("  [V%d] (byzantine) silent this round\n", v.ID)
		return
	}

	block := Block{
		Round:   round,
		Author:  v.ID,
		TxCount: 10,
		Parents: v.collectParents(round),
	}

	fmt.Printf("  [V%d] proposing %s with %d parents\n", v.ID, block.GetID(), len(block.Parents))

	v.net.Broadcast(v.ID, Message{
		Type:    MsgProposal,
		From:    v.ID,
		Payload: block,
	})
}

func (v *Validator) Handle(msg Message) {
	switch msg.Type {
	case MsgProposal:
		block := msg.Payload.(Block)
		fmt.Printf("  [V%d] received proposal %s\n", v.ID, block.GetID())

		// send vote back to proposer
		v.net.Send(v.ID, block.Author, Message{
			Type:    MsgVote,
			From:    v.ID,
			Payload: block.GetID(),
		})

	case MsgVote:
		blockID := msg.Payload.(BlockID)
		v.votes[blockID]++
		fmt.Printf("  [V%d] received vote for %s (%d/%d)\n", v.ID, blockID, v.votes[blockID], 2*v.F+1)

		if v.votes[blockID] == 2*v.F+1 {
			v.certify(blockID)
		}

	case MsgCertificate:
		cert := msg.Payload.(Certificate)
		fmt.Printf("  [V%d] received certificate %s\n", v.ID, cert.Block.GetID())
		v.dag.Add(&cert)
	}
}

func (v *Validator) certify(id BlockID) {
	round, author := parseBlockID(id)
	cert := Certificate{
		Block: Block{Round: round, Author: author, TxCount: 10},
		Votes: 2*v.F + 1,
	}
	v.dag.Add(&cert)
	fmt.Printf("  [V%d] certified %s\n", v.ID, id)
	v.net.Broadcast(v.ID, Message{
		Type:    MsgCertificate,
		From:    v.ID,
		Payload: cert,
	})
}

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
