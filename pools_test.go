package timerpools

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerPool_Subscribe(t *testing.T) {
	type args struct {
		delay    time.Duration
		callback func(ctx *TimeContext)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test case 1",
			args: args{
				delay: 1 * time.Second,
				callback: func(ctx *TimeContext) {
					fmt.Println(`pool:`, ctx.Pid, "timerId:", ctx.TimerId, " now:", ctx.Now, " delay:", ctx.Delay)
				},
			},
		},
		{
			name: "test case 2",
			args: args{
				delay: 2 * time.Second,
				callback: func(ctx *TimeContext) {
					fmt.Println(`pool:`, ctx.Pid, "timerId:", ctx.TimerId, " now:", ctx.Now, " delay:", ctx.Delay)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewTimerPool()
			p.Subscribe(tt.args.delay, tt.args.callback)
		})
	}
	time.Sleep(5 * time.Second)
}
