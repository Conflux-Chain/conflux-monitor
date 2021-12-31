package cmd

import (
	"fmt"
	"time"

	"github.com/Conflux-Chain/conflux-monitor/worker"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/inancgumus/screen"
	"github.com/spf13/cobra"
)

var epochGapCmd = &cobra.Command{
	Use:   "epochGap",
	Short: "Epoch gaps between latest_confirmed and latest_state",
	Run:   monitorEpochGap,
}

func init() {
	rootCmd.AddCommand(epochGapCmd)
}

func monitorEpochGap(*cobra.Command, []string) {
	cfx := sdk.MustNewClient(nodeURL)
	defer cfx.Close()

	go worker.MonitorEpochGaps(cfx, time.Second)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	screen.Clear()

	for range ticker.C {
		screen.MoveTopLeft()

		ss := worker.TimerEpochGaps.Snapshot()
		ps := ss.Percentiles([]float64{0.5, 0.9, 0.99})

		fmt.Printf("%-6s%10v\n", "min", ss.Min())
		fmt.Printf("%-6s%10.0f\n", "mean", ss.Mean())
		fmt.Printf("%-6s%10.0f\n", "p50", ps[0])
		fmt.Printf("%-6s%10.0f\n", "p90", ps[1])
		fmt.Printf("%-6s%10.0f\n", "p99", ps[2])
		fmt.Printf("%-6s%10v\n", "max", ss.Max())
	}
}
