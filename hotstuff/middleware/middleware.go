package middleware

import (
	"context"
	"distributed/hotstuff/pb"

	"go.dedis.ch/kyber/v3"
)

type Synchronizer interface {
	GetLeader(viewnumber ...int64) int32
	Start(ctx context.Context)
	StartTimeOutTimer(ctx context.Context, timeout context.CancelFunc)
	TimerReset() bool
	GetContext() (context.Context, context.CancelFunc)

	//元数据
	ViewNumber() int64
	Timeout() <-chan bool
	StoreVote(msgType pb.MsgType, NormalMsg *pb.VoteRequest, NewViewMsg ...*pb.NewViewMsg)
	GetVoter(msgType pb.MsgType) ([]int32, [][]byte)
	HighQC() *pb.QC
	QC(msgType pb.MsgType) *pb.QC //合成一个QC
}

type Chain interface {
	CreateBlock(ParentHash []byte, ViewNumber int64, QC *pb.QC, Cmd []byte, Proposer int32) *pb.Block
	GetBlock(hash []byte) *pb.Block
	PruneBlock(block *pb.Block, NewestChild *pb.Block) []string
	Store(block *pb.Block)
	StoreToTemp(block *pb.Block)
	GetBlockFromTemp(hash []byte) *pb.Block
}
type CRYP interface {
	Sign(msgType pb.MsgType, viewnumber int64, BlockHash []byte) ([]byte, error) // 用于对投票信息 "${消息类型},${视图号},${区块hash}" 进行签名
	Verify(voter int32, msg []byte, sig []byte) bool                             // 用于对投票的验证
	ThreshMock(voter []int32, sigs [][]byte) ([]byte, []byte, error)             // 门限签名的模拟实现
	ThreshVerifyMock(QC *pb.QC) bool                                             // 门限签名验证的模拟实现
	NormSign(msg []byte) ([]byte, error)                                         //普通消息的签名
}

// Deprecated: Use CRYP instead
type Crypto interface {
	PartSign(msg []byte) ([]byte, error)
	Sign(msg []byte) ([]byte, error)
	ThresholdSign(msg []byte, SigMap map[kyber.Point][]byte) ([]byte, kyber.Point, error)
	ThreshVerify(msg []byte, sig []byte, pubKey kyber.Point) bool
	Verify(msg []byte, sig []byte) bool
}
