package order

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zhulingbiezhi/go12306/common"
	"github.com/zhulingbiezhi/go12306/tools/conf"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/rest"
)

type OrderQueue struct {
	RestClient *rest.RestHttp
	Secret     string
	Seat       string
	TrainNo    string
}

func BuildOrderQueue(ctx context.Context, secret, seat string, trainNo string) Order {
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm).SetRestLogOption(&rest.LogOption{
		LogBody:   false,
		LogHeader: false,
	})
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		common.Cookie_RAIL_EXPIRATION: conf.Conf.RailExpire,
		common.Cookie_RAIL_DEVICEID:   conf.Conf.RailDevice,
	}))
	cookie, ok := ctx.Value("cookie").(map[string]*http.Cookie)
	if ok {
		rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
			cookie[common.Cookie_Apptk],
		}))
	}
	rs.SetHeader(map[string]interface{}{
		common.Header_USER_AGENT: common.UserAgentChrome,
	})
	return &OrderQueue{
		RestClient: rs,
		Secret:     secret,
		Seat:       seat,
		TrainNo:    trainNo,
	}
}

func (q *OrderQueue) Submit(ctx context.Context) error {
	vals := "secretList=" + q.Secret + "#" + q.Seat + "|"
	//resp := ChechFaceResponse{}
	body, err := q.RestClient.Do(http.MethodPost, conf.API_SUBMIT_QUEUE_TICKET_URL, vals)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func (q *OrderQueue) Check(ctx context.Context) error {
	logger.Info("start queue check ", q.TrainNo)
	err := q.ChechFace(ctx)
	if err != nil {
		return errors.Errorf(err, "ChechFace err")
	}
	success, err := q.GetSuccessRate(ctx)
	if err != nil {
		return errors.Errorf(err, "GetSuccessRate err")
	}
	if !success {
		return errors.Errorf(nil, "queue is too long")
	}
	logger.Info("end queue check")
	return nil
}

type ChechFaceResponse struct {
	ValidateMessagesShowID string `json:"validateMessagesShowId"`
	Status                 bool   `json:"status"`
	HttpStatus             int    `json:"httpstatus"`
	Data                   struct {
		LoginFlag     bool   `json:"login_flag"`
		IsShowQrcode  bool   `json:"is_show_qrcode"`
		FaceCheckCode string `json:"face_check_code"`
		FaceFlag      bool   `json:"face_flag"`
	} `json:"data"`
	Messages         []interface{} `json:"messages"`
	ValidateMessages struct {
	} `json:"validateMessages"`
}

func (q *OrderQueue) ChechFace(ctx context.Context) error {
	vals := "secretList=" + q.Secret + "#" + q.Seat + "|" + "&_json_att="
	resp := ChechFaceResponse{}
	rs := q.RestClient
	rs.SetHeader(map[string]interface{}{
		"Referer": "https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc",
		"Origin":  "https://kyfw.12306.cn",
	})
	body, err := rs.DoRest(http.MethodPost, conf.API_CHECH_QUEUE_TICKET_URL, vals).ParseJsonBody(&resp)
	if err != nil {
		return err
	}
	if resp.HttpStatus != http.StatusOK || resp.Status != true {
		return errors.Errorf(nil, "http code %d != 200, reason: %+v body : %+v", resp.HttpStatus, resp.Messages, string(body))
	}
	if !resp.Data.LoginFlag {
		return errors.Errorf(errors.LoginErr, "user not login")
	}
	if !resp.Data.FaceFlag {
		return errors.Errorf(nil, "response is not success")
	}
	return nil
}

func (q *OrderQueue) GetSuccessRate(ctx context.Context) (bool, error) {
	vals := "successSecret=" + q.Secret + "#" + q.Seat + "|"

	//resp := ChechFaceResponse{}
	body, err := q.RestClient.Do(http.MethodPost, conf.API_QUEUE_SUCCESS_RATE_URL, vals)
	if err != nil {
		return false, err
	}
	fmt.Println(string(body))
	return false, nil
}
