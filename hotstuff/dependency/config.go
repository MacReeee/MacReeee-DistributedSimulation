package dependency

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	DebugMode         = true //调试模式开关
	NumReplicas int32 = 4    //副本数量
)

var (
	Configs   = Config{}
	ReplicaID int32 //副本ID
)

type Config struct {
	Network network `json:"Network"`
}

type network struct {
	Latency     time.Duration //包含区块的传输延迟
	ProcessTime time.Duration //投票和不含区块的处理和传输时延
}

func LoadFromFile() {
	var filename string
	if DebugMode {
		filename = "config_debug.json"
	} else {
		filename = "config.json"
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Println("打开配置文件失败:", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("关闭文件失败:", err)
		}
	}(file)

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Configs)

	log.Println("配置文件加载成功:")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(Configs)
}

func ReadConfig() {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	encoder.Encode(Configs)
}

func GetLatency() time.Duration {
	//return time.Second * time.Duration(Configs.Network.BlockSize/Configs.Network.Speed)
	return Configs.Network.Latency * time.Millisecond
}

func GetProcessTime() time.Duration {
	return Configs.Network.ProcessTime * time.Millisecond
}
