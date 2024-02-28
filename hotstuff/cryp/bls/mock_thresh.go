package bls

import (
	"crypto/rand"
	"distributed/hotstuff/pb"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
)

// RandomStream 结构体实现 cipher.Stream 接口
type RandomStream struct{}

// XORKeyStream 方法填充 dst 为随机数据
// 由于生成随机数据，src 参数在此实现中被忽略
func (s *RandomStream) XORKeyStream(dst []byte, src []byte) {
	if _, err := io.ReadFull(rand.Reader, dst); err != nil {
		panic("随机数生成失败: " + err.Error()) // 在实际应用中应更优雅地处理错误
	}
}

type Signer struct {
	PublicKey  kyber.Point
	PrivateKey kyber.Scalar
}

var suite = bn256.NewSuite()

func NewSigner() *Signer {
	priv, pub := NewKeyPair(suite, &RandomStream{})
	return &Signer{
		PublicKey:  pub,
		PrivateKey: priv,
	}
}

// 对 "${消息类型},${视图号},${区块hash}" 进行签名
func (s *Signer) Sign(msgType pb.MsgType, viewnumber int64, BlockHash []byte) ([]byte, error) {
	msg := []byte(fmt.Sprintf("%d,%d,%x", msgType, viewnumber, BlockHash))
	return Sign(suite, s.PrivateKey, msg)
}

func (s *Signer) Verify(publicKey kyber.Point, msg []byte, sig []byte) bool {
	return Verify(suite, publicKey, msg, sig) == nil
}

func (s *Signer) AggregateSignatures(voter []int32, sigs ...[]byte) ([]byte, error) {
	return AggregateSignatures(suite, sigs...)
}

// 公钥是聚合*以前*的公钥，签名是聚合*以后*的签名；
// 各个节点首先对不同的消息进行签名，注意是不同的消息，然后将签名聚合
// 这个函数的作用是用一个签名同时验证多个消息
func (s *Signer) BatchVerify(publicKeys []kyber.Point, msgs [][]byte, sig []byte) bool {
	return BatchVerify(suite, publicKeys, msgs, sig) == nil
}

// 每个节点签名以后投票
// 投票的签名应当是 "${消息类型},${视图号},${区块hash}"，各个节点签名内容相同
// 投票信息中记录Voter，Leader根据Voter来选择聚合哪些公钥
// leader聚合所选公钥，聚合签名，并将两者都放入QC中
// 节点可通过QC中携带的聚合公钥和聚合签名来验证
// 聚合的公钥和签名都仅当轮有效
// 由上分析可知，传入投票者和其对应签名，可不按顺序，返回聚合的公钥和签名
func (s *Signer) MockThresh(voter []int32, sigs [][]byte) ([]byte, kyber.Point, error) {
	return nil, nil, nil
}
