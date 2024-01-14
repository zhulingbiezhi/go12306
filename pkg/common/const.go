package common

// header、cookie相关
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

// url
const (
	BASE_URL_OF_12306                = "https://kyfw.12306.cn"
	LEFT_TICKETS_URL                 = BASE_URL_OF_12306 + "/otn"
	API_BASE_LOGIN_URL               = BASE_URL_OF_12306 + "/passport/web/login"
	API_USER_LOGIN_CHECK             = BASE_URL_OF_12306 + "/otn/login/conf"
	API_AUTH_CODE_DOWNLOAD_URL       = BASE_URL_OF_12306 + "/passport/captcha/captcha-image"
	API_AUTH_CODE_BASE64_DOWNLOAD    = BASE_URL_OF_12306 + "/passport/captcha/captcha-image64"
	API_AUTH_CODE_CHECK_URL          = BASE_URL_OF_12306 + "/passport/captcha/captcha-check"
	API_AUTH_UAMTK_URL               = BASE_URL_OF_12306 + "/passport/web/auth/uamtk"
	API_AUTH_UAMAUTHCLIENT_URL       = BASE_URL_OF_12306 + "/otn/uamauthclient"
	API_USER_INFO_URL                = BASE_URL_OF_12306 + "/otn/modifyUser/initQueryUserInfoApi"
	API_USER_PASSENGERS_URL          = BASE_URL_OF_12306 + "/otn/confirmPassenger/getPassengerDTOs"
	API_SUBMIT_ORDER_REQUEST_URL     = BASE_URL_OF_12306 + "/otn/leftTicket/submitOrderRequest"
	API_CHECK_ORDER_INFO_URL         = BASE_URL_OF_12306 + "/otn/confirmPassenger/checkOrderInfo"
	API_INITDC_URL                   = BASE_URL_OF_12306 + "/otn/confirmPassenger/initDc" //生成订单时需要先请求这个页面
	API_GET_QUEUE_COUNT_URL          = BASE_URL_OF_12306 + "/otn/confirmPassenger/getQueueCount"
	API_CONFIRM_SINGLE_FOR_QUEUE_URL = BASE_URL_OF_12306 + "/otn/confirmPassenger/confirmSingleForQueue"
	API_QUERY_ORDER_WAIT_TIME_URL    = BASE_URL_OF_12306 + "/otn/confirmPassenger/queryOrderWaitTime?{}" // 排队查询
	API_QUERY_INIT_PAGE_URL          = BASE_URL_OF_12306 + "/otn/leftTicket/init"
	API_CHECH_QUEUE_TICKET_URL       = BASE_URL_OF_12306 + "/otn/afterNate/chechFace"          //候补验证
	API_SUBMIT_QUEUE_TICKET_URL      = BASE_URL_OF_12306 + "/otn/afterNate/submitOrderRequest" //候补提交
	API_QUEUE_SUCCESS_RATE_URL       = BASE_URL_OF_12306 + "/otn/afterNate/getSuccessRate"     //候补成功率查询
	//API_GET_BROWSER_DEVICE_ID_URL    = BASE_URL_OF_12306 + "/otn/HttpZF/logdevice"
	API_GET_BROWSER_DEVICE_ID_URL              = "https://12306-rail-id-v2.pjialin.com"
	API_FREE_CODE_QCR_API_URL                  = "http://127.0.0.1:8009/check"
	API_NOTIFICATION_BY_VOICE_CODE_URL         = "http://ali-voice.showapi.com/sendVoice?"
	API_NOTIFICATION_BY_VOICE_CODE_DINGXIN_URL = "http://yuyin2.market.alicloudapi.com/dx/voice_notice"
	API_CHECK_CDN_AVAILABLE                    = "https://%s/otn/dynamicJs/omseuuq"
)
