package crypto

import (
	"distributed/hotstuff/modules"
	"encoding/base64"
	"fmt"
	"log"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

var suite = pairing.NewSuiteBn256()

// 初始化时存储所有的公钥
var ALL_PUBLIC_KEYS []kyber.Point

func PartSign(msg []byte, privKey kyber.Scalar) ([]byte, error) {
	sig, err := bdn.Sign(suite, privKey, msg)
	return sig, err
}

func Verify(msg []byte, sig []byte, pubKey kyber.Point) bool {
	return bdn.Verify(suite, pubKey, msg, sig) == nil
}

func GetPrivKey(privateKey []byte) (kyber.Scalar, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(privateKey))
	if err != nil {
		return nil, err
	}

	privKey := suite.G2().Scalar()
	err = privKey.UnmarshalBinary(decoded)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func GetPubKey(publicKey []byte) (kyber.Point, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(publicKey))
	if err != nil {
		return nil, err
	}

	pubKey := suite.G2().Point()
	err = pubKey.UnmarshalBinary(decoded)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// pubKeys是参与签名的节点的公钥
func ThresholdSign(msg []byte, SigMap map[kyber.Point][]byte) ([]byte, kyber.Point, error) {
	var signatures [][]byte
	var pubKeys []kyber.Point
	// 创建一个 mask，标记所有参与的公钥
	mask, err := sign.NewMask(suite, ALL_PUBLIC_KEYS, nil)
	if err != nil {
		return nil, nil, err
	}
	for i := range pubKeys {
		mask.SetBit(i, false)
	}
	for i, v := range ALL_PUBLIC_KEYS {
		if sig, ok := SigMap[v]; ok {
			mask.SetBit(i, true)
			signatures = append(signatures, sig)
			pubKeys = append(pubKeys, v)
		}
	}

	// 聚合签名
	aggSig, err := bdn.AggregateSignatures(suite, signatures, mask)
	if err != nil {
		return nil, nil, err
	}
	//聚合公钥
	aggPub, _ := bdn.AggregatePublicKeys(suite, mask)
	if err != nil {
		fmt.Println("Error aggregating public keys:", err)
		return nil, nil, err
	}

	// 序列化聚合签名
	aggSigBytes, err := aggSig.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}

	return aggSigBytes, aggPub, nil
}

type SignerAndVerifier struct {
	privKey kyber.Scalar
	pubKey  kyber.Point
}

func NewSignerAndVerifier(priv []byte, pub []byte) *SignerAndVerifier {
	privKey, err := GetPrivKey(priv)
	if err != nil {
		log.Println("转换私钥失败:", err)
	}
	pubKey, err := GetPubKey(pub)
	if err != nil {
		log.Println("转换公钥失败:", err)
	}
	s := &SignerAndVerifier{
		privKey: privKey,
		pubKey:  pubKey,
	}
	modules.MODULES.SignerAndVerifier = s
	return s
}

func (s *SignerAndVerifier) Sign(msg []byte) ([]byte, error) {
	return bdn.Sign(suite, s.privKey, msg)
}

func (s *SignerAndVerifier) PartSign(msg []byte) ([]byte, error) {
	return PartSign(msg, s.privKey)
}

func (s *SignerAndVerifier) Verify(msg []byte, sig []byte) bool {
	return Verify(msg, sig, s.pubKey)
}

func (s *SignerAndVerifier) ThreshVerify(msg []byte, sig []byte, pubKey kyber.Point) bool {
	return Verify(msg, sig, pubKey)
}
