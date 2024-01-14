package job

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"github.com/zhulingbiezhi/go12306/config"
	"github.com/zhulingbiezhi/go12306/pkg/account"
	"github.com/zhulingbiezhi/go12306/pkg/order"
	"github.com/zhulingbiezhi/go12306/pkg/query"
	"github.com/zhulingbiezhi/go12306/pkg/train"
	"github.com/zhulingbiezhi/go12306/tools/logger"
)

type UserJob struct {
	Name               string
	Account            *account.AccountHelper
	Seats              []string
	ExceptTrainNumbers []string
	TrainNumbers       []string
	Dates              []string
	Station            *config.Station
	Members            []string
	QueryInterval      int
	AllowLessMember    bool
}

func (job *UserJob) Run(ctx context.Context) error {
	stopChan := make(chan struct{}, 1)
	orderChan := make(chan order.Order, 50)

	ctx = context.WithValue(ctx, "cookie", map[string]*http.Cookie{})

	//wait for first login
	wait, err := job.Account.Login(ctx)
	if err != nil {
		return err
	}
	<-wait
	//query
	go job.TicketQueryJob(ctx, orderChan, stopChan)
	//order
	go job.TicketOrderOrQueue(ctx, orderChan, stopChan)
	return nil
}

func (job *UserJob) Query(ctx context.Context, queueChan chan<- order.Order) error {
	for i := range job.Dates {
		go func(date string) {
			logger.GoroutineInit(logger.WithFields(logrus.Fields{
				"request_id": uuid.NewV1(),
			}))
			logger.Infof("start query job: %s, date: %s", job.Name, date)
			request := job.QueryLeftTicketRequest(date)
			ret, err := query.QueryLeftTicket(request)
			if err != nil {
				logger.Error(err)
				return
			}
			for _, value := range ret {
				orders := job.FilterResult(ctx, value)
				if len(orders) > 0 {
					for i := range orders {
						queueChan <- orders[i]
					}
				}
			}
			logger.Infof("end query job: %s, date: %s", job.Name, date)
		}(job.Dates[i])
	}
	return nil
}

func (job *UserJob) QueryLeftTicketRequest(date string) *query.QueryLeftTicketRequest {
	return &query.QueryLeftTicketRequest{
		FromStation: train.StationMap[job.Station.Left].Key,
		ToStation:   train.StationMap[job.Station.Arrive].Key,
		TrainDate:   date,
		PurposeCode: train.PurposeTypeAdult,
	}
}

func (job *UserJob) FilterResult(ctx context.Context, result *query.TicketResult) []order.Order {
	orders := make([]order.Order, 0)
	queues := make([]order.Order, 0)
	for _, seat := range job.Seats {
		seatKey := train.SeatType[seat]
		seatValue := ""
		switch seat {
		case "高级软卧":
			seatValue = result.SeniorSoftSleeper
		case "软卧", "一等卧":
			seatKey = train.SeatType["软卧"]
			seatValue = result.SoftSleeper
		case "软座":
			seatValue = result.SoftSeat
		case "特等座":
			seatValue = result.SpecialClassSeat
		case "无座":
			seatValue = result.StandingTicket
		case "硬卧":
			seatValue = result.HardSleeper
		case "二等卧":
			seatValue = result.HardSleeper
		case "硬座":
			seatValue = result.HardSeat
		case "二等座":
			seatValue = result.SecondClassSeat
		case "一等座":
			seatValue = result.FirstClassSeat
		case "商务座":
			seatValue = result.BusinessClassSeat
		case "动卧":
			seatValue = result.CRHSleeper
		default:
			logger.Errorf("not support seat type %s", seat)
		}
		if seatKey == "" || seatValue == "" {
			continue
		}
		//logger.Info(result.TrainCode, "---", seat, "---", seatValue)
		if seatValue == "无" {
			if result.HoubuTrainFlag == "1" {
				//logger.Info("add to queue: ", result.TrainCode, "---", seat, "---", seatValue)
				queues = append(queues, order.BuildOrderQueue(ctx, result.Secret, seatKey, result.TrainCode))
			}
		} else if seatValue != "" && seatValue != "*" {
			flag := false
			if seatValue == "有" {
				flag = true
			} else {
				num, err := strconv.ParseInt(seatValue, 10, 64)
				if err != nil {
					logger.Error(err)
					continue
				}
				if int(num) >= len(job.Members) {
					flag = true
				} else if job.AllowLessMember {
					flag = true
				}
			}
			if flag {
				logger.Info("add to order: ", result.TrainCode, "---", seat, "---", seatValue)
				orders = append(orders, order.BuildOrderTicket(ctx, result.Secret, seatKey))
			}
		}
	}
	//优先抢order的队列
	orders = append(orders, queues...)
	return orders
}

func (job *UserJob) TicketQueryJob(ctx context.Context, orderChan chan<- order.Order, stopChan <-chan struct{}) {
	logger.GoroutineInit(logger.WithFields(logrus.Fields{
		"request_id": uuid.NewV1(),
	}))
	sec := time.Second * time.Duration(job.QueryInterval)
	t := time.NewTimer(sec)
	index := 0
	for {
		index++
		select {
		case <-t.C:
			logger.Infof("query timer %d", index)
			if err := job.Query(ctx, orderChan); err != nil {
				logger.Error("job query err", err)
			}
		case <-stopChan:
			t.Stop()
			return
		}
		t.Reset(sec)
	}
}

func (job *UserJob) TicketOrderOrQueue(ctx context.Context, orderChan <-chan order.Order, stopChan chan<- struct{}) {
	logger.GoroutineInit(logger.WithFields(logrus.Fields{
		"request_id": uuid.NewV1(),
	}))
	for {
		select {
		case o := <-orderChan:
			if err := o.Check(ctx); err != nil {
				logger.Error("orderChan check err", err)
				break
			}
			if err := o.Submit(ctx); err != nil {
				logger.Error("orderChan submit err", err)
				break
			}
			stopChan <- struct{}{}
			return
		}
	}
}
