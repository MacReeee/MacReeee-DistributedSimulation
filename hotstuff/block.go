package hotstuff

import (
	"distributed/hotstuff/pb"
	"sync"
)

type Block struct {
	block    *pb.Block // 区块结构体，包含了区块的各个属性
	proposer int32     // 提议者的ID
	children []*Block  // 子区块列表
}

var GenesisBlock = &Block{
	block: &pb.Block{
		Hash:       []byte("000000000000"),         // 区块哈希
		ParentHash: []byte("FFFFFFFFFFFF"),         // 父区块哈希
		Height:     0,                              // 区块高度
		ViewNumber: 0,                              // 视图编号
		Qc:         nil,                            // 区块的QC（Quorum Certificate）
		Cmd:        []byte("Create Genesis Block"), // 区块的命令
	},
	proposer: 0,   // 提议者的ID
	children: nil, // 子区块列表
}

type blockchain struct {
	mut           sync.Mutex        // 互斥锁，用于保护区块链的并发访问
	pruneHeight   int64             // 剪枝高度
	blocks        map[string]*Block // 存储所有区块的映射表，以区块哈希为键
	blockAtHeight map[int64]*Block  // 存储每个高度的区块的映射表，以区块高度为键
	// pendingFetch  map[hotstuff.Hash]context.CancelFunc
}

func NewBlockChain() *blockchain {
	blockchain := &blockchain{
		blocks:        make(map[string]*Block), // 初始化区块映射表
		blockAtHeight: make(map[int64]*Block),  // 初始化高度映射表
	}
	blockchain.Store(GenesisBlock) // 存储创世区块
	return blockchain
}

func (bc *blockchain) Store(block *Block) {
	bc.mut.Lock()                                // 加锁
	defer bc.mut.Unlock()                        // 解锁
	bc.blocks[string(block.block.Hash)] = block  // 将区块存储到区块映射表中
	bc.blockAtHeight[block.block.Height] = block // 将区块存储到高度映射表中
	//存储父区块的children字段
	bc.blocks[string(block.block.ParentHash)].children = append(bc.blocks[string(block.block.ParentHash)].children, block)
}
