package worker

import (
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

var TimerPropagationPivotBlock = metrics.GetOrRegisterTimer("monitor/propagate/pivotBlock", nil)
var TimerPropagationAllBlocks = metrics.GetOrRegisterTimer("monitor/propagate/allBlocks", nil)

func MonitorBlockPropagation(cfx sdk.ClientOperator, epochCh <-chan uint64) {
	for epoch := range epochCh {
		for {
			if err := statBlockPropagation(cfx, epoch); err != nil {
				logrus.WithError(err).WithField("epoch", epoch).Info("Failed to statistic block propagation")
			} else {
				break
			}
		}
	}
}

func statBlockPropagation(cfx sdk.ClientOperator, epochNum uint64) error {
	epoch := types.NewEpochNumberUint64(epochNum)

	// Use current time as block execution time, since epoch polled timely.
	now := time.Now()

	blocks, err := cfx.GetBlocksByEpoch(epoch)
	if err != nil {
		return errors.WithMessagef(err, "Failed to get blocks by epoch %v", epochNum)
	}

	for i, v := range blocks {
		block, err := cfx.GetBlockSummaryByHash(v)
		if err != nil {
			return errors.WithMessagef(err, "Failed to get block summary, epoch = %v, hash = %v", epochNum, v)
		}

		mineTime := time.Unix(block.Timestamp.ToInt().Int64(), 0)
		TimerPropagationAllBlocks.Update(now.Sub(mineTime))

		if i == len(blocks)-1 {
			TimerPropagationPivotBlock.Update(now.Sub(mineTime))
		}
	}

	return nil
}
