package modules

import (
	"distributed/hotstuff/middleware"
	"distributed/hotstuff/pb"

	"google.golang.org/grpc"
)

type modules struct {
	Chain         middleware.Chain
	Synchronizer  middleware.Synchronizer
	Signer        middleware.CRYP
	ReplicaServer *grpc.Server
	ReplicaClient map[int32]*pb.HotstuffClient

	// Deprecated: Use CRYP instead
	SignerAndVerifier middleware.Crypto

	// ReplicaPubKey     map[int32]kyber.Point
}

var MODULES = &modules{}
