package cryp

import (
	"crypto/rand"
	"distributed/modules"
	pb2 "distributed/pb"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/bdn"
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

func NewSignerByID(ID int32) *Signer {
	sk := sks[ID-1]
	pk := pks[ID-1]

	var priv = suite.G2().Scalar()
	var pub = suite.G2().Point()

	sk_bin, _ := hex.DecodeString(sk)
	pk_bin, _ := hex.DecodeString(pk)

	priv.UnmarshalBinary(sk_bin)
	pub.UnmarshalBinary(pk_bin)

	var signer = &Signer{
		PublicKey:  pub,
		PrivateKey: priv,
	}
	modules.MODULES.Signer = signer
	return signer
}

// 对 "${消息类型},${视图号},${区块hash}" 进行签名，投票和发布QC都运用此签名规则
func (s *Signer) Sign(msgType pb2.MsgType, viewnumber int64, BlockHash []byte) ([]byte, error) {
	msg := []byte(fmt.Sprintf("%d,%d,%x", msgType, viewnumber, BlockHash))
	return bdn.Sign(suite, s.PrivateKey, msg)
}

func (s *Signer) NormSign(msg []byte) ([]byte, error) {
	return bdn.Sign(suite, s.PrivateKey, msg)
}

// 传入顺序：投票者，消息，签名
func (s *Signer) Verify(voter int32, msg []byte, sig []byte) bool {
	publicKey := suite.G2().Point()
	pkBin, err := hex.DecodeString(pks[voter-1])
	if err != nil {
		log.Println("解码公钥失败:", err)
		return false
	}
	publicKey.UnmarshalBinary(pkBin)
	return bdn.Verify(suite, publicKey, msg, sig) == nil
}

// 每个节点签名以后投票
// 投票的签名应当是 "${消息类型},${视图号},${区块hash}"，各个节点签名内容相同
// QC的内容应当是对收到的投票的聚合，也就是返回值的第一个
// 投票信息中记录Voter，Leader根据Voter来选择聚合哪些公钥
// leader聚合所选公钥，聚合签名，并将两者都放入QC中
// 节点可通过QC中携带的聚合公钥和聚合签名来验证
// 聚合的公钥和签名都仅当轮有效
// 由上分析可知，传入投票者和其对应签名，可不按顺序，返回聚合的签名和公钥
func (s *Signer) ThreshMock(voter []int32, sigs [][]byte) ([]byte, []byte, error) {
	// 获取voter对应的公钥
	var pubkeys []kyber.Point
	var pk kyber.Point
	for _, v := range voter {
		pk = suite.G2().Point()
		pk_bin, err := hex.DecodeString(pks[v-1])
		if err != nil {
			log.Println("解码公钥失败:", err)
			return nil, nil, err
		}
		pk.UnmarshalBinary(pk_bin)
		pubkeys = append(pubkeys, pk)
	}
	//聚合公钥
	aggregated := suite.G2().Point()
	for _, X := range pubkeys {
		aggregated.Add(aggregated, X)
	}
	//聚合签名
	var aggregatedSig = suite.G1().Point()
	for _, sig := range sigs {
		sigPoint := suite.G1().Point()
		if err := sigPoint.UnmarshalBinary(sig); err != nil {
			log.Println("解码签名失败:", err)
			return nil, nil, err
		}
		aggregatedSig.Add(aggregatedSig, sigPoint)
	}
	// 序列化签名和公钥
	aggSigBytes, err := aggregatedSig.MarshalBinary()
	if err != nil {
		log.Println("序列化聚合签名失败:", err)
		return nil, nil, err
	}
	aggPubKeyBytes, err := aggregated.MarshalBinary()
	if err != nil {
		log.Println("序列化聚合公钥失败:", err)
		return nil, nil, err
	}
	return aggSigBytes, aggPubKeyBytes, nil
}

// 由上所述，验证QC只需要传入收到的QC
func (s *Signer) ThreshVerifyMock(QC *pb2.QC) bool {
	msg := []byte(fmt.Sprintf("%d,%d,%x", QC.MsgType, QC.ViewNumber, QC.BlockHash))
	AggPubKey := suite.G2().Point()
	AggPubKey.UnmarshalBinary(QC.AggPubKey)
	return bdn.Verify(suite, AggPubKey, msg, QC.BlsSignature) == nil
}

func NewSigner() *Signer {
	priv, pub := bdn.NewKeyPair(suite, &RandomStream{})
	return &Signer{
		PublicKey:  pub,
		PrivateKey: priv,
	}
}

func (s *Signer) PubKey() []byte {
	pubKey, err := s.PublicKey.MarshalBinary()
	if err != nil {
		log.Println("序列化公钥失败:", err)
		return nil
	}
	return pubKey
}

func (s *Signer) PrivKey() []byte {
	privKey, err := s.PrivateKey.MarshalBinary()
	if err != nil {
		log.Println("序列化私钥失败:", err)
		return nil
	}
	return privKey
}
