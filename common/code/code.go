package code

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/zhulingbiezhi/go12306/common"
	"github.com/zhulingbiezhi/go12306/helpers/conf"
	"github.com/zhulingbiezhi/go12306/helpers/errors"
	"github.com/zhulingbiezhi/go12306/helpers/logger"
	"github.com/zhulingbiezhi/go12306/helpers/rest"
	"github.com/zhulingbiezhi/go12306/utils"
	"image"
	"image/jpeg"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type CodeResponse struct {
	Image         string `json:"image"`
	ResultMessage string `json:"result_message"`
	ResultCode    string `json:"result_code"`
}

func GetAuthCode(ctx context.Context, vals url.Values) (string, error) {
	if len(vals) == 0 {
		return "", errors.Errorf(nil, "data is empty")
	}
	code := CodeResponse{}
	//get code image
	{
		logger.Info("get auth code image")
		urlStr := conf.API_AUTH_CODE_BASE64_DOWNLOAD + "?" + vals.Encode()
		rs := rest.NewHttp()
		_, err := rs.DoRest(http.MethodGet, urlStr, nil).ParseJsonBody(&code)
		if err != nil {
			return "", err
		}
		cookie, ok := ctx.Value("cookie").(map[string]*http.Cookie)
		if ok {
			for key, ck := range rs.RespCookies() {
				switch key {
				case common.Cookie_PassportCt, common.Cookie_PassportSession:
					cookie[ck.Name] = rs.RespCookies()[key]
				}
			}
		}
		if code.ResultCode != "0" || code.ResultMessage != "生成验证码成功" {
			return "", errors.Errorf(nil, "code: %s, message: %s", code.ResultCode, code.ResultMessage)
		}
		logger.Info("get auth code image success")
	}
	//save picture
	index := 0
	if false {
		index++
		//save picture
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(code.Image))
		img, _, err := image.Decode(reader)
		if err != nil {
			return "", errors.Errorf(err, "image.Decode err")
		}
		dir := utils.GetCurDir()
		//Encode from image format to writer
		dstFileName := dir + fmt.Sprintf("/code_%d.jpeg", index)
		f, err := os.OpenFile(dstFileName, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			return "", err
		}
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return "", err
		}
	}
	//image recognition
	ret, err := GetVerifyCode(code.Image)
	if err != nil {
		return "", errors.Errorf(err, "GetVerifyCode err")
	}
	//check auth code
	answer := strings.Join(getImagePositionByOffset(ret), ",")
	success, err := CheckAuthCode(ctx, answer)
	if err != nil {
		return "", errors.Errorf(err, "CheckAuthCode err")
	}
	if !success {
		return "", errors.Errorf(nil, "CheckAuthCode fail")
	}
	return answer, nil
}

func GetVerifyCode(img string) ([]int, error) {
	logger.Info("get auth code from OCR")
	vals := make(url.Values, 0)
	vals.Set("img", img)
	ret := struct {
		Msg    string `json:"msg"`
		Result []int  `json:"result"`
	}{}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	_, err := rs.DoRest(http.MethodPost, conf.API_FREE_CODE_QCR_API_URL, vals.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Msg != "success" || len(ret.Result) == 0 {
		return nil, errors.Errorf(nil, "result is not success %+v", ret)
	}
	logger.Info("get auth code from OCR success")
	return ret.Result, nil
}

func CheckAuthCode(ctx context.Context, answer string) (bool, error) {
	logger.Info("check auth code")
	vals := url.Values{
		"login_site": []string{"E"},
		"module":     []string{"login"},
		"rand":       []string{"sjrand"},
		"_":          []string{strconv.FormatFloat(rand.Float64(), 'f', -1, 64)},
		"answer":     []string{answer},
	}
	urlStr := conf.API_AUTH_CODE_CHECK_URL + "?" + vals.Encode()
	ret := struct {
		ResultMessage string `json:"result_message"`
		ResultCode    string `json:"result_code"`
	}{}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	cookie, ok := ctx.Value("cookie").(map[string]*http.Cookie)
	if ok {
		rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
			cookie[common.Cookie_PassportSession],
			cookie[common.Cookie_PassportCt],
		}))
	}
	_, err := rs.DoRest(http.MethodGet, urlStr, nil).ParseJsonBody(&ret)
	if err != nil {
		return false, err
	}
	bSuccess := ret.ResultCode == "4"
	if bSuccess {
		logger.Info("check auth code success")
	}
	return bSuccess, nil
}

func getImagePositionByOffset(offsets []int) []string {
	positions := make([]string, 0)
	width := (75)
	height := (75)
	for _, offset := range offsets {
		random_x := rand.Intn(20) + -10
		random_y := rand.Intn(20) + -10
		offset = int(offset)
		x := width*((offset-1)%4+1) - width/2 + random_x
		y := height*int(math.Ceil(float64(offset)/4)) - height/2 + random_y

		positions = append(positions, strconv.Itoa(x))
		positions = append(positions, strconv.Itoa(y))
	}
	return positions
}
