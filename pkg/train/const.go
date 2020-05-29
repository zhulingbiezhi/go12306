package train

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
