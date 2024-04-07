package main

import (
	"distributed/blockchain"
	"distributed/consensus"
	"distributed/cryp"
	d "distributed/dependency"
	"distributed/modules"
	"distributed/view"
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"strconv"
)

func main() { //此主函数用于启动服务端

	//分析堆栈时取消注释
	//go func() {
	//	http.ListenAndServe("localhost:6060", nil)
	//}()

	//dir, err := os.Getwd()
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}
	//fmt.Println("Current directory:", dir)

	idptr := flag.Int("id", 1, "replica id")
	flag.Parse()
	id := int32(*idptr)
	d.ReplicaID = id
	d.LoadFromFile() //加载配置文件
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	if os.Remove("./output/replica_"+strconv.Itoa(int(id))+"_committed.json") != nil {
		log.Println("文件已删除")
	}

	for {
		blockchain.NewBlockChain()
		view.NewSync()
		cryp.NewSignerByID(id)

		server, listener := hotstuff.NewReplicaServer(id)
		log.Println("副本", id, "启动成功！")
		go server.Serve(*listener)

		<-modules.MODULES.Reset //接收到复位信号以后清理
	}
}
