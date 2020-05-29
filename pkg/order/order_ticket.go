package order

import (
	"context"
	"net/http"

	"github.com/zhulingbiezhi/go12306/common"
	"github.com/zhulingbiezhi/go12306/tools/conf"
	"github.com/zhulingbiezhi/go12306/tools/rest"
)

type Order interface {
	Check(ctx context.Context) error
	Submit(ctx context.Context) error
}

type OrderTicket struct {
	RestClient *rest.RestHttp
	Secret     string
	Seat       string
}

func BuildOrderTicket(ctx context.Context, secret, seat string) Order {
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
	return &OrderTicket{
		RestClient: rs,
		Secret:     secret,
		Seat:       seat,
	}
}

func (o *OrderTicket) Check(ctx context.Context) error {
	return nil
}

func (o *OrderTicket) Submit(ctx context.Context) error {
	return nil
}