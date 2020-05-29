package query

import (
	"testing"

	"github.com/zhulingbiezhi/go12306/pkg/helper"
)

func TestQueryLeftTicket11(t *testing.T) {
	type args struct {
		request *QueryLeftTicketRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test query",
			args: args{
				request: &QueryLeftTicketRequest{
					FromStation: helper.StationMap["深圳"].Key,
					ToStation:   helper.StationMap["吉安"].Key,
					TrainDate:   "2020-01-22",
					PurposeCode: helper.PurposeTypeAdult,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := QueryLeftTicket(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryLeftTicket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
