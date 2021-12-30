package cmd

import (
	"fmt"
	"time"

	"github.com/Conflux-Chain/conflux-monitor/epoch"
	"github.com/Conflux-Chain/conflux-monitor/worker"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/inancgumus/screen"
	"github.com/rcrowley/go-metrics"
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

		fmt.Println("========== Pivot Block Statistic ==========")
		printStat(worker.TimerPropagationPivotBlock)

		fmt.Println()
		fmt.Println()
		fmt.Println()

		fmt.Println("========== All Blocks Statistic ==========")
		printStat(worker.TimerPropagationAllBlocks)
	}
}

func printStat(timer metrics.Timer) {
	ss := timer.Snapshot()
	ps := ss.Percentiles([]float64{0.5, 0.9, 0.99})

	fmt.Println("[TPS]")
	fmt.Printf("mean: %10.1f\n", ss.RateMean())
	fmt.Printf("m1:   %10.1f\n", ss.Rate1())
	fmt.Printf("m5:   %10.1f\n", ss.Rate5())
	fmt.Printf("m15:  %10.1f\n", ss.Rate15())
	fmt.Println()
	fmt.Println("[Latency]")
	fmt.Printf("Min:  %10.1fs\n", time.Duration(ss.Min()).Seconds())
	fmt.Printf("Mean: %10.1fs\n", time.Duration(ss.Mean()).Seconds())
	fmt.Printf("P50:  %10.1fs\n", time.Duration(ps[0]).Seconds())
	fmt.Printf("P90:  %10.1fs\n", time.Duration(ps[1]).Seconds())
	fmt.Printf("P99:  %10.1fs\n", time.Duration(ps[2]).Seconds())
	fmt.Printf("Max:  %10.1fs\n", time.Duration(ss.Max()).Seconds())
}
