package dependency

import (
	"time"
)

var (
	ReplicaID int32     //副本ID
	Configs   = Config{ //配置文件
		BuildInfo: Buildinfo{
			DebugMode:   true,
			DockerMode:  false,
			Threshold:   0,
			NumReplicas: 0,
		},
	}
	DenyVote = false //是否拒绝投票
)

type Buildinfo struct {
	DebugMode   bool    //调试模式开关
	DockerMode  bool    //Docker模式开关
	Threshold   int     //调试模式阈值
	NumReplicas int32   //副本数量
	DenyVote    []int32 //拒绝投票的副本id
}

type network struct {
	// 注意分布是指数分布
	Latency      time.Duration //包含区块的传输延迟均值
	ProcessTime  time.Duration //投票和不含区块的处理和传输时延
	BASE_Timeout time.Duration //基础超时时间
	MAX_Timeout  time.Duration //最大超时时间
}

func GetLatency() time.Duration {
	configLatency := Configs.Network.Latency * 1000
	random := GenerateExpRand(float64(configLatency))
	latency := time.Duration(random)
	//fmt.Printf("生成的随机数: %v\n", random)
	return latency * time.Microsecond
}

func GetProcessTime() time.Duration {
	configProcessTime := Configs.Network.ProcessTime * 1000000
	random := GenerateExpRand(float64(configProcessTime))
	processTime := time.Duration(random)
	return processTime * time.Nanosecond
}

func GetBASE_Timeout() time.Duration {
	return Configs.Network.BASE_Timeout * time.Millisecond
}

func GetMAX_Timeout() time.Duration {
	return Configs.Network.MAX_Timeout * time.Millisecond
}

func GenerateExpRand(E float64) float64 {
	lambda := 1 / E
	return r.ExpFloat64() / lambda
}
