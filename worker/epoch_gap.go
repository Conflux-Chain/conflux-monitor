package worker

import (
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

var TimerEpochGaps = metrics.GetOrRegisterTimer("monitor/epochGaps/statedConfirmed", nil)

func MonitorEpochGaps(cfx sdk.ClientOperator, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		status, err := cfx.GetStatus()
		if err != nil {
			logrus.WithError(err).Warn("Failed to get status from fullnode")
			continue
		}

		epochGap := int64(status.LatestState) - int64(status.LatestConfirmed)
		TimerEpochGaps.Update(time.Duration(epochGap))
	}
}
