package middleware

import (
	d "distributed/dependency"
	pb2 "distributed/pb"
	"sync"

	"go.dedis.ch/kyber/v3"
)

type Synchronizer interface {
	GetLeader(viewnumber ...int64) int32
	Start()
	// StartTimeOutTimer(ctx context.Context, timeout context.CancelFunc)
	TimerReset() bool
	Success()
	GetOnly() *sync.Once //保证视图成功函数只执行一次

	//元数据
	ViewNumber() *int64
	Timeout() <-chan bool
	StoreVote(msgType pb2.MsgType, NormalMsg *pb2.VoteRequest, NewViewMsg ...*pb2.NewViewMsg) int
	GetVoter(msgType pb2.MsgType) ([]int32, [][]byte, *d.OnceWithDone) // 返回投票者、投票信息、对应的once
	GetOnce(megType pb2.MsgType) *d.OnceWithDone
	HighQC() *pb2.QC
	MU() *sync.Mutex
	ViewNumberPP()
	ViewNumberSet(v int64)

	//no use
	QC(msgType pb2.MsgType) *pb2.QC //合成一个QC

	//Debug
	Debug()
}

type Chain interface {
	CreateBlock(ParentHash []byte, ViewNumber int64, QC *pb2.QC, Cmd []byte, Proposer int32) *pb2.Block
	GetBlock(hash []byte) *pb2.Block
	PruneBlock(block *pb2.Block, NewestChild *pb2.Block) []string
	WriteToFile(NewestChild *pb2.Block)
	Store(block *pb2.Block)
	StoreToTemp(block *pb2.Block)
	GetBlockFromTemp(hash []byte) *pb2.Block

	//Debug
	GetBlockChain() (map[string]*pb2.Block, map[int64]*pb2.Block)
}
type CRYP interface {
	Sign(msgType pb2.MsgType, viewnumber int64, BlockHash []byte) ([]byte, error) // 用于对投票信息 "${消息类型},${视图号},${区块hash}" 进行签名
	Verify(voter int32, msg []byte, sig []byte) bool                              // 用于对投票的验证
	ThreshMock(voter []int32, sigs [][]byte) ([]byte, []byte, error)              // 门限签名的模拟实现
	ThreshVerifyMock(QC *pb2.QC) bool                                             // 门限签名验证的模拟实现
	NormSign(msg []byte) ([]byte, error)                                          //普通消息的签名
}

// Deprecated: Use CRYP instead
type Crypto interface {
	PartSign(msg []byte) ([]byte, error)
	Sign(msg []byte) ([]byte, error)
	ThresholdSign(msg []byte, SigMap map[kyber.Point][]byte) ([]byte, kyber.Point, error)
	ThreshVerify(msg []byte, sig []byte, pubKey kyber.Point) bool
	Verify(msg []byte, sig []byte) bool
}

type Server interface {
	SelfID() int32
}
