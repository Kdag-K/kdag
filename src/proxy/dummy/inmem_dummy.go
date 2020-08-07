package dummy

import (
	"github.com/sirupsen/logrus"

	"github.com/Kdag-K/kdag/src/proxy/inmem"
)

// InmemDummyClient is an in-memory implementation of the dummy app. It actually
// imlplements the AppGateway interface, and can be passed in the Babble
// constructor directly
type InmemDummyClient struct {
	*inmem.InmemProxy
	state  *State
	logger *logrus.Entry
}

//NewInmemDummyClient instantiates an InemDummyClient
func NewInmemDummyClient(logger *logrus.Entry) *InmemDummyClient {
	state := NewState(logger)

	proxy := inmem.NewInmemProxy(state, logger)

	client := &InmemDummyClient{
		InmemProxy: proxy,
		state:      state,
		logger:     logger,
	}

	return client
}

//SubmitTx sends a transaction to the Babble node via the InmemProxy
func (c *InmemDummyClient) SubmitTx(tx []byte) {
	c.InmemProxy.SubmitTx(tx)
}

//GetCommittedTransactions returns the state's list of transactions
func (c *InmemDummyClient) GetCommittedTransactions() [][]byte {
	return c.state.GetCommittedTransactions()
}
