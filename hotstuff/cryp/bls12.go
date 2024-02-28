package cryp

import (
	"crypto/rand"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
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

var (
	suite = pairing.NewSuiteBn256()
	sks   = []string{
		"436babacceaa0e985ce0b06321e50db789a5f9230b5ef3ffddc8ba64d2c896b8",
		"07ba027c422e35ac0b14794f858e924481e873fcb1771f9357e84055d37f580c",
		"80895cd18284111cb38a813c48b8533168b9adefe3ca5f8348261143391a5440",
		"7e079446464688b9f429ee9e6552e127ec413e3edffab7253b1475f7d79a7dc3",
	}
	pks = []string{
		"4e5bfe782aa1ddcdf6480170aa710f4ea366ed63f8b711ebf32305fb254470a7633934ce2a3647b58850823cf32da712643afde19b79d24659b5d7d3148a837a4e0fc74c919d39dfad40e41db5df66dcefcf171c26dfb73c181f176d767fa0686702e9b146bcaf2b6905bff1e25dce2078052474f742bf8c42eaf707afdcb3b7",
		"235e29bd16ad97a7ad0ebfd603fd24646720c794ac3acfb5b44fda366a1cb5b878515255f5e6c4550c8652138ab4feedbfa75c6fdd317b9b1bbafd1eae6ddac21422825a80240d787101b93f14fd37bdb63fa8cd0488032f5c87e69f115fda2e5b66422dc4d164a6ff9de252d40c7cca04f51228abce7ff5df40300936452dd8",
		"390d71fa7a3882048041b3d4d96cfb570c0f844b1df3bd3e73115d969e09a42f6d09ffafe980b37b1b1cc976793813f9ea23e82f8000e5badd05d6e66f65631e30793cd8930af7c20909a94e82b190dff7bd5898ee6b864383991e888db9c08b5e8b6f517f5be6e313c1a74b51fabfb1e77cf65ea9e3ed9a834cdeaa101479f6",
		"1118bd064e4f8c40669afa951d8293e2f4a14e12f02a229b6b40f21bb82853d4087dbc712c1a2e22a4047e0f163aa6fb5815a2e0c28ef19755832ac5bc008d4d55959029010376624a00c3ec38fc60501ee16c7bec0459b7b74bb43c3271422a33418b1bc2c7d1770215c2cf65267b7a06947a7276df773cbc81dbc2c4d80938",
	}
)

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

// 对 "${消息类型},${视图号},${区块hash}" 进行签名
func (s *Signer) Sign(msgType pb.MsgType, viewnumber int64, BlockHash []byte) ([]byte, error) {
	msg := []byte(fmt.Sprintf("%d,%d,%x", msgType, viewnumber, BlockHash))
	return bdn.Sign(suite, s.PrivateKey, msg)
}

func (s *Signer) NormSign(msg []byte) ([]byte, error) {
	return bdn.Sign(suite, s.PrivateKey, msg)
}

func (s *Signer) Verify(voter int32, msg []byte, sig []byte) bool {
	publicKey := suite.G2().Point()
	pk_bin, err := hex.DecodeString(pks[voter-1])
	if err != nil {
		log.Println("解码公钥失败:", err)
		return false
	}
	publicKey.UnmarshalBinary(pk_bin)
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
func (s *Signer) ThreshVerifyMock(QC *pb.QC) bool {
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
