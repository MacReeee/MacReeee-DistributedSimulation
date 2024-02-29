package main

import (
	"distributed/hotstuff/blockchain"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/cryp"
	"distributed/hotstuff/view"
	"flag"
	"log"
)

func main() { //此主函数用于启动服务端
	id := int32(*flag.Int("id", 1, "replica id"))
	flag.Parse()
	blockchain.NewBlockChain()
	view.New()
	cryp.NewSignerByID(id)
	server, listener := hotstuff.NewReplicaServer(id)
	log.Println("副本", id, "启动成功！")
	go hotstuff.Debug_Period_Out()
	server.Serve(*listener)
}
