package dependency

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

// 环境变量
var (
	r = rand.New(rand.NewSource(time.Now().UnixNano())) //随机数生成器

)

type Config struct {
	BuildInfo Buildinfo `json:"BuildInfo"`
	Network   network   `json:"Network"`
}

func LoadFromFile() {
	var filename string
	if Configs.BuildInfo.DebugMode {
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
	//检查自身ID是否在拒绝投票列表中
	for _, id := range Configs.BuildInfo.DenyVote {
		if id == ReplicaID {
			DenyVote = true
			break
		}
	}

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
