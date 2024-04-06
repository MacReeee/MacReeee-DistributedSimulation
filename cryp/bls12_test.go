package cryp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"distributed/cryp/bls"
	pb2 "distributed/pb"
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
	fmt.Println("privKey1: ", privKey1)
	fmt.Println("pubKey1: ", pubKey1)

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

	var TestVoteMsg = pb2.VoteRequest{
		MsgType:    pb2.MsgType_PREPARE_VOTE,
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

	var VoteMsg1 = pb2.VoteRequest{
		Voter:      1,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}
	var VoteMsg2 = pb2.VoteRequest{
		Voter:      2,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}
	var VoteMsg3 = pb2.VoteRequest{
		Voter:      3,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}
	var VoteMsg4 = pb2.VoteRequest{
		Voter:      4,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 123,
		Hash:       []byte("hashhash"),
	}

	PartSign1, _ := signer1.Sign(VoteMsg1.MsgType, VoteMsg1.ViewNumber, VoteMsg1.Hash)
	PartSign2, _ := signer2.Sign(VoteMsg2.MsgType, VoteMsg2.ViewNumber, VoteMsg2.Hash)
	PartSign3, _ := signer3.Sign(VoteMsg3.MsgType, VoteMsg3.ViewNumber, VoteMsg3.Hash)
	PartSign4, _ := signer4.Sign(VoteMsg4.MsgType, VoteMsg4.ViewNumber, VoteMsg4.Hash)

	ThreshSign, aggpub, _ := signer2.ThreshMock([]int32{1, 2, 3, 4}, [][]byte{PartSign1, PartSign2, PartSign3, PartSign4})
	var TestQC = pb2.QC{
		MsgType:      pb2.MsgType_PREPARE_VOTE,
		BlsSignature: ThreshSign,
		AggPubKey:    aggpub,
		BlockHash:    []byte("hashhash"),
		ViewNumber:   123,
	}

	fmt.Println(signer3.ThreshVerifyMock(&TestQC))
}

func Test_GetPrepareQCSigs(t *testing.T) {
	signer1 := NewSignerByID(1)
	signer2 := NewSignerByID(2)
	signer3 := NewSignerByID(3)
	signer4 := NewSignerByID(4)

	var VoteMsg1 = pb2.VoteRequest{
		Voter:      1,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	var VoteMsg2 = pb2.VoteRequest{
		Voter:      2,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	var VoteMsg3 = pb2.VoteRequest{
		Voter:      3,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	var VoteMsg4 = pb2.VoteRequest{
		Voter:      4,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}

	PartSign1, _ := signer1.Sign(VoteMsg1.MsgType, VoteMsg1.ViewNumber, VoteMsg1.Hash)
	PartSign2, _ := signer2.Sign(VoteMsg2.MsgType, VoteMsg2.ViewNumber, VoteMsg2.Hash)
	PartSign3, _ := signer3.Sign(VoteMsg3.MsgType, VoteMsg3.ViewNumber, VoteMsg3.Hash)
	PartSign4, _ := signer4.Sign(VoteMsg4.MsgType, VoteMsg4.ViewNumber, VoteMsg4.Hash)

	signer2.ThreshMock([]int32{1, 2, 3, 4}, [][]byte{PartSign1, PartSign2, PartSign3, PartSign4})
	var TestQC = pb2.QC{
		MsgType:      pb2.MsgType_PREPARE_VOTE,
		BlsSignature: []byte{28, 201, 16, 247, 213, 76, 151, 58, 250, 236, 79, 128, 122, 208, 217, 160, 143, 88, 54, 100, 139, 163, 76, 61, 181, 63, 167, 129, 66, 245, 88, 25, 21, 227, 11, 228, 66, 141, 175, 202, 151, 51, 11, 128, 65, 198, 218, 133, 123, 164, 170, 45, 207, 25, 255, 78, 238, 39, 217, 167, 127, 128, 89, 139},
		AggPubKey:    []byte{117, 17, 230, 255, 170, 193, 55, 5, 104, 254, 206, 140, 207, 13, 157, 251, 133, 127, 45, 101, 201, 13, 104, 232, 86, 99, 251, 120, 113, 181, 236, 203, 70, 82, 146, 114, 242, 245, 4, 11, 211, 137, 204, 26, 203, 162, 239, 53, 243, 152, 103, 109, 92, 66, 136, 231, 15, 124, 233, 177, 118, 254, 203, 130, 92, 210, 35, 180, 213, 26, 215, 163, 131, 111, 55, 119, 137, 211, 176, 127, 113, 180, 169, 35, 14, 211, 188, 213, 131, 150, 197, 222, 81, 79, 34, 226, 86, 112, 162, 123, 244, 203, 105, 228, 102, 225, 87, 44, 143, 133, 131, 146, 201, 123, 173, 133, 70, 157, 160, 103, 241, 161, 127, 114, 201, 70, 247, 75},
		BlockHash:    []byte("FFFFFFFFFFFF"),
		ViewNumber:   0,
	}

	fmt.Println(signer3.ThreshVerifyMock(&TestQC))
	// fmt.Println("sig: ")
	// for _, v := range TestQC.BlsSignature {
	// 	fmt.Print(v, ",")
	// }
	// fmt.Println("pub: ")
	// for _, v := range TestQC.AggPubKey {
	// 	fmt.Print(v, ",")
	// }
	// 将 ThreshSign, aggpub 转换成16进制字符串
	// ThreshSignHex := hex.EncodeToString(ThreshSign)
	// AggPubHex := hex.EncodeToString(aggpub)
	// fmt.Println("ThreshSignHex: ", ThreshSignHex)
	// fmt.Println("AggPubHex: ", AggPubHex)

	// sig := []byte("1cc910f7d54c973afaec4f807ad0d9a08f5836648ba34c3db53fa78142f5581915e30be4428dafca97330b8041c6da857ba4aa2dcf19ff4eee27d9a77f80598b")
	// msg := []byte(fmt.Sprintf("%d,%d,%x", pb.MsgType_PREPARE_VOTE, 0, []byte("FFFFFFFFFFFF")))
	// var pk = suite.G2().Point()
	// pk.UnmarshalBinary([]byte("7511e6ffaac1370568fece8ccf0d9dfb857f2d65c90d68e85663fb7871b5eccb46529272f2f5040bd389cc1acba2ef35f398676d5c4288e70f7ce9b176fecb825cd223b4d51ad7a3836f377789d3b07f71b4a9230ed3bcd58396c5de514f22e25670a27bf4cb69e466e1572c8f858392c97bad85469da067f1a17f72c946f74b"))

	// fmt.Println(bdn.Verify(suite, pk, msg, sig) == nil)
}

func Test_GetLockedQCSigs(t *testing.T) {
	signer1 := NewSignerByID(1)
	signer2 := NewSignerByID(2)
	signer3 := NewSignerByID(3)
	signer4 := NewSignerByID(4)

	var VoteMsg1 = pb2.VoteRequest{
		Voter:      1,
		MsgType:    pb2.MsgType_PRE_COMMIT_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	var VoteMsg2 = pb2.VoteRequest{
		Voter:      2,
		MsgType:    pb2.MsgType_PRE_COMMIT_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	var VoteMsg3 = pb2.VoteRequest{
		Voter:      3,
		MsgType:    pb2.MsgType_PRE_COMMIT_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	var VoteMsg4 = pb2.VoteRequest{
		Voter:      4,
		MsgType:    pb2.MsgType_PRE_COMMIT_VOTE,
		ViewNumber: 0,
		Hash:       []byte("FFFFFFFFFFFF"),
	}

	PartSign1, _ := signer1.Sign(VoteMsg1.MsgType, VoteMsg1.ViewNumber, VoteMsg1.Hash)
	PartSign2, _ := signer2.Sign(VoteMsg2.MsgType, VoteMsg2.ViewNumber, VoteMsg2.Hash)
	PartSign3, _ := signer3.Sign(VoteMsg3.MsgType, VoteMsg3.ViewNumber, VoteMsg3.Hash)
	PartSign4, _ := signer4.Sign(VoteMsg4.MsgType, VoteMsg4.ViewNumber, VoteMsg4.Hash)

	signer2.ThreshMock([]int32{1, 2, 3, 4}, [][]byte{PartSign1, PartSign2, PartSign3, PartSign4})
	var TestQC = pb2.QC{
		MsgType:      pb2.MsgType_PRE_COMMIT_VOTE,
		BlsSignature: []byte{56, 102, 243, 138, 71, 19, 119, 120, 28, 242, 150, 51, 203, 31, 93, 78, 112, 144, 141, 87, 48, 195, 12, 60, 15, 160, 155, 17, 48, 87, 131, 51, 39, 119, 159, 49, 183, 198, 110, 188, 38, 200, 189, 59, 237, 239, 28, 28, 91, 84, 231, 78, 5, 75, 141, 214, 29, 174, 5, 46, 32, 26, 68, 5},
		AggPubKey:    []byte{117, 17, 230, 255, 170, 193, 55, 5, 104, 254, 206, 140, 207, 13, 157, 251, 133, 127, 45, 101, 201, 13, 104, 232, 86, 99, 251, 120, 113, 181, 236, 203, 70, 82, 146, 114, 242, 245, 4, 11, 211, 137, 204, 26, 203, 162, 239, 53, 243, 152, 103, 109, 92, 66, 136, 231, 15, 124, 233, 177, 118, 254, 203, 130, 92, 210, 35, 180, 213, 26, 215, 163, 131, 111, 55, 119, 137, 211, 176, 127, 113, 180, 169, 35, 14, 211, 188, 213, 131, 150, 197, 222, 81, 79, 34, 226, 86, 112, 162, 123, 244, 203, 105, 228, 102, 225, 87, 44, 143, 133, 131, 146, 201, 123, 173, 133, 70, 157, 160, 103, 241, 161, 127, 114, 201, 70, 247, 75},
		BlockHash:    []byte("FFFFFFFFFFFF"),
		ViewNumber:   0,
	}

	fmt.Println(signer3.ThreshVerifyMock(&TestQC))
	// fmt.Println("sig: ")
	// for _, v := range TestQC.BlsSignature {
	// 	fmt.Print(v, ",")
	// }
	// fmt.Println("pub: ")
	// for _, v := range TestQC.AggPubKey {
	// 	fmt.Print(v, ",")
	// }
	// 将 ThreshSign, aggpub 转换成16进制字符串
	// ThreshSignHex := hex.EncodeToString(ThreshSign)
	// AggPubHex := hex.EncodeToString(aggpub)
	// fmt.Println("ThreshSignHex: ", ThreshSignHex)
	// fmt.Println("AggPubHex: ", AggPubHex)

	// sig := []byte("1cc910f7d54c973afaec4f807ad0d9a08f5836648ba34c3db53fa78142f5581915e30be4428dafca97330b8041c6da857ba4aa2dcf19ff4eee27d9a77f80598b")
	// msg := []byte(fmt.Sprintf("%d,%d,%x", pb.MsgType_PREPARE_VOTE, 0, []byte("FFFFFFFFFFFF")))
	// var pk = suite.G2().Point()
	// pk.UnmarshalBinary([]byte("7511e6ffaac1370568fece8ccf0d9dfb857f2d65c90d68e85663fb7871b5eccb46529272f2f5040bd389cc1acba2ef35f398676d5c4288e70f7ce9b176fecb825cd223b4d51ad7a3836f377789d3b07f71b4a9230ed3bcd58396c5de514f22e25670a27bf4cb69e466e1572c8f858392c97bad85469da067f1a17f72c946f74b"))

	// fmt.Println(bdn.Verify(suite, pk, msg, sig) == nil)
}

func Test_TestVerify(t *testing.T) {
	signer := NewSignerByID(1)
	msg := []byte("helloworld")
	msg1 := []byte("hellowo2rld")
	sig, _ := signer.NormSign(msg)

	res := signer.Verify(1, msg1, sig)
	fmt.Println(res)
}

// 公私钥对生成器
func Test_GenerateKeyPair2(t *testing.T) {
	privKey, pubKey := bls.NewKeyPair(suite, random.New())
	s, _ := privKey.MarshalBinary()
	p, _ := pubKey.MarshalBinary()
	//将公私钥转换成16进制字符串
	privHex := hex.EncodeToString(s)
	pubHex := hex.EncodeToString(p)
	fmt.Println("priv (hex): ", privHex)
	fmt.Println("pub (hex): ", pubHex)
}

func Test_NewKeyPair(t *testing.T) {
	var i int32 = 10
	signer5 := NewSignerByID(i)
	msg := []byte("hello world")
	sig5, _ := signer5.NormSign(msg)

	signer1 := NewSignerByID(1)
	fmt.Println(signer1.Verify(i, []byte("msg"), sig5))
	fmt.Println(signer1.Verify(i, msg, sig5))

}
