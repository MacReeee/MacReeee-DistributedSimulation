package hotstuff

import "distributed/hotstuff/pb"

func MatchingMsg(Type pb.MsgType, ViewNumber int64, TarType pb.MsgType, TarviewNumber int64) bool {
	return Type == TarType && ViewNumber == TarviewNumber
}

func SafeNode(block *pb.Block, qc *pb.QC) bool {
	return (string(block.ParentHash) == string(qc.BlockHash)) && //检查是否是父区块的子区块
		(string(block.ParentHash) == string(LockedQC.BlockHash) || //安全性
			qc.ViewNumber > LockedQC.ViewNumber) //活性
}

func Sign(msg []byte) []byte {
	return nil
}
