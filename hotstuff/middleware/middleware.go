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

type Block struct {
	Block    *pb.Block // 区块结构体，包含了区块的各个属性
	Proposer int32     // 提议者的ID
	Children []*Block  // 子区块列表
}

type Chain interface {
	CreateBlock(ParentHash []byte, Height int64, ViewNumber int64, QC *pb.QC, Cmd []byte) *Block
	GetBlock(hash []byte) *Block
	PruneBlock(block *Block, NewestChild *Block) []string
	Store(block *Block)
}
