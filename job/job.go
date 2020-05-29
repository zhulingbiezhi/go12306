package job

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"github.com/zhulingbiezhi/go12306/common"
	"github.com/zhulingbiezhi/go12306/helpers/conf"
	"github.com/zhulingbiezhi/go12306/helpers/errors"
	"github.com/zhulingbiezhi/go12306/helpers/logger"
	"github.com/zhulingbiezhi/go12306/order"
	"github.com/zhulingbiezhi/go12306/query"
	"github.com/zhulingbiezhi/go12306/user"
	"net/http"
	"strconv"
	"time"
)

var Jobs []*UserJob

type UserJob struct {
	Name               string   `mapstructure:"name"`
	Account            string   `mapstructure:"account"`
	Seats              []string `mapstructure:"seats"`
	ExceptTrainNumbers []string `mapstructure:"except_train_numbers"`
	TrainNumbers       []string `mapstructure:"train_numbers"`
	Dates              []string `mapstructure:"dates"`
	Station            *Station `mapstructure:"station"`
	Members            []string `mapstructure:"members"`
	QueryInterval      int      `mapstructure:"query_interval"`
	AllowLessMember    bool     `mapstructure:"allow_less_member"`
}

type Station struct {
	Left   string `mapstructure:"left"`
	Arrive string `mapstructure:"arrive"`
}

func init() {
	jobs := make([]*UserJob, 0)
	for _, jb := range conf.Conf.Jobs {
		job := &UserJob{}
		err := mapstructure.Decode(jb, job)
		if err != nil {
			logger.Error("mapstructure.Decode err", err)
		} else {
			jobs = append(jobs, job)
		}
	}
	logger.Info(awsutil.Prettify(jobs))
	Jobs = jobs
}

func (job *UserJob) Run(ctx context.Context) error {
	stopChan := make(chan struct{}, 1)
	loginChan := make(chan struct{}, 1)
	loginSuccessChan := make(chan bool, 1)
	orderChan := make(chan order.Order, 50)

	ctx = context.WithValue(ctx, "cookie", map[string]*http.Cookie{})

	//login heart beat
	go job.LoginJob(ctx, loginChan, loginSuccessChan, stopChan)
	//wait for first login
	WaitForLogin(loginChan, loginSuccessChan)
	//query
	go job.TicketQueryJob(ctx, orderChan, stopChan)
	//order
	go job.TicketOrderOrQueue(ctx, orderChan, stopChan)
	return nil
}

func (job *UserJob) Login(ctx context.Context) error {
	u, err := user.GetUser(job.Account)
	if err != nil {
		return errors.Errorf(err, "GetUser err")
	}
	return u.Login(ctx)
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
		FromStation: common.StationMap[job.Station.Left].Key,
		ToStation:   common.StationMap[job.Station.Arrive].Key,
		TrainDate:   date,
		PurposeCode: common.PurposeTypeAdult,
	}
}

func (job *UserJob) FilterResult(ctx context.Context, result *query.TicketResult) []order.Order {
	orders := make([]order.Order, 0)
	queues := make([]order.Order, 0)
	for _, seat := range job.Seats {
		seatKey := common.SeatType[seat]
		seatValue := ""
		switch seat {
		case "高级软卧":
			seatValue = result.SeniorSoftSleeper
		case "软卧", "一等卧":
			seatKey = common.SeatType["软卧"]
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

func WaitForLogin(loginChan chan<- struct{}, successChan chan bool) {
	loginChan <- struct{}{}
	for {
		select {
		case success := <-successChan:
			if success {
				return
			}
			loginChan <- struct{}{}
		}
	}
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
			t.Reset(sec)
			if err := job.Query(ctx, orderChan); err != nil {
				logger.Error("job query err", err)
			}
		case <-stopChan:
			t.Stop()
			return
		}
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

func (job *UserJob) LoginJob(ctx context.Context, loginChan <-chan struct{}, loginSuccessChan chan bool, stopChan <-chan struct{}) {
	logger.GoroutineInit(logger.WithFields(logrus.Fields{
		"request_id": uuid.NewV1(),
	}))
	sec := time.Second * time.Duration(conf.Conf.LoginHeartBeat)
	t := time.NewTimer(sec)
	for {
		select {
		case <-t.C:
			if err := job.Login(ctx); err != nil {
				logger.Error("job login err:", err)
				//5秒后重试登录
				t.Reset(time.Second * 5)
			} else {
				logger.Info("job login success")
				t.Reset(time.Second * time.Duration(conf.Conf.LoginHeartBeat))
			}
		case <-loginChan:
			if err := job.Login(ctx); err != nil {
				logger.Error("loginChan---job login err:", err)
				loginSuccessChan <- false
			} else {
				logger.Info("loginChan---job login success")
				loginSuccessChan <- true
			}
		case <-stopChan:
			t.Stop()
			return
		}
	}
}
