package blockchain

import (
	"crypto"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

var GenesisBlock = &pb.Block{
	Hash:       []byte("FFFFFFFFFFFF"),         // 区块哈希，对Cmd取哈希
	ParentHash: []byte("000000000000"),         // 父区块哈希
	Height:     0,                              // 区块高度
	ViewNumber: 0,                              // 视图编号
	Qc:         nil,                            // 区块的QC（Quorum Certificate）
	Cmd:        []byte("Create Genesis Block"), // 区块的命令
	Proposer:   0,                              // 提议者的ID
	Children:   make([]string, 0),              // 子区块列表
}

var TempBlockMap = make(map[string]*pb.Block)

type Blockchain struct {
	Mut sync.Mutex // 互斥锁，用于保护区块链的并发访问
	// PruneHeight   int64                // 剪枝高度
	Blocks        map[string]*pb.Block // 存储所有区块的映射表，以区块哈希为键
	BlockAtHeight map[int64]*pb.Block  // 存储每个高度的区块的映射表，以区块高度为键
	// pendingFetch  map[hotstuff.Hash]context.CancelFunc
	curHeight int64
}

func NewBlockChain() *Blockchain {
	blockchain := &Blockchain{
		Mut:           sync.Mutex{},
		Blocks:        make(map[string]*pb.Block), // 初始化区块映射表
		BlockAtHeight: make(map[int64]*pb.Block),  // 初始化高度映射表
		curHeight:     0,                          // 初始化区块链的高度
	}
	blockchain.Store(GenesisBlock) // 存储创世区块
	modules.MODULES.Chain = blockchain
	return blockchain
}

// 存储区快
func (bc *Blockchain) Store(block *pb.Block) {
	log.Println(string(block.Hash), string(block.ParentHash))
	bc.Mut.Lock()                          // 加锁
	bc.Blocks[string(block.Hash)] = block  // 将区块存储到哈希映射表中
	bc.BlockAtHeight[block.Height] = block // 将区块存储到高度映射表中
	//存储父区块的children字段
	if string(block.ParentHash) != "000000000000" {
		parent := bc.Blocks[string(block.ParentHash)]
		parent.Children = append(parent.Children, string(block.Hash))
	}
	bc.Mut.Unlock() // 解锁
}

func (bc *Blockchain) StoreToTemp(block *pb.Block) {
	bc.Mut.Lock() // 加锁
	defer bc.Mut.Unlock()
	TempBlockMap[string(block.Hash)] = block
}

// 给定区块的哈希，查找对应的区块
func (bc *Blockchain) GetBlock(hash []byte) *pb.Block {
	block := bc.Blocks[string(hash)]
	return block
}

func (bc *Blockchain) GetBlockFromTemp(hash []byte) *pb.Block {
	return TempBlockMap[string(hash)]
}

// 剪枝
// 此处可以检查被剪枝的区块来检测分叉，不做
// 被提交的最新区块的上一个区块才需要剪枝，因此在提交区块的时候记得调用剪枝
// 第一个参数是需要剪枝的区块，第二个参数是被提交的最新区块
func (chain *Blockchain) PruneBlock(block *pb.Block, NewestChild *pb.Block) []string {
	var deleted []string
	for _, child := range block.Children {
		// 如果子区块是最新区块的哈希，则跳过，不进行剪枝
		if child == string(NewestChild.Hash) {
			continue
		}
		// 将需要剪枝的子区块添加到删除列表中
		deleted = append(deleted, child)
		// 从区块链的映射表中删除该子区块
		delete(chain.BlockAtHeight, chain.Blocks[child].Height)
		delete(chain.Blocks, child)
		// 删除被剪枝区块中子区块字段对应key
		Children := chain.Blocks[string(block.Hash)].Children
		for i, v := range chain.Blocks[string(block.Hash)].Children {
			if v == child {
				chain.Blocks[string(block.Hash)].Children[i] = chain.Blocks[string(block.Hash)].Children[len(Children)-1]
				chain.Blocks[string(block.Hash)].Children = chain.Blocks[string(block.Hash)].Children[0 : len(Children)-1]
			}
		}
	}
	// 将NewestChild转化为json格式并存储到committed_blocks文件中
	type Data struct {
		ParentHash string
		Hash       string
		Height     int64
		CMD        string
		ViewNumber int64
		Proposer   int32
		Children   []string
	}
	//将NewestChild转化成Data类型
	data := Data{
		ParentHash: string(NewestChild.ParentHash),
		Hash:       string(NewestChild.Hash),
		Height:     NewestChild.Height,
		CMD:        string(NewestChild.Cmd),
		ViewNumber: NewestChild.ViewNumber,
		Proposer:   NewestChild.Proposer,
		Children:   NewestChild.Children,
	}
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Println("json转换失败:", err)
	}
	// 分不同节点不同文件存储
	file, err := os.OpenFile("./committed_blocks"+strconv.Itoa(1)+".json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("文件打开失败:", err)
	}
	defer file.Close()
	_, err = file.Write(jsonData)
	if err != nil {
		log.Println("文件写入失败:", err)
	}
	_, err = file.WriteString(",\n")
	if err != nil {
		log.Println("文件写入失败:", err)
	}
	// 返回被删除的子区块列表
	return deleted
}

func (chain *Blockchain) CreateBlock(ParentHash []byte, ViewNumber int64, QC *pb.QC, Cmd []byte, Proposer int32) *pb.Block {
	chain.curHeight++
	hasher := crypto.SHA256.New()
	hasher.Write(Cmd) //区块的哈希是Cmd的哈希
	hash := []byte(fmt.Sprintf("%x", hasher.Sum(nil)))
	block := &pb.Block{
		Hash:       hash,
		ParentHash: ParentHash,
		Height:     chain.curHeight,
		ViewNumber: ViewNumber,
		Qc:         QC,
		Cmd:        Cmd,
		Proposer:   Proposer,
		Children:   []string{},
	}
	return block
}

func (chain *Blockchain) GetBlockChain() (map[string]*pb.Block, map[int64]*pb.Block) {
	return chain.Blocks, chain.BlockAtHeight
}

func GetDebuginfo(block *pb.Block) struct {
	Hash       string
	ParentHash string
} {
	return struct {
		Hash       string
		ParentHash string
	}{
		Hash:       string(block.Hash),
		ParentHash: string(block.ParentHash),
	}
}
