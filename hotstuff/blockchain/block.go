package blockchain

import (
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"sync"
)

var GenesisBlock = &pb.Block{
	Hash:       []byte("000000000000"),         // 区块哈希
	ParentHash: []byte("FFFFFFFFFFFF"),         // 父区块哈希
	Height:     0,                              // 区块高度
	ViewNumber: 0,                              // 视图编号
	Qc:         nil,                            // 区块的QC（Quorum Certificate）
	Cmd:        []byte("Create Genesis Block"), // 区块的命令
	Proposer:   0,                              // 提议者的ID
	Children:   nil,                            // 子区块列表
}

var TempBlockMap = make(map[string]*pb.Block)

type Blockchain struct {
	Mut           sync.Mutex           // 互斥锁，用于保护区块链的并发访问
	PruneHeight   int64                // 剪枝高度
	Blocks        map[string]*pb.Block // 存储所有区块的映射表，以区块哈希为键
	BlockAtHeight map[int64]*pb.Block  // 存储每个高度的区块的映射表，以区块高度为键
	// pendingFetch  map[hotstuff.Hash]context.CancelFunc
}

func NewBlockChain() *Blockchain {
	blockchain := &Blockchain{
		Blocks:        make(map[string]*pb.Block), // 初始化区块映射表
		BlockAtHeight: make(map[int64]*pb.Block),  // 初始化高度映射表
	}
	blockchain.Store(GenesisBlock) // 存储创世区块
	modules.MODULES.Chain = blockchain
	return blockchain
}

// 存储区快
func (bc *Blockchain) Store(block *pb.Block) {
	bc.Mut.Lock()                          // 加锁
	defer bc.Mut.Unlock()                  // 解锁
	bc.Blocks[string(block.Hash)] = block  // 将区块存储到区块映射表中
	bc.BlockAtHeight[block.Height] = block // 将区块存储到高度映射表中
	//存储父区块的children字段
	bc.Blocks[string(block.ParentHash)].Children = append(bc.Blocks[string(block.ParentHash)].Children, string(block.Hash))
}

func (bc *Blockchain) StoreToTemp(block *pb.Block) {
	bc.Mut.Lock() // 加锁
	defer bc.Mut.Unlock()
	TempBlockMap[string(block.Hash)] = block
}

// 给定区块的哈希，查找对应的区块
func (bc *Blockchain) GetBlock(hash []byte) *pb.Block {
	return bc.Blocks[string(hash)]
}

func (bc *Blockchain) GetBlockFromTemp(hash []byte) *pb.Block {
	return TempBlockMap[string(hash)]
}

// 剪枝
// todo 此处可以检查被剪枝的区块来检测分叉，不做
// 被提交的最新区块的上一个区块才需要剪枝，因此在提交区块的时候记得调用剪枝
func (chain *Blockchain) PruneBlock(block *pb.Block, NewestChild *pb.Block) []string {
	var deleted []string
	for _, child := range block.Children {
		if child == string(NewestChild.Hash) {
			continue
		}
		deleted = append(deleted, child)
		delete(chain.Blocks, child)
	}
	return deleted
}

func (chain *Blockchain) CreateBlock(ParentHash []byte, Height int64, ViewNumber int64, QC *pb.QC, Cmd []byte) *pb.Block {
	block := &pb.Block{
		Hash:       nil,
		ParentHash: ParentHash,
		Height:     Height,
		ViewNumber: ViewNumber,
		Qc:         QC,
		Cmd:        Cmd,
		Proposer:   0,
		Children:   nil,
	}
	// todo: 在这里需要有某一个模块给出自身id，赋值给Proposer
	return block
}
