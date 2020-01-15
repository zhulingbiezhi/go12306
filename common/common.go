package common

import (
	"io/ioutil"
	"strings"
)

var (
	SeatType = map[string]string{
		"高级动卧": "A",
		"一等卧":  "I",
		"二等卧":  "J",
		"特等座":  "P",
		"一等座":  "M",
		"二等座":  "O",
		"动卧":   "F",
		"商务座":  "9",
		"高级软卧": "6",
		"软卧":   "4",
		"硬卧":   "3",
		"软座":   "2",
		"硬座":   "1",
		"其他":   "H",
		"无座":   "WZ",
	}
)

const (
	Header_USER_AGENT      = "User-Agent"
	Cookie_RAIL_DEVICEID   = "RAIL_DEVICEID"
	Cookie_RAIL_EXPIRATION = "RAIL_EXPIRATION"
	Cookie_PassportSession = "_passport_session"
	Cookie_PassportCt      = "_passport_ct"
	Cookie_Uamtk           = "uamtk"
	Cookie_Apptk           = "tk"

	UserAgentChrome = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36"
)

const (
	Query_TrainDate    = "leftTicketDTO.train_date"
	Query_FromStation  = "leftTicketDTO.from_station"
	Query_ToStation    = "leftTicketDTO.to_station"
	Query_PurposeCodes = "purpose_codes"
)

type PurposeType string

func (t PurposeType) String() string {
	return string(t)
}

const (
	PurposeTypeAdult   PurposeType = "ADULT"
	PurposeTypeStudent PurposeType = "0X00"
)

type StationInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Key      string `json:"key"`
	Spelling string `json:"spelling"`
}

var StationMap = make(map[string]*StationInfo)

func init() {
	b, err := ioutil.ReadFile("/Users/klook/work/golang/src/go12306/config/stations.txt")
	if err != nil {
		panic(err)
	}
	result := strings.Split(string(b), "@")
	for _, value := range result {
		infos := strings.Split(value, "|")
		if len(infos) == 6 {
			StationMap[infos[1]] = &StationInfo{
				ID:       infos[5],
				Name:     infos[1],
				Key:      infos[2],
				Spelling: infos[3],
			}
		}
	}
}
