package conf

import (
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/spf13/viper"
	"github.com/zhulingbiezhi/go12306/pkg/helper"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/utils"
)

var Conf *config

type config struct {
	Jobs           []map[string]interface{} `mapstructure:"jobs"`
	Accounts       []map[string]interface{} `mapstructure:"accounts"`
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
	b, err := ioutil.ReadFile(utils.GetCurDir() + "/config/stations.txt")
	if err != nil {
		panic(err)
	}
	result := strings.Split(string(b), "@")
	for _, value := range result {
		infos := strings.Split(value, "|")
		if len(infos) == 6 {
			helper.StationMap[infos[1]] = &helper.StationInfo{
				ID:       infos[5],
				Name:     infos[1],
				Key:      infos[2],
				Spelling: infos[3],
			}
		}
	}
}

func LoadConf() (*config, error) {
	viper.SetConfigName("testing")
	viper.AddConfigPath(utils.GetCurDir() + "/config")
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
