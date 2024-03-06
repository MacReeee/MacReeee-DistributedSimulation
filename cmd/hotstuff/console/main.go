package main

import hotstuff "distributed/hotstuff/consensus"

func ConnectToOthers() {
	for i := 0; i < 4; i++ {
		if i == int(hotstuff.ReplicaID) {
			continue
		}
		hotstuff.NewReplicaClient(int32(i + 1))
	}
}
