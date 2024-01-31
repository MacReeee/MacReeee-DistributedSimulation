package middleware

import (
	"context"
	"distributed/hotstuff/pb"

	"go.dedis.ch/kyber/v3"
)

type Synchronizer interface {
	GetLeader(viewnumber ...int64) int32
	Start(ctx context.Context)
	StartTimeOutTimer()
	TimerReset() bool
	GetContext() context.Context
}

type Crypto interface {
	PartSign(msg []byte) ([]byte, error)
	Sign(msg []byte) ([]byte, error)
	ThreshVerify(msg []byte, sig []byte, pubKey kyber.Point) bool
	Verify(msg []byte, sig []byte) bool
}

type Chain interface {
	CreateBlock(ParentHash []byte, Height int64, ViewNumber int64, QC *pb.QC, Cmd []byte) *pb.Block
	GetBlock(hash []byte) *pb.Block
	PruneBlock(block *pb.Block, NewestChild *pb.Block) []string
	Store(block *pb.Block)
	StoreTemp(block *pb.Block)
}
