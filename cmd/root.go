package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	nodeURL string

	rootCmd = &cobra.Command{
		Use:   "conflux-monitor",
		Short: "Utilities to monitor Conflux blockchain",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&nodeURL, "url", "http://main.confluxrpc.com", "Conflux fullnode RPC URL")
}

func exitIfErr(err error, msg string, args ...interface{}) {
	if err != nil {
		logrus.WithError(err).Fatalf(msg, args...)
	}
}

// Execute is the command line entrypoint.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("Failed to execute command")
	}
}
