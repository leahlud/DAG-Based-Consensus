package simulation

// Message types

type MsgType int

const (
	MsgProposal MsgType = iota
	MsgVote
	MsgCertificate
)

type Message struct {
	Type    MsgType
	From    int
	Payload any
}

// Network types

type Network struct {
	validators []*Validator
}

func NewNetwork() *Network {
	return &Network{}
}

func (n *Network) Register(validators []*Validator) {
	n.validators = validators
}

func (n *Network) Broadcast(from int, msg Message) {
	for _, v := range n.validators {
		if v.ID != from {
			v.Inbox <- msg
		}
	}
}
