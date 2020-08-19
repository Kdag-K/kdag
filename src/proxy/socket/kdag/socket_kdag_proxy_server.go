package kdag

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/Kdag-K/kdag/src/hashgraph"
	"github.com/Kdag-K/kdag/src/node/state"
	"github.com/Kdag-K/kdag/src/proxy"
	"github.com/sirupsen/logrus"
)

// SocketKdagProxyServer is the server component of the KdagProxy which
// responds to RPC requests from the client component of the AppProxy
type SocketKdagProxyServer struct {
	netListener *net.Listener
	rpcServer   *rpc.Server
	handler     proxy.ProxyHandler
	timeout     time.Duration
	logger      *logrus.Entry
}

// NewSocketKdagProxyServer creates a new SocketKdagProxyServer
func NewSocketKdagProxyServer(
	bindAddress string,
	handler proxy.ProxyHandler,
	timeout time.Duration,
	logger *logrus.Entry,
) (*SocketKdagProxyServer, error) {

	server := &SocketKdagProxyServer{
		handler: handler,
		timeout: timeout,
		logger:  logger,
	}

	if err := server.register(bindAddress); err != nil {
		return nil, err
	}

	return server, nil
}

func (p *SocketKdagProxyServer) register(bindAddress string) error {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("State", p)

	p.rpcServer = rpcServer

	l, err := net.Listen("tcp", bindAddress)

	if err != nil {
		return err
	}

	p.netListener = &l

	return nil
}

func (p *SocketKdagProxyServer) listen() error {
	for {
		conn, err := (*p.netListener).Accept()

		if err != nil {
			return err
		}

		go (*p.rpcServer).ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

// CommitBlock implements the AppProxy interface
func (p *SocketKdagProxyServer) CommitBlock(block hashgraph.Block, response *proxy.CommitResponse) (err error) {
	*response, err = p.handler.CommitHandler(block)

	p.logger.WithFields(logrus.Fields{
		"block":    block.Index(),
		"response": response,
		"err":      err,
	}).Debug("KdagProxyServer.CommitBlock")

	return
}

// GetSnapshot implements the AppProxy interface
func (p *SocketKdagProxyServer) GetSnapshot(blockIndex int, snapshot *[]byte) (err error) {
	*snapshot, err = p.handler.SnapshotHandler(blockIndex)

	p.logger.WithFields(logrus.Fields{
		"block":    blockIndex,
		"snapshot": snapshot,
		"err":      err,
	}).Debug("KdagProxyServer.GetSnapshot")

	return
}

// Restore implements the AppProxy interface
func (p *SocketKdagProxyServer) Restore(snapshot []byte, stateHash *[]byte) (err error) {
	*stateHash, err = p.handler.RestoreHandler(snapshot)

	p.logger.WithFields(logrus.Fields{
		"state_hash": stateHash,
		"err":        err,
	}).Debug("KdagProxyServer.Restore")

	return
}

// OnStateChanged implements the AppProxy interface
func (p *SocketKdagProxyServer) OnStateChanged(state state.State, obj *struct{}) (err error) {
	err = p.handler.StateChangeHandler(state)

	p.logger.WithFields(logrus.Fields{
		"state": state.String(),
		"err":   err,
	}).Debug("KdagProxyServer.OnStateChanged")

	return
}
