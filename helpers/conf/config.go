package conf

import (
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/spf13/viper"
	"go12306/helpers/logger"
)

var Conf *config

type config struct {
	Jobs           []map[string]interface{} `mapstructure:"jobs"`
	Users          []map[string]interface{} `mapstructure:"users"`
	RailDevice     string                   `mapstructure:"rail_device"`
	RailExpire     string                   `mapstructure:"rail_expire"`
	LoginHeartBeat int                      `mapstructure:"login_heart_beat"`
}

func init() {
	var err error
	Conf, err = LoadConf()
	if err != nil {
		panic(err)
	}
}

func LoadConf() (*config, error) {
	viper.SetConfigName("testing")
	//viper.AddConfigPath("config")
	viper.AddConfigPath("/Users/klook/work/golang/src/go12306/config")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	cfg := &config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	logger.Info(awsutil.Prettify(cfg))
	return cfg, nil
}
