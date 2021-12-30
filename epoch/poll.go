package epoch

import (
	"sync/atomic"
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/sirupsen/logrus"
)

type PollOption struct {
	EpochTag         *types.Epoch  // epoch tag to poll against, e.g. latest_state, latest_mined
	Interval         time.Duration // poll interval
	EpochsFallBehind uint64        // number of epochs that fall behind the specified epoch tag
}

func (opt *PollOption) applyDefault() {
	if opt.EpochTag == nil {
		opt.EpochTag = types.EpochLatestState
	}

	if opt.Interval == 0 {
		opt.Interval = 300 * time.Millisecond
	}
}

// EpochPoller polls epoch to handle in advance. Note, the `EpochPollHandler` shall be tolerant with
// pivot chain switch. Otherwise, please consider pub/sub or filter API instead.
type Poller struct {
	cfx       sdk.ClientOperator
	option    PollOption
	epochFrom uint64
}

func NewPoller(cfx sdk.ClientOperator, option ...PollOption) (*Poller, error) {
	var opt PollOption
	if len(option) > 0 {
		opt = option[0]
	}
	opt.applyDefault()

	epoch, err := cfx.GetEpochNumber(opt.EpochTag)
	if err != nil {
		return nil, err
	}

	return &Poller{
		cfx:       cfx,
		option:    opt,
		epochFrom: epoch.ToInt().Uint64() - opt.EpochsFallBehind,
	}, nil
}

func (poller *Poller) Poll(epochCh chan<- uint64) {
	ticker := time.NewTicker(poller.option.Interval)
	defer ticker.Stop()

	for range ticker.C {
		epoch, err := poller.cfx.GetEpochNumber(poller.option.EpochTag)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get epoch number")
			continue
		}

		epochFrom := atomic.LoadUint64(&poller.epochFrom)
		epochTo := epoch.ToInt().Uint64() - poller.option.EpochsFallBehind

		if epochFrom > epochTo {
			continue
		}

		for i := epochFrom; i <= epochTo; i++ {
			epochCh <- i
		}

		atomic.StoreUint64(&poller.epochFrom, epochTo+1)
	}
}
