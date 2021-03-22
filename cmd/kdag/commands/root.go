package commands

import (
	"github.com/spf13/cobra"
)

var (
	_config = NewDefaultCLIConf()
)

// RootCmd is the root command for Kdag
var RootCmd = &cobra.Command{
	Use:              "kdag",
	Short:            "kdag consensus",
	TraverseChildren: true,
}
