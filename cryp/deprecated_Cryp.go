package cryp

import (
	"distributed/modules"
	"encoding/base64"
	"fmt"
	"log"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

// 初始化时存储所有的公钥
var ALL_PUBLIC_KEYS = []kyber.Point{
	GetPubKey([]byte("T4EXVUEVDV+FdrC7EKqIp932R47XWkf6AKKEbhV6fqWFusIVi4gsUxnNNDBvSoO/+ckKXWQJ2ewklcTLJvWdQFl9hfkqQ/L/tezCpDZTxN8Hj4aRICXTw9rDHCDGjK/MRpDr6V7Xfz9K/LnYKZV0hhg3/GibzbGyunfImMC4n0s=")),
	GetPubKey([]byte("gpS8ARul05I5VV+qQR9s+ls4VwiWs2yCYJ2zB3AwHqFxuPttDqE6ZOpvpIru27P7zoXpPu4xyS3s/V5/Wi4Vam4ohGfud0ThKyyXObS53ATBGx1LgRU64dbx6FyrtO5MAkvzBsXR2Yl/VQ88IkGdumNkuZwpWRLEoeRVjeyinng=")),
	GetPubKey([]byte("js6Y32MFO384n5kVdkHaW7VIjZDTL3wnbfo7n/LscgQVSuqiuK24G8XyV/9XagHcdGXR200BPJATbbD3jtvs1ECAxCSnxY+fMDkrTYzTdBdTcvNq256Bmy+l9AaLstc2CXQsMoCJfvf1nCjfVrImrTkW5j/eCuhzXWO3PpdEwoo=")),
	GetPubKey([]byte("PHUkbumWlmbLVAiMiDH3R9TLVRrFlxektKLuJCzYzyU/RwMigm7T8lutcGwoWCrZG0YaWlyP6cJQGW5yYuB7Qo40AMLTLgExRNxFD057kXNby0XbaQERAOfngRfzqBkqCzPE8YgQh8XS3MunxG/2NBkM644r+h6nopkWqBWgqR0=")),
}

func PartSign(msg []byte, privKey kyber.Scalar) ([]byte, error) {
	sig, err := bdn.Sign(suite, privKey, msg)
	return sig, err
}

func Verify(msg []byte, sig []byte, pubKey kyber.Point) bool {
	return bdn.Verify(suite, pubKey, msg, sig) == nil
}

func GetPrivKey(privateKey []byte) kyber.Scalar { //将私钥转换为kyber.Scalar类型
	decoded, err := base64.StdEncoding.DecodeString(string(privateKey))
	if err != nil {
		log.Println("base64解码私钥失败:", err)
		return nil
	}

	privKey := suite.G2().Scalar()
	err = privKey.UnmarshalBinary(decoded)
	if err != nil {
		log.Println("解码私钥失败:", err)
		return nil
	}

	return privKey
}

func GetPubKey(publicKey []byte) kyber.Point { //将公钥转换为kyber.Point类型
	decoded, err := base64.StdEncoding.DecodeString(string(publicKey))
	if err != nil {
		log.Println("base64解码公钥失败:", err)
		return nil
	}

	pubKey := suite.G2().Point()
	err = pubKey.UnmarshalBinary(decoded)
	if err != nil {
		log.Println("解码公钥失败:", err)
		return nil
	}

	return pubKey
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

	fmt.Println("mask:", mask)
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
	privKey := GetPrivKey(priv)

	pubKey := GetPubKey(pub)

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

func (s *SignerAndVerifier) ThresholdSign(msg []byte, SigMap map[kyber.Point][]byte) ([]byte, kyber.Point, error) {
	return ThresholdSign(msg, SigMap)
}
