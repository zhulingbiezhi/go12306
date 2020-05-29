package cookie

import (
	"context"
	"testing"
)

func TestGetRailDevice(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test get device id ",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetRailDevice(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("GetRailDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
