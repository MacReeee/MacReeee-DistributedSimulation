package main

import (
	"distributed/blockchain"
	"distributed/consensus"
	"distributed/cryp"
	d "distributed/dependency"
	"distributed/view"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
)

func main() { //此主函数用于启动服务端

	//分析堆栈时取消注释
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	idptr := flag.Int("id", 1, "replica id")
	flag.Parse()
	id := int32(*idptr)
	d.ReplicaID = id
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	if os.Remove("./output/replica_"+strconv.Itoa(int(id))+"_committed.json") != nil {
		log.Println("文件已删除")
	}

	blockchain.NewBlockChain()
	view.NewSync()
	cryp.NewSignerByID(id)

	server, listener := hotstuff.NewReplicaServer(id)
	d.LoadFromFile() //加载配置文件
	log.Println("副本", id, "启动成功！")
	//go hotstuff.Debug_Period_Out()
	server.Serve(*listener)
}
