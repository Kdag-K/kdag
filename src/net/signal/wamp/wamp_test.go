package wamp

import (
	"strings"
	"testing"
	"time"

	"github.com/Kdag-K/kdag/src/common"
	"github.com/Kdag-K/kdag/src/config"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const certFile = "test_data/cert.pem"
const keyFile = "test_data/key.pem"
const signalTimeout = 5 * time.Second

func TestWampSelfSigned(t *testing.T) {
	url := "localhost:2443"
	realm := config.DefaultSignalRealm
	certFile := "test_data/cert.pem"
	keyFile := "test_data/key.pem"

	server, err := NewServer(url,
		realm,
		certFile,
		keyFile,
		common.NewTestLogger(t, logrus.DebugLevel).WithField("component", "signal-server"))
	
	require.NoError(t, err)

	go server.Run()
	defer server.Shutdown()
	// Allow the server some time to run otherwise we get some connection
	// refused errors
	time.Sleep(time.Second)

	callee, err := NewClient(url,
		realm,
		"callee",
		certFile,
		false,
		signalTimeout,
		common.NewTestLogger(t, logrus.DebugLevel).WithField("component", "signal-client callee"))
	
	require.NoError(t, err)

	defer callee.Close()

	err := callee.Listen()
	require.NoError(t, err)

	caller, err := NewClient(
		url,
		realm,
		"caller",
		certFile,
		false,
		signalTimeout,
		common.NewTestLogger(t, logrus.DebugLevel).WithField("component", "signal-client caller"))
	
	require.NoError(t, err)

	defer caller.Close()

	// We expect the call to reach the callee and to generate an
	// ErrProcessingOffer error because the SDP is empty. We are only trying to
	// test that the RPC call is relayed and that the handler on the receiving
	// end is called
	_, err = caller.Offer("callee", webrtc.SessionDescription{})
	require.NoError(t, err)
	if err == nil || !strings.Contains(err.Error(), ErrProcessingOffer) {
		t.Fatal("Should have receveived an ErrProcessingOffer")

	}
}