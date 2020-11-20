package commands

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Kdag-K/kdag/src/kdag"
	aproxy "github.com/Kdag-K/kdag/src/proxy/socket/app"
)

//NewRunCmd returns the command that starts a Kdag node
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run node",
		PreRunE: bindFlagsLoadViper,
		RunE:    runKdag,
	}
	AddRunFlags(cmd)
	return cmd
}

/*******************************************************************************
* RUN
*******************************************************************************/

func runKdag(cmd *cobra.Command, args []string) error {

	_config.Kdag.Logger().WithFields(logrus.Fields{
		"ProxyAddr":  _config.ProxyAddr,
		"ClientAddr": _config.ClientAddr,
	}).Debug("Config Proxy")

	p, err := aproxy.NewSocketAppProxy(
		_config.ClientAddr,
		_config.ProxyAddr,
		_config.Kdag.HeartbeatTimeout,
		_config.Kdag.Logger(),
	)

	if err != nil {
		_config.Kdag.Logger().Error("Cannot initialize socket AppGateway:", err)
		return err
	}

	_config.Kdag.Proxy = p

	engine := kdag.NewKdag(&_config.Kdag)

	if err := engine.Init(); err != nil {
		_config.Kdag.Logger().Error("Cannot initialize engine:", err)
		return err
	}

	engine.Run()

	return nil
}

/*******************************************************************************
* CONFIG
*******************************************************************************/

func initConfig(cmd *cobra.Command) error {
	return viper.BindPFlag(_config.Kdag.BindAddr, cmd.PersistentFlags().Lookup(_config.Kdag.BindAdd))
}


//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {

	cmd.Flags().String("datadir", _config.Kdag.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log", _config.Kdag.LogLevel, "debug, info, warn, error, fatal, panic")
	cmd.Flags().String("moniker", _config.Kdag.Moniker, "Optional name")
	cmd.Flags().BoolP("maintenance-mode", "R", _config.Kdag.MaintenanceMode, "Start Kdag in a suspended (non-gossipping) state")

	// Network
	cmd.Flags().StringP("listen", "l", _config.Kdag.BindAddr, "Listen IP:Port for kdag node")
	cmd.Flags().StringP("advertise", "a", _config.Kdag.AdvertiseAddr, "Advertise IP:Port for kdag node")
	cmd.Flags().DurationP("timeout", "t", _config.Kdag.TCPTimeout, "TCP Timeout")
	cmd.Flags().DurationP("join-timeout", "j", _config.Kdag.JoinTimeout, "Join Timeout")
	cmd.Flags().Int("max-pool", _config.Kdag.MaxPool, "Connection pool size max")
	
        // WebRTC
	cmd.Flags().Bool("webrtc", _config.Kdag.WebRTC, "Use WebRTC transport")
	cmd.Flags().String("signal-addr", _config.Kdag.SignalAddr, "IP:Port of WebRTC signaling server")
	cmd.Flags().Bool("signal-skip-verify", _config.Kdag.SignalSkipVerify, "(Insecure) Accept any certificate presented by the signal server")
	cmd.Flags().String("ice-addr", _config.Kdag.ICEAddress, "URI of a server providing ICE services such as STUN and TURN")
	cmd.Flags().String("ice-username", _config.Kdag.ICEUsername, "Username to authenticate to the ICE server")
	cmd.Flags().String("ice-password", _config.Kdag.ICEPassword, "Password to authenticate to the ICE server")

	// Proxy
	cmd.Flags().StringP("proxy-listen", "p", _config.ProxyAddr, "Listen IP:Port for kdag proxy")
	cmd.Flags().StringP("client-connect", "c", _config.ClientAddr, "IP:Port to connect to client")

	// Service
	cmd.Flags().Bool("no-service", _config.Kdag.NoService, "Disable HTTP service")
	cmd.Flags().StringP("service-listen", "s", _config.Kdag.ServiceAddr, "Listen IP:Port for HTTP service")

	// Store
	cmd.Flags().Bool("store", _config.Kdag.Store, "Use badgerDB instead of in-mem DB")
	cmd.Flags().String("db", _config.Kdag.DatabaseDir, "Dabatabase directory")
	cmd.Flags().Bool("bootstrap", _config.Kdag.Bootstrap, "Load from database")
	cmd.Flags().Int("cache-size", _config.Kdag.CacheSize, "Number of items in LRU caches")

	// Node configuration
	cmd.Flags().Duration("heartbeat", _config.Kdag.HeartbeatTimeout, "Timer frequency when there is something to gossip about")
	cmd.Flags().Duration("slow-heartbeat", _config.Kdag.SlowHeartbeatTimeout, "Timer frequency when there is nothing to gossip about")
	cmd.Flags().Int("sync-limit", _config.Kdag.SyncLimit, "Max number of events for sync")
	cmd.Flags().Bool("fast-sync", _config.Kdag.EnableFastSync, "Enable FastSync")
	cmd.Flags().Int("suspend-limit", _config.Kdag.SuspendLimit, "Limit of undetermined events before entering suspended state")
}

// Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, args []string) error {
	// Register flags with viper. Include flags from this command and all other
	// persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// first unmarshal to read from CLI flags
	if err := viper.Unmarshal(_config); err != nil {
		return err
	}

	// look for config file in [datadir]/kdag.toml (.json, .yaml also work)
	viper.SetConfigName("kdag")               // name of config file (without extension)
	viper.AddConfigPath(_config.Kdag.DataDir) // search root directory

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_config.Kdag.Logger().Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		_config.Kdag.Logger().Debugf("No config file found in: %s", filepath.Join(_config.Kdag.DataDir, "kdag.toml"))
	} else {
		return err
	}

	// second unmarshal to read from config file
	return viper.Unmarshal(_config)
}

func logLevel(l string) logrus.Level {
	switch l {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}
