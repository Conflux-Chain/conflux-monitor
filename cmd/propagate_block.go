package cmd

import (
	"fmt"
	"time"

	"github.com/Conflux-Chain/conflux-monitor/epoch"
	"github.com/Conflux-Chain/conflux-monitor/worker"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/inancgumus/screen"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var blockPropagationCmd = &cobra.Command{
	Use:   "blockPropagation",
	Short: "Statistic the block propagation latency",
	Run:   monitorBlockPropagation,
}

func init() {
	rootCmd.AddCommand(blockPropagationCmd)
}

func monitorBlockPropagation(*cobra.Command, []string) {
	cfx := sdk.MustNewClient(nodeURL)
	defer cfx.Close()

	epochCh := make(chan uint64, 1024)
	defer close(epochCh)

	// IMPORTANT: statistic against the latest state, so as to use time.Now() as block execution time.
	epochPoller, err := epoch.NewPoller(cfx)
	exitIfErr(err, "Failed to new epoch poller")

	go worker.MonitorBlockPropagation(cfx, epochCh)
	go epochPoller.Poll(epochCh)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	screen.Clear()

	for range ticker.C {
		if unhandled := len(epochCh); unhandled > 10 {
			logrus.Fatalf("Too many epoch unhandled: %v", unhandled)
		}

		screen.MoveTopLeft()

		ss1 := worker.TimerPropagationPivotBlock.Snapshot()
		ss2 := worker.TimerPropagationAllBlocks.Snapshot()
		ps1 := ss1.Percentiles([]float64{0.5, 0.9, 0.99})
		ps2 := ss2.Percentiles([]float64{0.5, 0.9, 0.99})

		fmt.Printf("%-6s%10s%10s\n", "[TPS]", "Pivot", "All")
		fmt.Printf("%-6s%10.1f%10.1f\n", "mean", ss1.RateMean(), ss2.RateMean())
		fmt.Printf("%-6s%10.1f%10.1f\n", "m1", ss1.Rate1(), ss2.Rate1())
		fmt.Printf("%-6s%10.1f%10.1f\n", "m5", ss1.Rate5(), ss2.Rate5())
		fmt.Printf("%-6s%10.1f%10.1f\n", "m15", ss1.Rate15(), ss2.Rate15())
		fmt.Println()
		fmt.Println("[Latency (secs)]")
		fmt.Printf("%-6s%10.1f%10.1f\n", "min", time.Duration(ss1.Min()).Seconds(), time.Duration(ss2.Min()).Seconds())
		fmt.Printf("%-6s%10.1f%10.1f\n", "mean", time.Duration(ss1.Mean()).Seconds(), time.Duration(ss2.Mean()).Seconds())
		fmt.Printf("%-6s%10.1f%10.1f\n", "p50", time.Duration(ps1[0]).Seconds(), time.Duration(ps2[0]).Seconds())
		fmt.Printf("%-6s%10.1f%10.1f\n", "p90", time.Duration(ps1[1]).Seconds(), time.Duration(ps2[1]).Seconds())
		fmt.Printf("%-6s%10.1f%10.1f\n", "p99", time.Duration(ps1[2]).Seconds(), time.Duration(ps2[2]).Seconds())
		fmt.Printf("%-6s%10.1f%10.1f\n", "max", time.Duration(ss1.Max()).Seconds(), time.Duration(ss2.Max()).Seconds())
	}
}
