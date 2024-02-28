package main

import (
	"distributed/hotstuff/blockchain"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/cryp"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/view"
	"flag"
	"fmt"
)

func main() { //此主函数用于启动服务端
	id := int32(*flag.Int("id", 1, "replica id"))
	flag.Parse()

	modules := modules.MODULES
	blockchain.NewBlockChain()
	view.New()
	cryp.NewSignerByID(id)
	hotstuff.NewReplicaServer(id)
	fmt.Println(modules)
}
