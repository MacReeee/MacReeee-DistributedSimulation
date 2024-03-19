package main

import (
	hotstuff "distributed/hotstuff/consensus"
	d "distributed/hotstuff/dependency"
)

func main() {
	for i := 0; i < 4; i++ {
		if i == int(d.ReplicaID) {
			continue
		}
		hotstuff.NewReplicaClient(int32(i + 1))
	}

}
