package cryp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"distributed/hotstuff/cryp/bls"
	"distributed/hotstuff/pb"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/util/random"
)

func TestThresholdSign(t *testing.T) {
	msg := []byte("Hello World")
	priv1 := GetPrivKey([]byte("iQ1S0Jg0dncRdKVdU06nYNWYCESf/64a7S3/qtSXvXs="))
	pub1 := GetPubKey([]byte("T4EXVUEVDV+FdrC7EKqIp932R47XWkf6AKKEbhV6fqWFusIVi4gsUxnNNDBvSoO/+ckKXWQJ2ewklcTLJvWdQFl9hfkqQ/L/tezCpDZTxN8Hj4aRICXTw9rDHCDGjK/MRpDr6V7Xfz9K/LnYKZV0hhg3/GibzbGyunfImMC4n0s="))
	priv2 := GetPrivKey([]byte("dtav8hvEgXiUIgX708EDOMdnHkmd4kN09dnUw3ecmE0="))
	pub2 := GetPubKey([]byte("gpS8ARul05I5VV+qQR9s+ls4VwiWs2yCYJ2zB3AwHqFxuPttDqE6ZOpvpIru27P7zoXpPu4xyS3s/V5/Wi4Vam4ohGfud0ThKyyXObS53ATBGx1LgRU64dbx6FyrtO5MAkvzBsXR2Yl/VQ88IkGdumNkuZwpWRLEoeRVjeyinng="))
	priv3 := GetPrivKey([]byte("hU0WsKy2GFaR53Jf3vmq9dUPSOFrVi5JK4wZjc4Zv/M="))
	pub3 := GetPubKey([]byte("js6Y32MFO384n5kVdkHaW7VIjZDTL3wnbfo7n/LscgQVSuqiuK24G8XyV/9XagHcdGXR200BPJATbbD3jtvs1ECAxCSnxY+fMDkrTYzTdBdTcvNq256Bmy+l9AaLstc2CXQsMoCJfvf1nCjfVrImrTkW5j/eCuhzXWO3PpdEwoo="))
	sig1, _ := PartSign(msg, priv1)
	sig2, _ := PartSign(msg, priv2)
	sig3, _ := PartSign(msg, priv3)

	// 创建映射保证公钥和签名一一对应，map[公钥]签名
	var sigMap = make(map[kyber.Point][]byte)
	sigMap[pub1] = sig1
	sigMap[pub2] = sig2
	sigMap[pub3] = sig3

	thSig, aggPub, _ := ThresholdSign(msg, sigMap)
	fmt.Println(Verify(msg, thSig, aggPub))
	fmt.Println(Verify([]byte("abcabc"), thSig, aggPub))
}

func TestNilVerify(t *testing.T) {
	msg := []byte("Hello World")
	priv := GetPrivKey([]byte("iQ1S0Jg0dncRdKVdU06nYNWYCESf/64a7S3/qtSXvXs="))
	pub := GetPubKey([]byte("T4EXVUEVDV+FdrC7EKqIp932R47XWkf6AKKEbhV6fqWFusIVi4gsUxnNNDBvSoO/+ckKXWQJ2ewklcTLJvWdQFl9hfkqQ/L/tezCpDZTxN8Hj4aRICXTw9rDHCDGjK/MRpDr6V7Xfz9K/LnYKZV0hhg3/GibzbGyunfImMC4n0s="))

	sig, _ := PartSign(msg, priv)
	fmt.Println(Verify(msg, sig, pub))
	fmt.Println(Verify([]byte("Hello Worlds"), sig, pub))
}

func TestAgg(t *testing.T) {

	msg := []byte("Hello World")

	priv1 := GetPrivKey([]byte("iQ1S0Jg0dncRdKVdU06nYNWYCESf/64a7S3/qtSXvXs="))
	pub1 := GetPubKey([]byte("T4EXVUEVDV+FdrC7EKqIp932R47XWkf6AKKEbhV6fqWFusIVi4gsUxnNNDBvSoO/+ckKXWQJ2ewklcTLJvWdQFl9hfkqQ/L/tezCpDZTxN8Hj4aRICXTw9rDHCDGjK/MRpDr6V7Xfz9K/LnYKZV0hhg3/GibzbGyunfImMC4n0s="))

	priv2 := GetPrivKey([]byte("dtav8hvEgXiUIgX708EDOMdnHkmd4kN09dnUw3ecmE0="))
	pub2 := GetPubKey([]byte("gpS8ARul05I5VV+qQR9s+ls4VwiWs2yCYJ2zB3AwHqFxuPttDqE6ZOpvpIru27P7zoXpPu4xyS3s/V5/Wi4Vam4ohGfud0ThKyyXObS53ATBGx1LgRU64dbx6FyrtO5MAkvzBsXR2Yl/VQ88IkGdumNkuZwpWRLEoeRVjeyinng="))

	priv3 := GetPrivKey([]byte("hU0WsKy2GFaR53Jf3vmq9dUPSOFrVi5JK4wZjc4Zv/M="))
	pub3 := GetPubKey([]byte("js6Y32MFO384n5kVdkHaW7VIjZDTL3wnbfo7n/LscgQVSuqiuK24G8XyV/9XagHcdGXR200BPJATbbD3jtvs1ECAxCSnxY+fMDkrTYzTdBdTcvNq256Bmy+l9AaLstc2CXQsMoCJfvf1nCjfVrImrTkW5j/eCuhzXWO3PpdEwoo="))

	sig1, _ := PartSign(msg, priv1)
	sig2, _ := PartSign(msg, priv2)
	sig3, _ := PartSign(msg, priv3)

	// 创建映射保证公钥和签名一一对应，map[公钥]签名
	var sigMap = make(map[kyber.Point][]byte)
	sigMap[pub1] = sig1
	sigMap[pub2] = sig2
	sigMap[pub3] = sig3

	var signatures [][]byte
	var pubKeys []kyber.Point
	mask, _ := sign.NewMask(suite, ALL_PUBLIC_KEYS, nil)
	for i, v := range ALL_PUBLIC_KEYS {
		if sig, ok := sigMap[v]; ok {
			mask.SetBit(i, true)
			signatures = append(signatures, sig)
			pubKeys = append(pubKeys, v)
		}
	}

	aggSig, _ := bdn.AggregateSignatures(suite, signatures, mask)
	aggSigBytes, err := aggSig.MarshalBinary()
	if err != nil {
		t.Errorf("Error marshalling aggregated signature: %v", err)
		return
	}
	aggPub, _ := bdn.AggregatePublicKeys(suite, mask)
	//以上是签名部分，以下是验证部分
	fmt.Println(bdn.Verify(suite, aggPub, msg, aggSigBytes) == nil)
	fmt.Println(bdn.Verify(suite, aggPub, nil, aggSigBytes) == nil)
	fmt.Println(bdn.Verify(suite, aggPub, []byte("jhgjhgjh"), aggSigBytes) == nil)
}

func TestOnce(t *testing.T) {
	var once = sync.Once{}
	func1 := func() {
		fmt.Println("func1")
	}
	func2 := func() {
		fmt.Println("func2")
	}
	once.Do(func1)
	once.Do(func2)
	fmt.Println("end")
}

func TestMap(t *testing.T) {
	var m1 = make(map[int]*string)
	var m2 = make(map[string]*string)
	s1 := "hello"
	// s2 := "world"
	m1[1] = &s1
	m2["1"] = &s1
	fmt.Println("m1: ", m1)
	fmt.Println("m2: ", m2)
	delete(m1, 1)
	fmt.Println("m1: ", m1)
	fmt.Println("m2: ", m2)
}

func Test_Ecdsa(t *testing.T) {
	// 生成密钥对
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "密钥生成失败: %s", err)
		return
	}

	// 模拟一个需要签名的消息
	message := "需要签名的消息"
	hash := sha256.Sum256([]byte(message))

	// 签名
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "签名失败: %s", err)
		return
	}

	fmt.Printf("签名结果: (r=%s, s=%s)\n", r, s)

	// 验证签名
	valid := ecdsa.Verify(&privKey.PublicKey, hash[:], r, s)
	if valid {
		fmt.Println("签名验证成功！")
	} else {
		fmt.Println("签名验证失败。")
	}
}

func Test_bls(t *testing.T) {
	suite := bn256.NewSuite()
	privKey1, pubKey1 := bls.NewKeyPair(suite, random.New())
	// fmt.Println("privKey1: ", privKey1)
	// fmt.Println("pubKey1: ", pubKey1)

	privKey2, pubKey2 := bls.NewKeyPair(suite, random.New())
	// fmt.Println("privKey2: ", privKey2)
	// fmt.Println("pubKey2: ", pubKey2)

	privKey3, pubKey3 := bls.NewKeyPair(suite, random.New())
	// fmt.Println("privKey3: ", privKey3)
	// fmt.Println("pubKey3: ", pubKey3)

	privKey4, pubKey4 := bls.NewKeyPair(suite, random.New())
	// fmt.Println("privKey4: ", privKey4)
	// fmt.Println("pubKey4: ", pubKey4)

	// message1 := []byte("Hello, BLS!1 ")
	// message2 := []byte("Hello, BLS!2 ")
	// message3 := []byte("Hello, BLS!3 ")
	// message4 := []byte("Hello, BLS!4 ")

	message1 := []byte("Hello, BLS! ")
	message2 := []byte("Hello, BLS! ")
	message3 := []byte("Hello, BLS! ")
	message4 := []byte("Hello, BLS! ")

	signature, _ := bls.Sign(suite, privKey1, message1)
	signature2, _ := bls.Sign(suite, privKey2, message2)
	signature3, _ := bls.Sign(suite, privKey3, message3)
	signature4, _ := bls.Sign(suite, privKey4, message4)

	AggPub := bls.AggregatePublicKeys(suite, pubKey1, pubKey2, pubKey3, pubKey4)
	AggSig, _ := bls.AggregateSignatures(suite, signature, signature2, signature3, signature4)

	// fmt.Println("单独验证结果:", bls.Verify(suite, pubKey1, message1, signature) == nil)

	fmt.Println("聚合验证结果:", bls.Verify(suite, AggPub, message1, AggSig) == nil)
	// fmt.Println("用私钥验证结果:", bls.Verify(suite, AggPub, message1, signature) == nil)
	// fmt.Println("测试同一聚合签名，不同消息的聚合验证结果:", bls.Verify(suite, AggPub, []byte("message"), AggSig) == nil)
	// fmt.Println("测试群验证结果：",
	// 	bls.BatchVerify(suite,
	// 		[]kyber.Point{pubKey1,
	// 			pubKey2,
	// 			pubKey3,
	// 			pubKey4,
	// 		},
	// 		[][]byte{message1, message2, message3, message4},
	// 		AggSig,
	// 	) == nil)
}

func Test_Rand(t *testing.T) {
	priv, pub := bls.NewKeyPair(suite, &RandomStream{})
	fmt.Println("priv: ", priv)
	fmt.Println("pub: ", pub)
}

func Test_GenerateKeyPair(t *testing.T) {
	for i := 0; i < 4; i++ {
		signer := NewSigner()

		// 转换为16进制字符串
		privHex := hex.EncodeToString(signer.PrivKey())
		pubHex := hex.EncodeToString(signer.PubKey())

		fmt.Println("priv (hex): ", privHex)
		fmt.Println("pub (hex): ", pubHex)
		fmt.Printf("\n")
	}
}

func Test_NewFromBin(t *testing.T) {
	signer := NewSignerByID(1)
	message := "hello world"
	message2 := "hello world2"
	sig, _ := bdn.Sign(suite, signer.PrivateKey, []byte(message))
	fmt.Println(bdn.Verify(suite, signer.PublicKey, []byte(message), sig) == nil)
	fmt.Println(bdn.Verify(suite, signer.PublicKey, []byte(message2), sig) == nil)
}

func Test_SignAndVeri(t *testing.T) {
	signer1 := NewSignerByID(1)
	signer2 := NewSignerByID(2)
	// signer3 := NewSignerByID(3)
	// signer4 := NewSignerByID(4)

	var TestVoteMsg = pb.VoteRequest{
		MsgType:    pb.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}

	sig1, _ := signer1.Sign(TestVoteMsg.MsgType, TestVoteMsg.ViewNumber, TestVoteMsg.Hash)

	vfmsg := []byte(fmt.Sprintf("%d,%d,%x", TestVoteMsg.MsgType, TestVoteMsg.ViewNumber, TestVoteMsg.Hash))
	msg := []byte("hello world")
	fmt.Println(signer2.Verify(1, vfmsg, sig1))
	fmt.Println(signer2.Verify(1, msg, sig1))
}

func Test_ThreshAndVerifyMock(t *testing.T) {
	signer1 := NewSignerByID(1)
	signer2 := NewSignerByID(2)
	signer3 := NewSignerByID(3)
	signer4 := NewSignerByID(4)

	var VoteMsg1 = pb.VoteRequest{
		Voter:      1,
		MsgType:    pb.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}
	var VoteMsg2 = pb.VoteRequest{
		Voter:      2,
		MsgType:    pb.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}
	var VoteMsg3 = pb.VoteRequest{
		Voter:      3,
		MsgType:    pb.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}
	var VoteMsg4 = pb.VoteRequest{
		Voter:      4,
		MsgType:    pb.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}

	PartSign1, _ := signer1.Sign(VoteMsg1.MsgType, VoteMsg1.ViewNumber, VoteMsg1.Hash)
	PartSign2, _ := signer2.Sign(VoteMsg2.MsgType, VoteMsg2.ViewNumber, VoteMsg2.Hash)
	PartSign3, _ := signer3.Sign(VoteMsg3.MsgType, VoteMsg3.ViewNumber, VoteMsg3.Hash)
	PartSign4, _ := signer4.Sign(VoteMsg4.MsgType, VoteMsg4.ViewNumber, VoteMsg4.Hash)

	ThreshSign, aggpub, _ := signer2.ThreshMock([]int32{1, 2, 3, 4}, [][]byte{PartSign1, PartSign2, PartSign3, PartSign4})
	var TestQC = pb.QC{
		MsgType:      pb.MsgType_PREPARE_VOTE,
		BlsSignature: ThreshSign,
		AggPubKey:    aggpub,
		BlockHash:    []byte("hashhash"),
		ViewNumber:   123,
	}

	fmt.Println(signer3.ThreshVerifyMock(&TestQC))
}
