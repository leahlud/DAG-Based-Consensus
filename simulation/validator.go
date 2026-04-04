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
}

func NewValidator(id, f int, isByzantine bool, net *Network) *Validator {
	return &Validator{
		ID:        id,
		F:         f,
		Byzantine: isByzantine,
		Inbox:     make(chan Message, 100),
		dag:       NewDAG(),
		net:       net,
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
		Parents: []BlockID{}, // todo: collect from dag
	}

	fmt.Printf("  [V%d] proposing %s\n", v.ID, block.GetID())

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

	case MsgVote:
		fmt.Printf("  [V%d] received vote from V%d\n", v.ID, msg.From)

	case MsgCertificate:
		cert := msg.Payload.(Certificate)
		fmt.Printf("  [V%d] received certificate %s\n", v.ID, cert.Block.GetID())
	}
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
