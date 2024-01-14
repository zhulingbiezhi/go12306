package account

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/zhulingbiezhi/go12306/config"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
)

type AccountHelper struct {
	*Account
	ctxVal     atomic.Value
	cancelChan chan context.CancelCauseFunc
}

func (ah *AccountHelper) Init() {
	go func() {
		ah.Loop()
	}()
}

func (ah *AccountHelper) Login(ctx1 context.Context) (<-chan struct{}, error) {
	ctx, ok := ah.ctxVal.Load().(context.Context)
	if ok {
		return ctx.Done(), nil
	}
	ctx, cancel := context.WithCancelCause(ctx1)
	ah.ctxVal.Store(ctx)
	select {
	case ah.cancelChan <- cancel:
	default:
		logger.Error()
		cancel(errors.New("cancelChan is full"))
	}
	return ctx.Done(), context.Cause(ctx)
}

func (ah *AccountHelper) Loop() {
	ah.cancelChan = make(chan context.CancelCauseFunc)
	sec := time.Second * time.Duration(config.Conf.LoginHeartBeat)
	t := time.NewTicker(sec)
	for {
		select {
		case cancel := <-ah.cancelChan:
			err := ah.Account.Login(context.Background())
			if err != nil {
				logger.Error(err)
				cancel(err)
			} else {
				cancel(nil)
			}
		case <-t.C:
			ctx, cancel := context.WithCancelCause(context.Background())
			ah.ctxVal.Store(ctx)
			err := ah.Account.Login(ctx)
			if err != nil {
				logger.Error(err)
				cancel(err)
			} else {
				cancel(nil)
			}
		}
		t.Reset(sec)
	}
}
