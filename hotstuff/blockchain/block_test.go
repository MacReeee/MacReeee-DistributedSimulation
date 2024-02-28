package blockchain

import (
	"bytes"
	"crypto"
	"distributed/hotstuff/pb"
	"encoding/json"
	"fmt"
	"testing"
)

func Test_New(t *testing.T) {
	name := "hasher对象的使用"
	if name == "测试创建区块链实例" {
		chain := NewBlockChain()
		fmt.Println(chain)
	} else if name == "hasher对象的使用" {
		hasher := crypto.SHA256.New()
		hasher.Write([]byte("0000000000"))
		hashByte := hasher.Sum(nil)
		hash := fmt.Sprintf("%x", hashByte)
		fmt.Println(hash)
	}
}

type BLOCK struct {
	ParentHash string
	Hash       string
	Height     int64
	Cmd        string
	ViewNumber int64
	Proposer   int32
	Children   []string
}

func PrintChain(chain map[string]*pb.Block) {
	var block = BLOCK{}
	for _, b := range chain {
		block.ParentHash = string(b.ParentHash)
		block.Hash = string(b.Hash)
		block.Height = b.Height
		block.Cmd = string(b.Cmd)
		block.ViewNumber = b.ViewNumber
		block.Proposer = b.Proposer
		block.Children = b.Children

		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")

		err := encoder.Encode(block)
		if err != nil {
			fmt.Println("Error marshalling block:", err)
			return
		}
		fmt.Println(buffer.String())
	}

}

func Test_CreateBlock(t *testing.T) {
	chain := NewBlockChain()
	block := chain.CreateBlock([]byte("FFFFFFFFFFFF"), 0, nil, []byte("CMD of View: 1"), 1)
	chain.Store(block)
	block = chain.CreateBlock([]byte("3135bf433bdabb3f54568efa9e7c4b40113a7e73cdab92eaee47f649466c8d85"), 2, nil, []byte("CMD of View: 2"), 1)
	chain.Store(block)
	block = chain.CreateBlock([]byte("3135bf433bdabb3f54568efa9e7c4b40113a7e73cdab92eaee47f649466c8d85"), 2, nil, []byte("CMD of View: 3"), 1)
	chain.Store(block)
	PrintChain(chain.Blocks)
	fmt.Println("剪枝后：")
	chain.PruneBlock(chain.Blocks["3135bf433bdabb3f54568efa9e7c4b40113a7e73cdab92eaee47f649466c8d85"], chain.Blocks["17138ccb8156c5008835da8cddeaad0fefc9f168e9353d7dd55831f7e4fd172b"])
	PrintChain(chain.Blocks)
}
