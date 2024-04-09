package hotstuff

import (
	"context"
	d "distributed/dependency"
	"distributed/modules"
	pb2 "distributed/pb"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"os"
	stsync "sync"
	"time"
)

type State string

const (
	Idle        State = "Idle"
	Prepare     State = "Prepare"
	Precommit   State = "Precommit"
	Commit      State = "Commit"
	inConsensus State = "Consensusing"
	Switching   State = "Switching"
)

func MatchingMsg(Type pb2.MsgType, ViewNumber int64, TarType pb2.MsgType, TarviewNumber int64) (bool, error) {
	condition1 := (Type == TarType)
	if !condition1 {
		return false, fmt.Errorf("消息类型不匹配")
	}
	condition2 := (ViewNumber == TarviewNumber)
	if !condition2 {
		log.Println("\n Type:", Type, "ViewNumber:", ViewNumber, "TarType:", TarType, "TarviewNumber:", TarviewNumber)
		return false, fmt.Errorf("视图号不匹配")
	}
	return condition1 && condition2, nil
}

func QCMarshal(qc *pb2.QC) []byte {
	qcjson, err := json.Marshal(qc)
	if err != nil {
		log.Println("json序列化失败:", err)
	}
	return qcjson
}

func writeFatalErr(errinfo string) {
	file, err := os.OpenFile("err.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("打开文件失败:", err)
	}
	defer file.Close()
	_, err = file.WriteString(errinfo)
	if err != nil {
		log.Println("写入文件失败:", err)
	}
}

type ReplicaServer struct {
	mu        stsync.Mutex
	count     int
	threshold int
	wg        stsync.WaitGroup
	once      stsync.Once
	ctx       context.Context

	state State
	cond  *stsync.Cond

	ID             int32
	PrepareQC      *pb2.QC
	LockedQC       *pb2.QC
	lastVote       int64
	TempViewNumber int64
	TimeoutRecord  int

	pb2.UnimplementedHotstuffServer
}

func NewReplicaServer(id int32, ctx context.Context) (*grpc.Server, *net.Listener) {
	addr := fmt.Sprintf(":%d", id+4000)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("副本服务监听失败:", err)
	}
	server := grpc.NewServer()
	//初始化LockedQC和PrepareQC
	PrepareQC := &pb2.QC{
		BlsSignature: []byte("GenesisPrepareQC"),
		AggPubKey:    []byte("GenesisPrepareQCSignature"),
		Voter:        []int32{},
		MsgType:      pb2.MsgType_PREPARE_VOTE,
		ViewNumber:   0,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}
	LockedQC := &pb2.QC{
		BlsSignature: []byte("GenesisLockedQC"),
		AggPubKey:    []byte("GenesisLockedQCSignature"),
		Voter:        []int32{},
		MsgType:      pb2.MsgType_PRE_COMMIT_VOTE,
		ViewNumber:   0,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}

	var thresh int
	if d.Configs.BuildInfo.DebugMode {
		thresh = d.Configs.BuildInfo.Threshold
	} else {
		thresh = 3
	}

	replicaserver := &ReplicaServer{
		threshold:      thresh,
		count:          0,
		wg:             stsync.WaitGroup{},
		once:           stsync.Once{},
		ctx:            ctx,
		state:          Idle,
		PrepareQC:      PrepareQC,
		LockedQC:       LockedQC,
		lastVote:       0,
		TempViewNumber: 0,
		ID:             id,
		TimeoutRecord:  0,
	}
	replicaserver.cond = stsync.NewCond(&replicaserver.mu)

	pb2.RegisterHotstuffServer(server, replicaserver)
	go replicaserver.NextView() // debug模式下预防视图超时阻塞，正式使用替换成NextView函数
	//log.Println("副本服务启动成功: ", addr)
	modules.MODULES.ReplicaServer = server
	modules.MODULES.ReplicaServerStruct = replicaserver
	return server, &listener
}

func NewReplicaClient(id int32) *pb2.HotstuffClient {
	target := fmt.Sprintf(":%d", id+4000)
	if d.Configs.BuildInfo.DockerMode {
		host := "node" + fmt.Sprintf("%d", id)
		target = fmt.Sprintf("%v:%d", host, id+4000)
	}
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("副本", id, "连接失败:", err)
		return nil
	}
	client := pb2.NewHotstuffClient(conn)
	log.Println("连接副本", id, "成功")
	modules.MODULES.ReplicaClient[id] = &client
	return &client
}

func (s *ReplicaServer) GetBlock(ctx context.Context, b *pb2.SyncBlock) (*pb2.Block, error) {
	chain := modules.MODULES.Chain
	return chain.GetBlock([]byte(b.Hash)), nil
}

func (s *ReplicaServer) SafeNode(block *pb2.Block, qc *pb2.QC) bool {
	condition1 := (string(block.ParentHash) == string(qc.BlockHash))         //检查是否是QC中描述的父区块的子区块
	condition2 := (string(block.ParentHash) == string(s.LockedQC.BlockHash)) //安全性
	condition3 := (qc.ViewNumber > s.LockedQC.ViewNumber)                    //活性
	return condition1 && (condition2 || condition3)
}

func (s *ReplicaServer) waitForIdle() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for s.state != Idle {
		s.cond.Wait()
	}
}

func (s *ReplicaServer) WaitForState(state State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for s.state != state {
		s.cond.Wait()
	}
}

func (s *ReplicaServer) SetState(state State) {
	defer func() {
		r := recover()
		if r != nil {
			FuncName := "SetState"
			log.Println(FuncName, "函数异常", r)
			panic(r)
		}
	}()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
	s.cond.Broadcast()
}

func (s *ReplicaServer) StopVoting(viewnumber int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastVote < viewnumber {
		s.lastVote = viewnumber
	}
}

func (s *ReplicaServer) RecordTimeOutLogToFile() {
	file, err := os.OpenFile("./output/timeout.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("打开文件失败:", err)
	}
	defer file.Close()
	// 年/月/日 时:分:秒 节点[ID] 视图号[ViewNumber] : 当前超时次数TimeoutRecord
	now := time.Now().Format("2006/01/02 15:04:05")
	record := fmt.Sprintf("%s 节点[%d] 视图号[%d] : 当前超时次数 %d\n", now, s.ID, s.TempViewNumber, s.TimeoutRecord)
	_, err = file.WriteString(record)
	if err != nil {
		log.Println("写入文件失败:", err)
	}
}
