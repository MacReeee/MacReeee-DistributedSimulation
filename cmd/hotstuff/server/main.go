package main

import (
	"distributed/hotstuff/blockchain"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/crypto"
	"distributed/hotstuff/view"
)

func main() {
	blockchain.NewBlockChain()
	view.New()
	crypto.NewSignerAndVerifier(hotstuff.PrivateKey, hotstuff.PublicKey)
	hotstuff.NewReplicaServer()
}
