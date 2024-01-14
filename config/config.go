package config

import (
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/spf13/viper"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/utils"
)

var Conf *Config

type Config struct {
	Accounts       []*Account `yaml:"accounts"`
	RailDevice     string     `yaml:"rail_device"`
	RailExpire     string     `yaml:"rail_expire"`
	LoginHeartBeat int        `yaml:"login_heart_beat"`
	Jobs           []*Job     `yaml:"jobs"`
}

type Account struct {
	AccountName     string `yaml:"account_name"`
	AccountPassword string `yaml:"account_password"`
	UUID            string `yaml:"uuid"`
}

type Station struct {
	Left   string `yaml:"left"`
	Arrive string `yaml:"arrive"`
}

type Job struct {
	Station            *Station `yaml:"station"`
	Name               string   `yaml:"name"`
	Account            string   `yaml:"account"`
	Dates              []string `yaml:"dates"`
	Members            []string `yaml:"members"`
	AllowLessMember    bool     `yaml:"allow_less_member"`
	Seats              []string `yaml:"seats"`
	TrainNumbers       []string `yaml:"train_numbers"`
	ExceptTrainNumbers []string `yaml:"except_train_numbers"`
	QueryInterval      int      `yaml:"query_interval"`
	QueueTicketFlag    bool     `yaml:"queue_ticket_flag"`
}

func LoadConf() (*Config, error) {
	viper.SetConfigName("testing")
	viper.AddConfigPath(utils.GetCurDir() + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	Conf = &Config{}
	if err := viper.Unmarshal(Conf); err != nil {
		return nil, err
	}
	logger.Info(awsutil.Prettify(Conf))
	return Conf, nil
}
