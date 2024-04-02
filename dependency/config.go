package dependency

import (
	"time"
)

var (
	DebugMode         = true //调试模式开关
	NumReplicas int32 = 2    //副本数量
	ReplicaID   int32        //副本ID
)

type network struct {
	// 注意分布是指数分布
	Latency     time.Duration //包含区块的传输延迟均值
	ProcessTime time.Duration //投票和不含区块的处理和传输时延
}

func GetLatency() time.Duration {
	latency := time.Duration(GenerateExpRand(float64(Configs.Network.Latency)))
	return latency * time.Millisecond
}

func GetProcessTime() time.Duration {
	ProcessTime := time.Duration(GenerateExpRand(float64(Configs.Network.ProcessTime)))
	return ProcessTime * time.Millisecond
}

func GenerateExpRand(lambda float64) float64 {
	return r.ExpFloat64() / lambda
}
