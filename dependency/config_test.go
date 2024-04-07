package dependency

import (
	"encoding/json"
	"os"
	"testing"
)

type test_struct struct {
	A int
	B string
	C float64
}

func Test_ToFile(t *testing.T) {
	file, err := os.Create("config.json")
	if err != nil {
		t.Error("创建文件失败")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	config := Config{
		Buildinfo{
			DebugMode:   true,
			DockerMode:  false,
			Threshold:   6,
			NumReplicas: 10,
		},
		network{
			Latency:     100,
			ProcessTime: 20,
		},
	}

	err = encoder.Encode(config)
	if err != nil {
		t.Error("写入文件失败")
	}
}
