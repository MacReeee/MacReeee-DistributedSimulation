package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"distributed/hotstuff/blockchain"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/cryp"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/view"
	"fmt"
	"testing"

	stcryp "crypto"

	"github.com/niclabs/tcrsa"
	blscryp "go.dedis.ch/dela/crypto"

	"go.dedis.ch/dela/crypto/bls"
)

func Test_main(t *testing.T) {
	modules := modules.MODULES
	blockchain.NewBlockChain()
	view.New()
	cryp.NewSignerByID(1)
	hotstuff.NewReplicaServer(1)
	fmt.Println(modules)
}

func Test_AggregateSign(t *testing.T) {
	signerA := bls.NewSigner()
	signerB := bls.NewSigner()
	signerC := bls.NewSigner()

	publicKeys := []blscryp.PublicKey{ //有三个人参与但是却只有两个人签名，因此无法验证
		signerA.GetPublicKey(),
		signerB.GetPublicKey(),
		signerC.GetPublicKey(),
	}

	message := []byte("Hello, World!")
	message1 := []byte("Hello, World")

	sigA, _ := signerA.Sign(message)
	sigB, _ := signerB.Sign(message)

	aggsig, _ := signerA.Aggregate(sigA, sigB)

	verifier, _ := signerA.GetVerifierFactory().FromArray(publicKeys)

	fmt.Println((verifier.Verify(message, aggsig) == nil))
	fmt.Println((verifier.Verify(message1, aggsig) == nil))
}

func Test_Threshold(t *testing.T) {
	//
	// 四个人都先生成自己的随机数Ki
	// k1, _ := threshold.GetRandom32Bytes()
	// k2, _ := threshold.GetRandom32Bytes()
	// k3, _ := threshold.GetRandom32Bytes()
	// k4, _ := threshold.GetRandom32Bytes()

	// 各方计算自己的 Ri = Ki*G，G代表基点

	// message := []byte("Example message")
	// curve := elliptic.P256()

	// // Simulate 3 participants
	// participants := 3
	// threshold := 2

	// // Step 1: Generate random Ki for each participant
	// var Kis [][]byte
	// for i := 0; i < participants; i++ {
	//     ki, _ := tss_sign.GetRandom32Bytes()
	//     Kis = append(Kis, ki)
	// }

	// // Step 2: Calculate Ri using Ki for each participant
	// var Ris [][]byte
	// publicKey := &ecdsa.PublicKey{Curve: curve}
	// for _, ki := range Kis {
	//     ri := tss_sign.GetRiUsingRandomBytes(publicKey, ki)
	//     Ris = append(Ris, ri)
	// }

	// // Step 3: Aggregate Ris to calculate R
	// R := tss_sign.GetRUsingAllRi(publicKey, Ris)

	// // Assume each participant's index and simulate the private key for simplicity
	// var Xs []*big.Int
	// for i := 1; i <= participants; i++ {
	//     Xs = append(Xs, big.NewInt(int64(i)))
	// }

	// // Simulate private key of a participant for simplicity
	// privateKey := &ecdsa.PrivateKey{D: big.NewInt(12345), PublicKey: *publicKey}

	// // Step 6: Calculate S(i) for each participant
	// var Sis [][]byte
	// for i := 0; i < threshold; i++ {
	//     coef := tss_sign.getCoefForXi(Xs, i, curve)
	//     xiWithCoef := tss_sign.GetXiWithcoef(Xs, i, privateKey)
	//     si := tss_sign.GetSiUsingKCRMWithCoef(Kis[i], publicKey.X.Bytes(), R, message, xiWithCoef)
	//     Sis = append(Sis, si)
	// }

	// // Step 7: Aggregate Sis to calculate S
	// S := tss_sign.GetSUsingAllSi(Sis)

	// // Generate the threshold signature
	// sig, err := tss_sign.GenerateTssSignSignature(S, R)
	// if err != nil {
	//     fmt.Printf("Error generating threshold signature: %v\n", err)
	//     return
	// }

	// fmt.Printf("Threshold signature: %x\n", sig)
}

func Test_tcrsa(t *testing.T) {
	Example()
}

const exampleK = 3
const exampleL = 5

const exampleHashType = stcryp.SHA256
const exampleSize = 2048
const exampleMessage = "Hello world"

func Example() {
	// First we need to get the values of K and L from somewhere.
	k := uint16(exampleK)
	l := uint16(exampleL)

	// Generate keys provides to us with a list of keyShares and the key metainformation.
	keyShares, keyMeta, err := tcrsa.NewKey(exampleSize, uint16(k), uint16(l), nil)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	// Then we need to prepare the document we want to sign, so we hash it and pad it using PKCS v1.15.
	docHash := sha256.Sum256([]byte(exampleMessage))
	docPKCS1, err := tcrsa.PrepareDocumentHash(keyMeta.PublicKey.Size(), stcryp.SHA256, docHash[:])
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	sigShares := make(tcrsa.SigShareList, l)
	var i uint16

	// Now we sign with at least k nodes and check immediately the signature share for consistency.
	for i = 0; i < l; i++ {
		sigShares[i], err = keyShares[i].Sign(docPKCS1, exampleHashType, keyMeta)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		if err := sigShares[i].Verify(docPKCS1, keyMeta); err != nil {
			panic(fmt.Sprintf("%v", err))
		}
	}

	// Having all the signature shares we needed, we join them to create a real signature.
	signature, err := sigShares.Join(docPKCS1, keyMeta)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	// Finally we check the signature with Golang's cryp/rsa PKCSv1.15 verification routine.
	if err := rsa.VerifyPKCS1v15(keyMeta.PublicKey, stcryp.SHA256, docHash[:], signature); err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	fmt.Println("ok")
	// Output: ok 测试server
}

func Test_Main(t *testing.T) {
	// client := hotstuff.NewReplicaClient(1)
}
