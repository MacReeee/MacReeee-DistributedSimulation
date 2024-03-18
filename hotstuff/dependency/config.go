package dependency

import "time"

var (
	DebugMode = true //调试模式开关
)

var (
	BlockSize   = 1e5                                                   //单位：比特
	Speed       = 0.5e6                                                 //传输速度
	Latency     = time.Millisecond * time.Duration(BlockSize/Speed) * 0 //包含区块的传输延迟
	ProcessTime = time.Millisecond * 0                                  //投票和不含区块的处理和传输时延

	Network = &network{}
)

type network struct {
	BlockSize   float64       //单位：比特
	Speed       float64       //传输速度
	Latency     time.Duration //包含区块的传输延迟
	ProcessTime time.Duration //投票和不含区块的处理和传输时延
}

func (n *network) GetLatency() time.Duration {
	return time.Millisecond * time.Duration(n.BlockSize/n.Speed)
}

func (n *network) GetProcessTime() time.Duration {
	return n.ProcessTime
}

func (n *network) LoadFromFile() {
	n.BlockSize = BlockSize
	n.Speed = Speed
	n.Latency = Latency
	n.ProcessTime = ProcessTime
}
