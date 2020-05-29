package slack

import (
	"context"
	"testing"
)

func TestNotification(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Notification(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Notification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
