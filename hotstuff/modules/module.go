package modules

type modules struct {
	Chain             any
	Synchronizer      any
	SignerAndVerifier any
	ReplicaServer     any
	ReplicaClient     map[int32]any
}

var MODULES = &modules{}
