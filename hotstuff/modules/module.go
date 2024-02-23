package modules

import (
	"distributed/hotstuff/middleware"
	"distributed/hotstuff/pb"

	"google.golang.org/grpc"
)

type modules struct {
	Chain             middleware.Chain
	Synchronizer      middleware.Synchronizer
	SignerAndVerifier middleware.Crypto
	ReplicaServer     *grpc.Server
	ReplicaClient     map[int32]*pb.HotstuffClient
	// ReplicaPubKey     map[int32]kyber.Point
}

var MODULES = &modules{}
