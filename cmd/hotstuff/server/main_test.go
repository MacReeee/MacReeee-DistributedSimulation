package main

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"distributed/consensus"
	"distributed/cryp"
	pb2 "distributed/pb"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	stcryp "crypto"

	"github.com/niclabs/tcrsa"
	blscryp "go.dedis.ch/dela/crypto"

	"go.dedis.ch/dela/crypto/bls"
)

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
	// Output: ok12
}

func Test_NewView(t *testing.T) {
	client := *hotstuff.NewReplicaClient(1)
	signer2 := cryp.NewSignerByID(2)
	var QC = GetValidQC(pb2.MsgType_NEW_VIEW)
	qcjson := hotstuff.QCMarshal(QC)
	sig, _ := signer2.NormSign(qcjson)
	NewViewMsg := &pb2.NewViewMsg{
		ProposalId: 2,
		MsgType:    pb2.MsgType_NEW_VIEW,
		ViewNumber: 1,
		Qc:         QC,
		Signature:  sig, //todo应当对QC进行签名，暂时省略1
	}
	_, err := client.NewView(context.Background(), NewViewMsg)
	log.Println(err)
}

// func Test_Server(t *testing.T) {
// 	client := *hotstuff.NewReplicaClient(1)
// 	signer2 := cryp.NewSignerByID(2)
// 	var QC = GetValidQC(pb.MsgType_NEW_VIEW)
// 	qcjson := hotstuff.QCMarshal(QC)
// 	sig, _ := signer2.NormSign(qcjson)
// 	NewViewMsg := &pb.NewViewMsg{
// 		ProposalId: 2,
// 		MsgType:    pb.MsgType_NEW_VIEW,
// 		ViewNumber: 1,
// 		Qc:         QC,
// 		Signature:  sig, //todo应当对QC进行签名，暂时省略123
// 	}
// 	Proposal, err := client.NewView(context.Background(), NewViewMsg)
// 	ProposalMsg := &pb.Proposal{
// 		Block: Proposal.GetBlock(),
// 		Qc:    Proposal.Block.GetQc(),
// 		// Aggqc:      nil,
// 		// ProposalId: 0,
// 		Proposer:   Proposal.GetProposer(),
// 		ViewNumber: Proposal.GetViewNumber(),
// 		Signature:  Proposal.GetSignature(),
// 		// Timestamp:  0,
// 		MsgType: Proposal.GetMsgType(),
// 	}
// 	log.Println("NewView错误: ", err)
// 	// fmt.Println(ProposalMsg)

// 	PrepareVoteMsg, err := client.Prepare(context.Background(), ProposalMsg)
// 	log.Println("Propose错误: ", err)
// 	// log.Println("消息格式：", Proposal.GetMsgType())
// 	// log.Println(PrepareVoteMsg)
// 	// time.Sleep(1 * time.Second)

// 	PreCommitMsg, err := client.VotePrepare(context.Background(), PrepareVoteMsg)
// 	log.Println("VotePrepare错误: ", err)
// 	// log.Println(PreCommitMsg.Signature)
// 	/*----------------------- 以上步骤测试完毕 -----------------------*/

// 	client.PreCommit(context.Background(), PreCommitMsg)
// }

func GetValidQC(msgType pb2.MsgType) *pb2.QC {
	signer1 := cryp.NewSignerByID(1)
	signer2 := cryp.NewSignerByID(2)
	signer3 := cryp.NewSignerByID(3)
	signer4 := cryp.NewSignerByID(4)

	// var signMsg = []byte(fmt.Sprintf("%d,%d,%x", pb.MsgType_PREPARE, 2, []byte("hashhash")))

	signature1, _ := signer1.Sign(msgType, 1, []byte("FFFFFFFFFFFF"))
	signature2, _ := signer2.Sign(msgType, 1, []byte("FFFFFFFFFFFF"))
	signature3, _ := signer3.Sign(msgType, 1, []byte("FFFFFFFFFFFF"))
	signature4, _ := signer4.Sign(msgType, 1, []byte("FFFFFFFFFFFF"))

	aggsig, aggpub, _ := signer2.ThreshMock([]int32{1, 2, 3, 4}, [][]byte{signature1, signature2, signature3, signature4})

	var QC = &pb2.QC{
		BlsSignature: aggsig,
		AggPubKey:    aggpub,
		Voter:        []int32{1, 2, 3, 4},
		MsgType:      msgType,
		ViewNumber:   1,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}
	return QC
}

func CreatePreCommitTestInput() *pb2.Precommit {

	signer2 := cryp.NewSignerByID(2)

	sig, _ := signer2.Sign(pb2.MsgType_PRE_COMMIT, 2, []byte("FFFFFFFFFFFF"))

	var PreCommitMsg = &pb2.Precommit{
		Id:         2,
		MsgType:    pb2.MsgType_PRE_COMMIT,
		ViewNumber: 2,
		Qc:         GetValidQC(pb2.MsgType_PRE_COMMIT),
		Signature:  sig,
	}
	return PreCommitMsg
}

func Test_PreCommit(t *testing.T) {
	client := *hotstuff.NewReplicaClient(1)
	PreCommitMsg := CreatePreCommitTestInput()
	_, err := client.PreCommit(context.Background(), PreCommitMsg)
	log.Println(err)
}

func Test_VotePreCommit(t *testing.T) {
	signer2 := cryp.NewSignerByID(2)
	sig, _ := signer2.Sign(pb2.MsgType_PRE_COMMIT_VOTE, 1, []byte("FFFFFFFFFFFF"))
	vote := pb2.VoteRequest{
		ViewNumber: 1,
		MsgType:    pb2.MsgType_PRE_COMMIT_VOTE,
		Voter:      2,
		Signature:  sig,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	client1 := *hotstuff.NewReplicaClient(1)
	CommitMsg, err := client1.VotePreCommit(context.Background(), &vote)
	log.Println("测试VotePreCommit的错误信息如下: ", err)
	log.Println("测试VotePreCommit的返回信息如下: ", CommitMsg)
	// _, err = client1.Debug(context.Background(), &pb.DebugMsg{})
	// fmt.Println(vote)
}

func Test_Commit(t *testing.T) {
	signer2 := cryp.NewSignerByID(2)

	QC := GetValidQC(pb2.MsgType_COMMIT)
	qcjson, _ := json.Marshal(QC)
	sig, _ := signer2.NormSign(qcjson)

	CommitMsg := &pb2.CommitMsg{
		Id:         2,
		MsgType:    pb2.MsgType_COMMIT,
		ViewNumber: 1,
		Qc:         QC,
		Hash:       []byte("FFFFFFFFFFFF"),
		Signature:  sig,
	}

	client1 := *hotstuff.NewReplicaClient(1)
	VoteRequestMsg, err := client1.Commit(context.Background(), CommitMsg)
	log.Println("测试Commit的错误信息如下: ", err)
	log.Println("测试Commit的返回信息如下: ", VoteRequestMsg)

	_, err = client1.Debug(context.Background(), &pb2.DebugMsg{})
}

func Test_VoteCommit(t *testing.T) {
	signer2 := cryp.NewSignerByID(2)
	sig, _ := signer2.Sign(pb2.MsgType_COMMIT_VOTE, 1, []byte("FFFFFFFFFFFF"))
	vote := pb2.VoteRequest{
		ViewNumber: 1,
		MsgType:    pb2.MsgType_COMMIT_VOTE,
		Voter:      2,
		Signature:  sig,
		Hash:       []byte("FFFFFFFFFFFF"),
	}
	client1 := *hotstuff.NewReplicaClient(1)
	_, err := client1.VoteCommit(context.Background(), &vote)
	log.Println("测试VoteCommit的错误信息如下: ", err)
}

func Test_Decide(t *testing.T) {
	signer2 := cryp.NewSignerByID(2)
	QC := GetValidQC(pb2.MsgType_COMMIT)
	qcjson, _ := json.Marshal(QC)
	sig, _ := signer2.NormSign(qcjson)

	cmd := []byte("CMD of View:  1")
	hasher := crypto.SHA256.New()
	hasher.Write(cmd)
	hash := []byte(fmt.Sprintf("%x", hasher.Sum(nil)))
	DecideMsg := &pb2.DecideMsg{
		MsgType:    pb2.MsgType_DECIDE,
		ViewNumber: 1,
		Qc:         QC,
		Hash:       hash,
		Signature:  sig,
	}
	client1 := *hotstuff.NewReplicaClient(1)
	_, err := client1.Decide(context.Background(), DecideMsg)
	log.Println("测试Decide的错误信息如下: ", err)
}
