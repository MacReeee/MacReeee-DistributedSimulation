package crypto

import (
	"fmt"
	"sync"
	"testing"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
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
	fmt.Println(bdn.Verify(suite, pub1, nil, aggSigBytes))
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
