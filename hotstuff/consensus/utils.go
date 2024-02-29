package hotstuff

import (
	"distributed/hotstuff/middleware"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func MatchingMsg(Type pb.MsgType, ViewNumber int64, TarType pb.MsgType, TarviewNumber int64) (bool, error) {
	condition1 := (Type == TarType)
	if !condition1 {
		return false, fmt.Errorf("消息类型不匹配")
	}
	condition2 := (ViewNumber == TarviewNumber)
	if !condition2 {
		return false, fmt.Errorf("视图号不匹配")
	}
	//log.Println("MatchingMsg:", condition1, condition2)
	return condition1 && condition2, nil
}

func Sign(msg []byte) []byte {
	return nil
}

func ViewSuccess(sync middleware.Synchronizer) {
	_, success := sync.GetContext()
	success()
}

func QCMarshal(qc *pb.QC) []byte {
	qcjson, err := json.Marshal(qc)
	if err != nil {
		log.Println("json序列化失败:", err)
	}
	return qcjson
}

func Debug_Period_Out() {
	var (
		sync = modules.MODULES.Synchronizer
		//cryp  = modules.MODULES.Signer
		//chain = modules.MODULES.Chain
	)
	for {
		log.Println("当前视图: ", sync.ViewNumber())

		time.Sleep(10 * time.Second)
		fmt.Printf("\n\n")
	}
}
