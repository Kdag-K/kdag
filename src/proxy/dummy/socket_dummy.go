package dummy

import (
	"time"

	"github.com/sirupsen/logrus"

	socket "github.com/Kdag-K/kdag/src/proxy/socket/kdag"
)

// DummySocketClient is a socket implementation of the dummy app. Kdag and the
// app run in separate processes and communicate through TCP sockets using
// a SocketKdagProxy and a SocketAppProxy.
type DummySocketClient struct {
	state       *State
	babbleProxy *socket.SocketKdagProxy
	logger      *logrus.Entry
}

//NewDummySocketClient instantiates a DummySocketClient and starts the
//SocketKdagProxy
func NewDummySocketClient(clientAddr string, nodeAddr string, logger *logrus.Entry) (*DummySocketClient, error) {
	state := NewState(logger)

	babbleProxy, err := socket.NewSocketKdagProxy(nodeAddr, clientAddr, state, 1*time.Second, logger)

	if err != nil {
		return nil, err
	}

	client := &DummySocketClient{
		state:       state,
		babbleProxy: babbleProxy,
		logger:      logger,
	}

	return client, nil
}

//SubmitTx sends a transaction to Kdag via the SocketProxy
func (c *DummySocketClient) SubmitTx(tx []byte) error {
	return c.babbleProxy.SubmitTx(tx)
}
