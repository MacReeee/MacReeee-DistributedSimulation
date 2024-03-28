package hotstuff

import (
	"context"
	d "distributed/dependency"
	"distributed/middleware"
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

func ViewSuccess(sync middleware.Synchronizer) {
	_, success := sync.GetContext()
	sync.ViewNumberPP()
	once := sync.GetOnly()
	once.Do(success)
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
	mu stsync.Mutex
	// sigs      map[kyber.Point][]byte
	count     int
	threshold int
	wg        stsync.WaitGroup
	once      stsync.Once

	state State
	cond  *stsync.Cond

	ID        int32
	PrepareQC *pb2.QC
	LockedQC  *pb2.QC
	lastVote  int64

	pb2.UnimplementedHotstuffServer
}

func NewReplicaServer(id int32) (*grpc.Server, *net.Listener) {
	addr := fmt.Sprintf(":%d", id+4000)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("副本服务监听失败:", err)
	}
	server := grpc.NewServer()
	//初始化LockedQC和PrepareQC
	PrepareQC := &pb2.QC{
		BlsSignature: []byte{28, 201, 16, 247, 213, 76, 151, 58, 250, 236, 79, 128, 122, 208, 217, 160, 143, 88, 54, 100, 139, 163, 76, 61, 181, 63, 167, 129, 66, 245, 88, 25, 21, 227, 11, 228, 66, 141, 175, 202, 151, 51, 11, 128, 65, 198, 218, 133, 123, 164, 170, 45, 207, 25, 255, 78, 238, 39, 217, 167, 127, 128, 89, 139},
		AggPubKey:    []byte{117, 17, 230, 255, 170, 193, 55, 5, 104, 254, 206, 140, 207, 13, 157, 251, 133, 127, 45, 101, 201, 13, 104, 232, 86, 99, 251, 120, 113, 181, 236, 203, 70, 82, 146, 114, 242, 245, 4, 11, 211, 137, 204, 26, 203, 162, 239, 53, 243, 152, 103, 109, 92, 66, 136, 231, 15, 124, 233, 177, 118, 254, 203, 130, 92, 210, 35, 180, 213, 26, 215, 163, 131, 111, 55, 119, 137, 211, 176, 127, 113, 180, 169, 35, 14, 211, 188, 213, 131, 150, 197, 222, 81, 79, 34, 226, 86, 112, 162, 123, 244, 203, 105, 228, 102, 225, 87, 44, 143, 133, 131, 146, 201, 123, 173, 133, 70, 157, 160, 103, 241, 161, 127, 114, 201, 70, 247, 75},
		Voter:        []int32{1, 2, 3, 4},
		MsgType:      pb2.MsgType_PREPARE_VOTE,
		ViewNumber:   0,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}
	LockedQC := &pb2.QC{
		BlsSignature: []byte{56, 102, 243, 138, 71, 19, 119, 120, 28, 242, 150, 51, 203, 31, 93, 78, 112, 144, 141, 87, 48, 195, 12, 60, 15, 160, 155, 17, 48, 87, 131, 51, 39, 119, 159, 49, 183, 198, 110, 188, 38, 200, 189, 59, 237, 239, 28, 28, 91, 84, 231, 78, 5, 75, 141, 214, 29, 174, 5, 46, 32, 26, 68, 5},
		AggPubKey:    []byte{117, 17, 230, 255, 170, 193, 55, 5, 104, 254, 206, 140, 207, 13, 157, 251, 133, 127, 45, 101, 201, 13, 104, 232, 86, 99, 251, 120, 113, 181, 236, 203, 70, 82, 146, 114, 242, 245, 4, 11, 211, 137, 204, 26, 203, 162, 239, 53, 243, 152, 103, 109, 92, 66, 136, 231, 15, 124, 233, 177, 118, 254, 203, 130, 92, 210, 35, 180, 213, 26, 215, 163, 131, 111, 55, 119, 137, 211, 176, 127, 113, 180, 169, 35, 14, 211, 188, 213, 131, 150, 197, 222, 81, 79, 34, 226, 86, 112, 162, 123, 244, 203, 105, 228, 102, 225, 87, 44, 143, 133, 131, 146, 201, 123, 173, 133, 70, 157, 160, 103, 241, 161, 127, 114, 201, 70, 247, 75},
		Voter:        []int32{1, 2, 3, 4},
		MsgType:      pb2.MsgType_PRE_COMMIT_VOTE,
		ViewNumber:   0,
		BlockHash:    []byte("FFFFFFFFFFFF"),
	}

	var thresh int
	if d.DebugMode {
		thresh = 3
	} else {
		thresh = 3
	}

	replicaserver := &ReplicaServer{
		threshold: thresh,
		count:     0,
		wg:        stsync.WaitGroup{},
		once:      stsync.Once{},
		state:     Idle,
		PrepareQC: PrepareQC,
		LockedQC:  LockedQC,
		lastVote:  0,
		ID:        id,
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
	conn, err := grpc.Dial(fmt.Sprintf(":%d", id+4000), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("副本", id, "连接失败:", err)
		return nil
	}
	client := pb2.NewHotstuffClient(conn)
	log.Println("连接副本", id, "成功")
	// client.NewView(context.Background(), &pb.NewViewMsg{})
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
