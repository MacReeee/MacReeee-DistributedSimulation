package modules

import (
	"distributed/middleware"
	"distributed/pb"
	"encoding/json"

	"google.golang.org/grpc"
)

type modules struct {
	Chain               middleware.Chain
	Synchronizer        middleware.Synchronizer
	Signer              middleware.CRYP
	ReplicaServer       *grpc.Server
	ReplicaServerStruct middleware.Server
	ReplicaClient       map[int32]*pb.HotstuffClient
	Reset               chan struct{} //复位信号

	// Deprecated: Use CRYP instead
	SignerAndVerifier middleware.Crypto
}

var MODULES = &modules{
	ReplicaClient: make(map[int32]*pb.HotstuffClient),
}

func (m *modules) MarshalToJSON() ([]byte, error) {
	return json.Marshal(m)
}
