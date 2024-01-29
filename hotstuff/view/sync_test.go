package view

//
//import (
//	"context"
//	"distributed/hotstuff/pb"
//	"reflect"
//	"sync"
//	"testing"
//	"time"
//)
//
//func TestNew(t *testing.T) {
//	type args struct {
//		viewDuration ViewDuration
//	}
//	tests := []struct {
//		name string
//		args args
//		want *Synchronize
//	}{
//		{
//			name: "test1",
//			args: args{
//				NewViewDuration(float64(time.Second*20), 1),
//			},
//			want: &Synchronize{},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := New(tt.args.viewDuration); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("New() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestNewViewDuration(t *testing.T) {
//	type args struct {
//		maxTimeout float64
//		multiplier float64
//	}
//	tests := []struct {
//		name string
//		args args
//		want ViewDuration
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := NewViewDuration(tt.args.maxTimeout, tt.args.multiplier); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("NewViewDuration() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestSynchronize_GetLeader(t *testing.T) {
//	type fields struct {
//		mut         sync.Mutex
//		currentView int64
//		highQC      *pb.QC
//		highTC      *pb.QC
//		duration    ViewDuration
//		timer       *time.Timer
//		timeouts    map[int64]map[int32]*pb.TimeoutMsg
//	}
//	type args struct {
//		viewnumber int64
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   int32
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Synchronize{
//				mut:         tt.fields.mut,
//				currentView: tt.fields.currentView,
//				highQC:      tt.fields.highQC,
//				highTC:      tt.fields.highTC,
//				duration:    tt.fields.duration,
//				timer:       tt.fields.timer,
//				timeouts:    tt.fields.timeouts,
//			}
//			if got := s.GetLeader(tt.args.viewnumber); got != tt.want {
//				t.Errorf("GetLeader() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestSynchronize_Start(t *testing.T) {
//	type fields struct {
//		mut         sync.Mutex
//		currentView int64
//		highQC      *pb.QC
//		highTC      *pb.QC
//		duration    ViewDuration
//		timer       *time.Timer
//		timeouts    map[int64]map[int32]*pb.TimeoutMsg
//	}
//	type args struct {
//		ctx context.Context
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Synchronize{
//				mut:         tt.fields.mut,
//				currentView: tt.fields.currentView,
//				highQC:      tt.fields.highQC,
//				highTC:      tt.fields.highTC,
//				duration:    tt.fields.duration,
//				timer:       tt.fields.timer,
//				timeouts:    tt.fields.timeouts,
//			}
//			s.Start(tt.args.ctx)
//		})
//	}
//}
//
//func TestSynchronize_startTimeOutTimer(t *testing.T) {
//	type fields struct {
//		mut         sync.Mutex
//		currentView int64
//		highQC      *pb.QC
//		highTC      *pb.QC
//		duration    ViewDuration
//		timer       *time.Timer
//		timeouts    map[int64]map[int32]*pb.TimeoutMsg
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Synchronize{
//				mut:         tt.fields.mut,
//				currentView: tt.fields.currentView,
//				highQC:      tt.fields.highQC,
//				highTC:      tt.fields.highTC,
//				duration:    tt.fields.duration,
//				timer:       tt.fields.timer,
//				timeouts:    tt.fields.timeouts,
//			}
//			s.startTimeOutTimer()
//		})
//	}
//}
//
//func Test_viewDuration_Duration(t *testing.T) {
//	type fields struct {
//		startTime  time.Time
//		max        float64
//		timeoutMul float64
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   time.Duration
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &viewDuration{
//				startTime:  tt.fields.startTime,
//				max:        tt.fields.max,
//				timeoutMul: tt.fields.timeoutMul,
//			}
//			if got := v.Duration(); got != tt.want {
//				t.Errorf("Duration() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_viewDuration_ViewStarted(t *testing.T) {
//	type fields struct {
//		startTime  time.Time
//		max        float64
//		timeoutMul float64
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &viewDuration{
//				startTime:  tt.fields.startTime,
//				max:        tt.fields.max,
//				timeoutMul: tt.fields.timeoutMul,
//			}
//			v.ViewStarted()
//		})
//	}
//}
//
//func Test_viewDuration_ViewSucceeded(t *testing.T) {
//	type fields struct {
//		startTime  time.Time
//		max        float64
//		timeoutMul float64
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &viewDuration{
//				startTime:  tt.fields.startTime,
//				max:        tt.fields.max,
//				timeoutMul: tt.fields.timeoutMul,
//			}
//			v.ViewSucceeded()
//		})
//	}
//}
//
//func Test_viewDuration_ViewTimeout(t *testing.T) {
//	type fields struct {
//		startTime  time.Time
//		max        float64
//		timeoutMul float64
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &viewDuration{
//				startTime:  tt.fields.startTime,
//				max:        tt.fields.max,
//				timeoutMul: tt.fields.timeoutMul,
//			}
//			v.ViewTimeout()
//		})
//	}
//}
