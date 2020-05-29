package query

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/zhulingbiezhi/go12306/pkg/common"
	"github.com/zhulingbiezhi/go12306/pkg/train"
	"github.com/zhulingbiezhi/go12306/tools/conf"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/rest"
)

func GetQueryApiUrl() (string, error) {
	rs := rest.NewHttp()
	body, err := rs.Do(http.MethodGet, common.API_QUERY_INIT_PAGE_URL, nil)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile("var CLeftTicketUrl = '(.*)';")
	match := re.FindStringSubmatch(string(body))
	return common.LEFT_TICKETS_URL + "/" + match[1], nil
}

type QueryLeftTicketRequest struct {
	FromStation string
	ToStation   string
	TrainDate   string
	PurposeCode train.PurposeType
}

type QueryLeftTicketResponse struct {
	HttpStatus int    `json:"httpstatus"`
	Reason     string `json:"reason"`
	Data       struct {
		Result []string `json:"result"`
		Flag   string   `json:"flag"`
	}
	Status bool `json:"status"`
}

type TicketResult struct {
	Secret              string `json:"secret"`                // 0
	ButtonTextInfo      string `json:"button_text_info"`      // 1
	TrainNumber         string `json:"train_number"`          // 2
	TrainCode           string `json:"train_code"`            // 3
	StartStation        string `json:"start_station"`         // 4
	EndStation          string `json:"end_station"`           // 5
	FromStation         string `json:"from_station"`          // 6
	ToStation           string `json:"to_station"`            // 7
	StartTime           string `json:"start_time"`            // 8
	EndTime             string `json:"end_time"`              // 9
	CostTime            string `json:"cost_time"`             // 10
	CanBuy              bool   `json:"can_buy"`               // 11
	YPInfo              string `json:"yp_info"`               // 12
	StartTrainDate      string `json:"start_train_date"`      // 13
	TrainSeatFeature    string `json:"train_seat_feature"`    // 14
	LocationCode        string `json:"location_code"`         // 15
	FromStationNo       string `json:"from_station_no"`       // 16
	ToStationNo         string `json:"to_station_no"`         // 17
	IsSupportCard       string `json:"is_support_card"`       // 18
	ControlledTrainFlag string `json:"controlled_train_flag"` // 19
	gg_num              string `json:"gg_num"`                // 20 gg_num
	SeniorSoftSleeper   string `json:"senior_soft_sleeper"`   // 21 高级软卧
	qt_num              string `json:"qt_num"`                // 22 qt_num
	SoftSleeper         string `json:"soft_sleeper"`          // 23 软卧、一等卧
	SoftSeat            string `json:"soft_seat"`             // 24 软座
	SpecialClassSeat    string `json:"special_class_seat"`    // 25 特等座
	StandingTicket      string `json:"standing_ticket"`       // 26 无座、站票
	yb_num              string `json:"yb_num"`                // 27 yb_num
	HardSleeper         string `json:"hard_sleeper"`          // 28 硬卧、二等卧
	HardSeat            string `json:"hard_seat"`             // 29 硬座
	SecondClassSeat     string `json:"second_class_seat"`     // 30 二等座
	FirstClassSeat      string `json:"first_class_seat"`      // 31 一等座
	BusinessClassSeat   string `json:"business_class_seat"`   // 32 商务座
	CRHSleeper          string `json:"crh_sleeper"`           // 33 动卧
	YpEx                string `json:"yp_ex"`                 // 34
	SeatTypes           string `json:"seat_types"`            // 35
	ExchangeTrainFlag   string `json:"exchange_train_flag"`   // 36
	HoubuTrainFlag      string `json:"houbu_train_flag"`      // 37
	HoubuSeatLimit      string `json:"houbu_seat_limit"`      // 38
	FwFlag              string `json:"dw_flag"`               // 46
}

func QueryLeftTicket(request *QueryLeftTicketRequest) ([]*TicketResult, error) {
	urlStr, err := GetQueryApiUrl()
	if err != nil {
		return nil, errors.Errorf(err, "GetQueryApiUrl err")
	}
	vals := url.Values{}
	vals.Set(train.Query_TrainDate, request.TrainDate)
	vals.Set(train.Query_FromStation, request.FromStation)
	vals.Set(train.Query_ToStation, request.ToStation)
	vals.Set(train.Query_PurposeCodes, request.PurposeCode.String())
	rs := rest.NewHttp()
	rs.SetHeader(map[string]interface{}{
		common.Header_USER_AGENT: common.UserAgentChrome,
	})
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		common.Cookie_RAIL_EXPIRATION: conf.Conf.RailExpire,
		common.Cookie_RAIL_DEVICEID:   conf.Conf.RailDevice,
	}))
	//特定排序
	subUrlFormat := "leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=%s"
	urlStr = urlStr + "?" + fmt.Sprintf(subUrlFormat, request.TrainDate, request.FromStation, request.ToStation, request.PurposeCode)
	ret := QueryLeftTicketResponse{}
	body, err := rs.DoRest(http.MethodGet, urlStr, nil).ParseJsonBody(&ret)
	if err != nil {
		return nil, err
	}
	if ret.HttpStatus != http.StatusOK || ret.Status != true || len(ret.Data.Result) == 0 || ret.Data.Flag != "1" {
		return nil, errors.Errorf(nil, "http code %d != 200, reason: %s body : %+v", ret.HttpStatus, ret.Reason, string(body))
	}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	results := make([]*TicketResult, 0)
	for _, result := range ret.Data.Result {
		ret := strings.Split(result, "|")
		//arr := make([]string, 0)
		//for key, value := range ret {
		//	if utils.IntIsIn(key, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13) {
		//		continue
		//	}
		//	arr = append(arr, fmt.Sprintf("%d-%s", key, value))
		//}
		//fmt.Println(strings.Join(arr, "|"))
		results = append(results, &TicketResult{
			Secret:              ret[0],
			ButtonTextInfo:      ret[1],
			TrainNumber:         ret[2],
			TrainCode:           ret[3],
			StartStation:        ret[4],
			EndStation:          ret[5],
			FromStation:         ret[6],
			ToStation:           ret[7],
			StartTime:           ret[8],
			EndTime:             ret[9],
			CostTime:            ret[10],
			CanBuy:              ret[11] == "Y",
			YPInfo:              ret[12],
			StartTrainDate:      ret[13],
			TrainSeatFeature:    ret[14],
			LocationCode:        ret[15],
			FromStationNo:       ret[16],
			ToStationNo:         ret[17],
			IsSupportCard:       ret[18],
			ControlledTrainFlag: ret[19],
			gg_num:              ret[20],
			SeniorSoftSleeper:   ret[21],
			qt_num:              ret[22],
			SoftSleeper:         ret[23],
			SoftSeat:            ret[24],
			SpecialClassSeat:    ret[25],
			StandingTicket:      ret[26],
			yb_num:              ret[27],
			HardSleeper:         ret[28],
			HardSeat:            ret[29],
			SecondClassSeat:     ret[30],
			FirstClassSeat:      ret[31],
			BusinessClassSeat:   ret[32],
			CRHSleeper:          ret[33],
			YpEx:                ret[34],
			SeatTypes:           ret[35],
			ExchangeTrainFlag:   ret[36],
			HoubuTrainFlag:      ret[37],
			HoubuSeatLimit:      ret[38],
			FwFlag:              ret[46],
		})
	}
	return results, nil
}
