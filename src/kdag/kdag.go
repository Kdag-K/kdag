package kdag

import (
	"os"
	_ "time"
	
	"github.com/sirupsen/logrus"
	
	"github.com/Kdag-K/kdag/src/config"
	"github.com/Kdag-K/kdag/src/crypto/keys"
	h "github.com/Kdag-K/kdag/src/hashgraph"
	"github.com/Kdag-K/kdag/src/net"
	"github.com/Kdag-K/kdag/src/net/signal/wamp"
	"github.com/Kdag-K/kdag/src/node"
	"github.com/Kdag-K/kdag/src/peers"
	"github.com/Kdag-K/kdag/src/service"
)

// Kdag is a struct containing the key parts of a kdag node
type Kdag struct {
	Config       *config.Config
	Node         *node.Node
	Transport    net.Transport
	Store        h.Store
	Peers        *peers.PeerSet
	GenesisPeers *peers.PeerSet
	Service      *service.Service
	logger       *logrus.Entry
}

// NewKdag is a factory method to produce
// a Kdag instance.
func NewKdag(c *config.Config) *Kdag {
	engine := &Kdag{
		Config: c,
		logger: c.Logger(),
	}

	return engine
}

// Init initialises the kdag engine
func (b *Kdag) Init() error {

	b.logger.Debug("validateConfig")
	if err := b.validateConfig(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() validateConfig")
	}

	b.logger.Debug("initKey")
	if err := b.initKey(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initKey")
		return err
	}
	b.logger.Debug("initPeers")
	if err := b.initPeers(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initPeers")
		return err
	}

	b.logger.Debug("initStore")
	if err := b.initStore(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initStore")
		return err
	}

	b.logger.Debug("initTransport")
	if err := b.initTransport(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initTransport")
		return err
	}

	b.logger.Debug("initKey")
	if err := b.initKey(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initKey")
		return err
	}

	b.logger.Debug("initNode")
	if err := b.initNode(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initNode")
		return err
	}

	b.logger.Debug("initService")
	if err := b.initService(); err != nil {
		b.logger.WithError(err).Error("kdag.go:Init() initService")
		return err
	}

	return nil
}

// Run starts the Kdag Node running
func (b *Kdag) Run() {
	if b.Service != nil && b.Config.ServiceAddr != "" {
		go b.Service.Serve()
	}

	b.Node.Run(true)
}

func (b *Kdag) validateConfig() error {
	// If --datadir was explicitely set, but not --db, the following line will
	// update the default database dir to be inside the new datadir
	b.Config.SetDataDir(b.Config.DataDir)

	logFields := logrus.Fields{
		"kdag.DataDir":          b.Config.DataDir,
		"kdag.ServiceAddr":      b.Config.ServiceAddr,
		"kdag.NoService":        b.Config.NoService,
		"kdag.MaxPool":          b.Config.MaxPool,
		"kdag.LogLevel":         b.Config.LogLevel,
		"kdag.Moniker":          b.Config.Moniker,
		"kdag.HeartbeatTimeout": b.Config.HeartbeatTimeout,
		"kdag.TCPTimeout":       b.Config.TCPTimeout,
		"kdag.JoinTimeout":      b.Config.JoinTimeout,
		"kdag.CacheSize":        b.Config.CacheSize,
		"kdag.SyncLimit":        b.Config.SyncLimit,
		"kdag.EnableFastSync":   b.Config.EnableFastSync,
		"kdag.MaintenanceMode":  b.Config.MaintenanceMode,
		"kdag.SuspendLimit":     b.Config.SuspendLimit,
	}

	// WebRTC requires signaling and ICE servers
	if b.Config.WebRTC {
		logFields["kdag.WebRTC"] = b.Config.WebRTC
		logFields["kdag.SignalAddr"] = b.Config.SignalAddr
		logFields["kdag.SignalRealm"] = b.Config.SignalRealm
		logFields["kdag.SignalSkipVerify"] = b.Config.SignalSkipVerify
		logFields["kdag.ICEAddress"] = b.Config.ICEAddress
		logFields["kdag.ICEUsername"] = b.Config.ICEUsername
	} else {
		logFields["kdag.BindAddr"] = b.Config.BindAddr
		logFields["kdag.AdvertiseAddr"] = b.Config.AdvertiseAddr
	}
	// Maintenance-mode only works with bootstrap
	if b.Config.MaintenanceMode {
		b.logger.Debug("Config maintenance-mode => bootstrap")
		b.Config.Bootstrap = true
	}

	// Bootstrap only works with store
	if b.Config.Bootstrap {
		b.logger.Debug("Config boostrap => store")
		b.Config.Store = true
	}

	if b.Config.Store {
		logFields["kdag.Store"] = b.Config.Store
		logFields["kdag.DatabaseDir"] = b.Config.DatabaseDir
		logFields["kdag.Bootstrap"] = b.Config.Bootstrap
	}

	// SlowHeartbeat cannot be less than Heartbeat
	if b.Config.SlowHeartbeatTimeout < b.Config.HeartbeatTimeout {
		b.logger.Debugf("SlowHeartbeatTimeout (%v) cannot be less than Heartbeat (%v)",
			b.Config.SlowHeartbeatTimeout,
			b.Config.HeartbeatTimeout)
		b.Config.SlowHeartbeatTimeout = b.Config.HeartbeatTimeout
	}
	logFields["kdag.SlowHeartbeatTimeout"] = b.Config.SlowHeartbeatTimeout

	b.logger.WithFields(logFields).Debug("Config")

	return nil
}

func (b *Kdag) initTransport() error {
	if b.Config.MaintenanceMode {
		return nil
	}

	if b.Config.WebRTC {
		signal, err := wamp.NewClient(
			b.Config.SignalAddr,
			b.Config.SignalRealm,
			keys.PublicKeyHex(&b.Config.Key.PublicKey),
			b.Config.CertFile(),
			b.Config.SignalSkipVerify,
			b.Config.TCPTimeout,
			b.Config.Logger().WithField("component", "webrtc-signal"),
		)

		if err != nil {
			return err
		}

		webRTCTransport, err := net.NewWebRTCTransport(
			signal,
			b.Config.ICEServers(),
			b.Config.MaxPool,
			b.Config.TCPTimeout,
			b.Config.JoinTimeout,
			b.Config.Logger().WithField("component", "webrtc-transport"),
		)

		if err != nil {
			return err
		}

		b.Transport = webRTCTransport
	} else {
		tcpTransport, err := net.NewTCPTransport(
		b.Config.BindAddr,
		b.Config.AdvertiseAddr,
		b.Config.MaxPool,
		b.Config.TCPTimeout,
		b.Config.JoinTimeout,
		b.Config.Logger(),
	)

	if err != nil {
		return err
	}

		b.Transport = tcpTransport
 }

	return nil
	}

	// peers.json
func (b *Kdag) initPeers() error {
	peerStore := peers.NewJSONPeerSet(b.Config.DataDir, true)

	participants, err := peerStore.PeerSet()
	if err != nil {
		return err
	}

	b.Peers = participants
	b.logger.Debug("Loaded Peers")

	// Set Genesis Peer Set from peers.genesis.json
	genesisPeerStore := peers.NewJSONPeerSet(b.Config.DataDir, false)

	genesisParticipants, err := genesisPeerStore.PeerSet()
	if err != nil { // If there is any error, the current peer set is used as the genesis peer set
		b.logger.Debugf("could not read peers.genesis.json: %v", err)
		b.GenesisPeers = participants
	} else {
		b.GenesisPeers = genesisParticipants
	}

	return nil
}

func (b *Kdag) initStore() error {
	if !b.Config.Store {
		b.logger.Debug("Creating InmemStore")
		b.Store = h.NewInmemStore(b.Config.CacheSize)
	} else {
		dbPath := b.Config.DatabaseDir

		b.logger.WithField("path", dbPath).Debug("Creating BadgerStore")

		if !b.Config.Bootstrap {
			b.logger.Debug("No Bootstrap")

			backup := backupFileName(dbPath)

			err := os.Rename(dbPath, backup)

			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
				b.logger.Debug("Nothing to backup")
			} else {
				b.logger.WithField("path", backup).Debug("Created backup")
			}
		}

		b.logger.WithField("path", dbPath).Debug("Opening BadgerStore")

		dbStore, err := h.NewBadgerStore(
			b.Config.CacheSize,
			dbPath,
			b.Config.MaintenanceMode,
			b.logger)
		if err != nil {
			return err
		}

		b.Store = dbStore
	}

	return nil
}

func (b *Kdag) initKey() error {
	if b.Config.Key == nil {
		simpleKeyfile := keys.NewSimpleKeyfile(b.Config.Keyfile())

		privKey, err := simpleKeyfile.ReadKey()
		if err != nil {
			b.logger.Errorf("Error reading private key from file: %v", err)
		}

		b.Config.Key = privKey
	}
	return nil
}

func (b *Kdag) initNode() error {

	validator := node.NewValidator(b.Config.Key, b.Config.Moniker)

	p, ok := b.Peers.ByID[validator.ID()]
	if ok {
		if p.Moniker != validator.Moniker {
			b.logger.WithFields(logrus.Fields{
				"json_moniker": p.Moniker,
				"cli_moniker":  validator.Moniker,
			}).Debugf("Using moniker from peers.json file")
			validator.Moniker = p.Moniker
		}
	}

	b.Config.Logger().WithFields(logrus.Fields{
		"genesis_peers": len(b.GenesisPeers.Peers),
		"peers":         len(b.Peers.Peers),
		"id":            validator.ID(),
		"moniker":       validator.Moniker,
	}).Debug("PARTICIPANTS")

	b.Node = node.NewNode(
		b.Config,
		validator,
		b.Peers,
		b.GenesisPeers,
		b.Store,
		b.Transport,
		b.Config.Proxy,
	)

	return b.Node.Init()
}

func (b *Kdag) initService() error {
	if !b.Config.NoService {
		b.Service = service.NewService(b.Config.ServiceAddr, b.Node, b.Config.Logger())
	}
	return nil
}

// backupFileName implements the naming convention for database backups:
// badger_db--UTC--<created_at UTC ISO8601>
func backupFileName(base string) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("%s--UTC--%s", base, toISO8601(ts))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}
