package simulation

// MsgType identifies the kind of message being sent between validators
type MsgType int

const (
	MsgProposal    MsgType = iota // a new block proposal from a validator
	MsgVote                       // a vote for a received proposal
	MsgCertificate                // a block that has received 2f+1 votes
)

// Message is the unit of communication between validators
type Message struct {
	Type    MsgType
	From    int
	Payload any
}

// Network simulates message passing between validators with direct
// channel message sending between goroutines.
type Network struct {
	validators []*Validator
}

func NewNetwork() *Network {
	return &Network{}
}

// Register wires the validators into the network so they can receive messages
func (n *Network) Register(validators []*Validator) {
	n.validators = validators
}

// Broadcast sends a message from one validator to all others
func (n *Network) Broadcast(from int, msg Message) {
	for _, v := range n.validators {
		if v.ID != from {
			v.Inbox <- msg
		}
	}
}

// Send delivers a message from one validator to a specific target
func (n *Network) Send(from, to int, msg Message) {
	n.validators[to].Inbox <- msg
}
