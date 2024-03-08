package hotstuff

import (
	"context"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	"net"
	"strconv"
	stsync "sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DebugMode = true //调试模式开关

	wg       stsync.WaitGroup //中断控制
	StopFlag = false          //中断标志
)

var (
	ReplicaNum = 4
)

type ReplicaServer struct {
	mu stsync.Mutex
	// sigs      map[kyber.Point][]byte
	count     int
	threshold int
	wg        stsync.WaitGroup
	once      stsync.Once

	ID        int32
	PrepareQC *pb.QC
	LockedQC  *pb.QC

	pb.UnimplementedHotstuffServer
}

func NewReplicaServer(id int32) (*grpc.Server, *net.Listener) {
	var (
	//sync  = modules.MODULES.Synchronizer
	//cryp = modules.MODULES.Signer
	//chain = modules.MODULES.Chain
	)
	addr := fmt.Sprintf(":%d", id+4000)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("副本服务监听失败:", err)
	}
	server := grpc.NewServer()
	//初始化LockedQC和PrepareQC
	PrepareQC := &pb.QC{
		BlsSignature: []byte{28, 201, 16, 247, 213, 76, 151, 58, 250, 236, 79, 128, 122, 208, 217, 160, 143, 88, 54, 100, 139, 163, 76, 61, 181, 63, 167, 129, 66, 245, 88, 25, 21, 227, 11, 228, 66, 141, 175, 202, 151, 51, 11, 128, 65, 198, 218, 133, 123, 164, 170, 45, 207, 25, 255, 78, 238, 39, 217, 167, 127, 128, 89, 139},
		AggPubKey:    []byte{117, 17, 230, 255, 170, 193, 55, 5, 104, 254, 206, 140, 207, 13, 157, 251, 133, 127, 45, 101, 201, 13, 104, 232, 86, 99, 251, 120, 113, 181, 236, 203, 70, 82, 146, 114, 242, 245, 4, 11, 211, 137, 204, 26, 203, 162, 239, 53, 243, 152, 103, 109, 92, 66, 136, 231, 15, 124, 233, 177, 118, 254, 203, 130, 92, 210, 35, 180, 213, 26, 215, 163, 131, 111, 55, 119, 137, 211, 176, 127, 113, 180, 169, 35, 14, 211, 188, 213, 131, 150, 197, 222, 81, 79, 34, 226, 86, 112, 162, 123, 244, 203, 105, 228, 102, 225, 87, 44, 143, 133, 131, 146, 201, 123, 173, 133, 70, 157, 160, 103, 241, 161, 127, 114, 201, 70, 247, 75},
		Voter:        []int32{1, 2, 3, 4},
		MsgType:      pb.MsgType_PREPARE_VOTE,
		ViewNumber:   0,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}
	LockedQC := &pb.QC{
		BlsSignature: []byte{56, 102, 243, 138, 71, 19, 119, 120, 28, 242, 150, 51, 203, 31, 93, 78, 112, 144, 141, 87, 48, 195, 12, 60, 15, 160, 155, 17, 48, 87, 131, 51, 39, 119, 159, 49, 183, 198, 110, 188, 38, 200, 189, 59, 237, 239, 28, 28, 91, 84, 231, 78, 5, 75, 141, 214, 29, 174, 5, 46, 32, 26, 68, 5},
		AggPubKey:    []byte{117, 17, 230, 255, 170, 193, 55, 5, 104, 254, 206, 140, 207, 13, 157, 251, 133, 127, 45, 101, 201, 13, 104, 232, 86, 99, 251, 120, 113, 181, 236, 203, 70, 82, 146, 114, 242, 245, 4, 11, 211, 137, 204, 26, 203, 162, 239, 53, 243, 152, 103, 109, 92, 66, 136, 231, 15, 124, 233, 177, 118, 254, 203, 130, 92, 210, 35, 180, 213, 26, 215, 163, 131, 111, 55, 119, 137, 211, 176, 127, 113, 180, 169, 35, 14, 211, 188, 213, 131, 150, 197, 222, 81, 79, 34, 226, 86, 112, 162, 123, 244, 203, 105, 228, 102, 225, 87, 44, 143, 133, 131, 146, 201, 123, 173, 133, 70, 157, 160, 103, 241, 161, 127, 114, 201, 70, 247, 75},
		Voter:        []int32{1, 2, 3, 4},
		MsgType:      pb.MsgType_PRE_COMMIT_VOTE,
		ViewNumber:   0,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}

	var thresh int
	if DebugMode {
		thresh = 1
	} else {
		thresh = 3
	}
	replicaserver := &ReplicaServer{
		threshold: thresh,
		count:     0,
		wg:        stsync.WaitGroup{},
		once:      stsync.Once{},
		PrepareQC: PrepareQC,
		LockedQC:  LockedQC,
		ID:        id,
	}

	pb.RegisterHotstuffServer(server, replicaserver)
	go NextView(replicaserver) // debug模式下预防视图超时阻塞，正式使用替换成NextView函数
	//log.Println("副本服务启动成功: ", addr)
	modules.MODULES.ReplicaServer = server
	modules.MODULES.ReplicaServerStruct = replicaserver
	return server, &listener
}

func NewReplicaClient(id int32) *pb.HotstuffClient {
	conn, err := grpc.Dial(fmt.Sprintf(":%d", id+4000), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("副本", id, "连接失败:", err)
		return nil
	}
	client := pb.NewHotstuffClient(conn)
	log.Println("连接副本", id, "成功")
	// client.NewView(context.Background(), &pb.NewViewMsg{})
	modules.MODULES.ReplicaClient[id] = &client
	return &client
}

func (s *ReplicaServer) Prepare(ctx context.Context, Proposal *pb.Proposal) (*pb.VoteRequest, error) {
	log.Println("接收到来自 ", Proposal.Proposer, " 的提案")

	//用以控制台控制中断
	if StopFlag {
		wg.Add(1)
		wg.Wait()
	}
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	if ok, err := MatchingMsg(Proposal.MsgType, Proposal.ViewNumber, pb.MsgType_PREPARE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	//log.Println(Proposal)
	if !s.SafeNode(Proposal.Block, Proposal.Qc) {
		return nil, fmt.Errorf("proposal is not safe")
	}
	//视图成功并退出
	ViewSuccess(sync)
	fmt.Println("\n", "视图", *sync.ViewNumber(), "的log: ")

	// 临时存储区块
	chain.StoreToTemp(Proposal.Block)

	//对提案进行签名
	sig, err := cryp.Sign(pb.MsgType_PREPARE_VOTE, *sync.ViewNumber(), Proposal.Block.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PrepareVoteMsg := &pb.VoteRequest{
		// ProposalId: Proposal.ProposalId,
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       Proposal.Block.Hash,
		MsgType:    pb.MsgType_PREPARE_VOTE,
	}

	var leader pb.HotstuffClient
	if DebugMode {
		leader = *modules.MODULES.ReplicaClient[1]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
		log.Println("视图 ", *sync.ViewNumber(), " 的领导者是: ", sync.GetLeader())
	}
	go leader.VotePrepare(context.Background(), PrepareVoteMsg)
	return PrepareVoteMsg, nil
}

func (s *ReplicaServer) VotePrepare(ctx context.Context, vote *pb.VoteRequest) (*pb.Precommit, error) {
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)

	//如果已经触发阈值条件，不在接收后续的Prepare投票
	_, _, once := sync.GetVoter(pb.MsgType_PREPARE_VOTE)
	if once.IsDone() {
		return nil, nil
	}

	log.Println("接收到来自 ", vote.Voter, " 的Prepare投票")
	if ok, err := MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_PREPARE_VOTE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		return nil, fmt.Errorf("prepare vote signature is not valid")
	}
	sync.StoreVote(pb.MsgType_PREPARE_VOTE, vote)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		voters, sigs, once := sync.GetVoter(pb.MsgType_PREPARE_VOTE)
		var PreCommitMsg = &pb.Precommit{}
		once.Do(func() { //调用其他副本的PreCommit
			// 合成QC
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb.MsgType_PREPARE_VOTE,
				ViewNumber:   *sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)

			PreCommitMsg = &pb.Precommit{
				Id:         s.ID,
				MsgType:    pb.MsgType_PRE_COMMIT,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC,
				Hash:       vote.Hash,
				Signature:  sig,
			}
			for _, client := range modules.MODULES.ReplicaClient {
				go (*client).PreCommit(context.Background(), PreCommitMsg)
			}
			s.count = 0 //重置计数器
		})
		return PreCommitMsg, nil
	}
	return nil, nil
}

func (s *ReplicaServer) PreCommit(ctx context.Context, PrecommitMsg *pb.Precommit) (*pb.VoteRequest, error) {
	log.Println("接收到来自 ", PrecommitMsg.Id, " 的PreCommit消息")
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	if ok, err := MatchingMsg(PrecommitMsg.MsgType, PrecommitMsg.ViewNumber, pb.MsgType_PRE_COMMIT, *sync.ViewNumber()); !ok {
		return nil, err
	}

	s.PrepareQC = PrecommitMsg.Qc //更新PrepareQC

	sync.TimerReset()
	chain.Store(chain.GetBlockFromTemp(PrecommitMsg.Hash)) // 存储这一轮的区块

	sig, err := cryp.Sign(pb.MsgType_PRE_COMMIT_VOTE, *sync.ViewNumber(), PrecommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PreCommitVoteMsg := &pb.VoteRequest{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       PrecommitMsg.Hash,
		MsgType:    pb.MsgType_PRE_COMMIT_VOTE,
	}
	var leader pb.HotstuffClient
	if DebugMode {
		leader = *modules.MODULES.ReplicaClient[1]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	}
	go leader.VotePreCommit(context.Background(), PreCommitVoteMsg)
	return PreCommitVoteMsg, nil
}

func (s *ReplicaServer) VotePreCommit(ctx context.Context, vote *pb.VoteRequest) (*pb.CommitMsg, error) {
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)

	//如果已经触发阈值条件，不在接收后续的PreCommit投票
	_, _, once := sync.GetVoter(pb.MsgType_PRE_COMMIT_VOTE)
	if once.IsDone() {
		return nil, nil
	}

	log.Println("接收到来自 ", vote.Voter, " 的PreCommit投票")
	if ok, err := MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_PRE_COMMIT_VOTE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		log.Println("Pre_Commit Vote 的签名验证失败")
		return nil, fmt.Errorf("Pre_Commit Vote 的签名验证失败")
	}
	sync.StoreVote(pb.MsgType_PRE_COMMIT_VOTE, vote)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		voters, sigs, once := sync.GetVoter(pb.MsgType_PRE_COMMIT_VOTE)
		var CommitMsg = &pb.CommitMsg{}
		once.Do(func() { //调用其他副本的Commit
			// 合成QC
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb.MsgType_PRE_COMMIT_VOTE,
				ViewNumber:   *sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			CommitMsg = &pb.CommitMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_COMMIT,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC,
				Hash:       vote.Hash,
				Signature:  sig,
			}
			for _, client := range modules.MODULES.ReplicaClient {
				go (*client).Commit(context.Background(), CommitMsg)
			}
			s.count = 0 //重置计数器
		})
		return CommitMsg, nil
	}
	return nil, nil
}

func (s *ReplicaServer) Commit(ctx context.Context, CommitMsg *pb.CommitMsg) (*pb.VoteRequest, error) {
	log.Println("接收到来自 ", CommitMsg.Id, " 的Commit消息")
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	if ok, err := MatchingMsg(CommitMsg.MsgType, CommitMsg.ViewNumber, pb.MsgType_COMMIT, *sync.ViewNumber()); !ok {
		return nil, err
	}

	s.LockedQC = CommitMsg.Qc //更新LockedQC

	sync.TimerReset() //重置计时器

	Sig, err := cryp.Sign(pb.MsgType_COMMIT_VOTE, *sync.ViewNumber(), CommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	CommitVoteMsg := &pb.VoteRequest{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  Sig,
		Hash:       CommitMsg.Hash,
		MsgType:    pb.MsgType_COMMIT_VOTE,
	}
	var leader pb.HotstuffClient
	if DebugMode {
		leader = *modules.MODULES.ReplicaClient[1]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	}
	go leader.VoteCommit(context.Background(), CommitVoteMsg)
	return CommitVoteMsg, nil
}

func (s *ReplicaServer) VoteCommit(ctx context.Context, vote *pb.VoteRequest) (*pb.DecideMsg, error) {
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	log.Println("接收到来自 ", vote.Voter, " 的Commit投票")

	//如果已经触发阈值条件，不在接收后续的Commit投票
	_, _, once := sync.GetVoter(pb.MsgType_COMMIT_VOTE)
	if once.IsDone() {
		return nil, nil
	}

	if ok, err := MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_COMMIT_VOTE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		return nil, fmt.Errorf("commit vote signature is not valid")
	}
	sync.StoreVote(pb.MsgType_COMMIT_VOTE, vote)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		voters, sigs, once := sync.GetVoter(pb.MsgType_COMMIT_VOTE)
		var DecideMsg = &pb.DecideMsg{}
		once.Do(func() { //调用其他副本的Decide
			// 合成QC
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb.MsgType_COMMIT_VOTE,
				ViewNumber:   *sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			DecideMsg = &pb.DecideMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_DECIDE,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC, //暂未获取
				Hash:       vote.Hash,
				Signature:  sig, //暂未获取
			}
			for _, client := range modules.MODULES.ReplicaClient {
				go (*client).Decide(context.Background(), DecideMsg)
			}
			s.count = 0 //重置计数器
		})
		return DecideMsg, nil
	}
	return nil, nil
}

func (s *ReplicaServer) Decide(ctx context.Context, DecideMsg *pb.DecideMsg) (*pb.NewViewMsg, error) {
	log.Println("接收到来自 ", DecideMsg.Id, " 的Decide消息")
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	if ok, err := MatchingMsg(DecideMsg.MsgType, DecideMsg.ViewNumber, pb.MsgType_DECIDE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	sync.TimerReset() //重置计时器
	sig, err := cryp.Sign(pb.MsgType_NEW_VIEW, *sync.ViewNumber(), DecideMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	curViewNumber++
	newestBlock := chain.GetBlock(DecideMsg.Hash)

	//剪枝，并存储已经稳定上链的区块
	chain.PruneBlock(chain.GetBlock(newestBlock.ParentHash), newestBlock)

	NewViewMsg := &pb.NewViewMsg{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       DecideMsg.Hash,
		MsgType:    pb.MsgType_NEW_VIEW,
		Qc:         s.PrepareQC,
	}

	var leader pb.HotstuffClient
	if DebugMode {
		leader = *modules.MODULES.ReplicaClient[1]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()+1]
	}
	go leader.NewView(context.Background(), NewViewMsg)
	return NewViewMsg, nil
}

func (s *ReplicaServer) NewView(ctx context.Context, NewViewMsg *pb.NewViewMsg) (*pb.Proposal, error) {
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("接收到来自 ", NewViewMsg.Voter, " 的NewView消息")

	//如果已经触发阈值条件，不在接收后续的NewView消息
	_, _, once := sync.GetVoter(pb.MsgType_NEW_VIEW)
	if once.IsDone() {
		return nil, nil
	}

	if ok, err := MatchingMsg(NewViewMsg.MsgType, NewViewMsg.ViewNumber, pb.MsgType_NEW_VIEW, *sync.ViewNumber()); !ok {
		return nil, err
	}
	//签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", NewViewMsg.MsgType, NewViewMsg.ViewNumber, NewViewMsg.Hash))
	if !cryp.Verify(NewViewMsg.Voter, msg, NewViewMsg.Signature) {
		log.Println("NewView消息签名验证失败")
		return nil, fmt.Errorf("newview msg signature is not valid")
	}
	sync.StoreVote(pb.MsgType_NEW_VIEW, nil, NewViewMsg)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		_, _, once := sync.GetVoter(pb.MsgType_NEW_VIEW)
		var ProposalMsg = &pb.Proposal{}
		once.Do(func() { //调用其他副本的Propose
			// 获取QC
			HighQC := sync.HighQC()
			// 对QC签名作为自己的签名
			qcjson := QCMarshal(HighQC)
			sig, _ := cryp.NormSign(qcjson)
			// 创建区块
			block := chain.CreateBlock(NewViewMsg.Qc.BlockHash, *sync.ViewNumber()+1, HighQC, []byte("CMD of View: "+strconv.Itoa(int(*sync.ViewNumber()+1))), s.ID)
			ProposalMsg = &pb.Proposal{
				Block: block,
				Qc:    HighQC,
				// Aggqc: nil, //hotstuff中用不到
				// ProposalId: 0, //暂未获取，未考虑清楚是否需要该字段
				Proposer:   s.ID,
				ViewNumber: *sync.ViewNumber(),
				Signature:  sig,
				// Timestamp:  0, //暂时用不到
				MsgType: pb.MsgType_PREPARE,
			}
			for _, client := range modules.MODULES.ReplicaClient {
				go (*client).Prepare(context.Background(), ProposalMsg)
			}

			s.count = 0 //重置计数器
		})
		return ProposalMsg, nil
	}
	return nil, nil
}

//如果连续失败多个视图，如何保障节点之间的视图对齐？

func (s *ReplicaServer) SafeNode(block *pb.Block, qc *pb.QC) bool {
	condition1 := (string(block.ParentHash) == string(qc.BlockHash))         //检查是否是父区块的子区块
	condition2 := (string(block.ParentHash) == string(s.LockedQC.BlockHash)) //安全性
	condition3 := (qc.ViewNumber > s.LockedQC.ViewNumber)                    //活性
	return condition1 && (condition2 || condition3)
}

func NextView(s *ReplicaServer) { //所有的wait for阶段超时都会调用这个函数，//todo 记得go调用
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	for {
		select {
		case <-sync.Timeout():
			var QC = s.PrepareQC
			//对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			var leader pb.HotstuffClient
			if DebugMode {
				leader = *modules.MODULES.ReplicaClient[1]
			} else {
				leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
			}
			NewViewMsg := &pb.NewViewMsg{
				ProposalId: s.ID,
				MsgType:    pb.MsgType_NEW_VIEW,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC,
				Signature:  sig, //todo应当对QC进行签名，暂时省略
			}
			leader.NewView(context.Background(), NewViewMsg)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

/*-------- debug functions --------*/

func NextViewDebugMode(s *ReplicaServer) { //所有的wait for阶段超时都会调用这个函数，//todo 记得go调用
	var (
		sync = modules.MODULES.Synchronizer
		//cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	i := 1
	for {
		<-sync.Timeout()
		// log.Println("捕获到视图超时，已捕获超时次数: ", i, "当前视图号: ", sync.ViewNumber())
		i++
	}
}
