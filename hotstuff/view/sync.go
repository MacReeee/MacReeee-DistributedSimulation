package view

import (
	"distributed/hotstuff/pb"
	"sync"
	"time"
)

type Synchronize struct {
	mut         sync.Mutex
	currentView int64
	highQC      *pb.QC //开启一个视图需要一个HighQC
	highTC      *pb.QC
	duration    ViewDuration
	timer       *time.Timer
	timeouts    map[int64]map[int32]*pb.TimeoutMsg
}
