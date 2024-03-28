// Deprecated: 该文件中的变量已经不再使用
package hotstuff

import (
	"distributed/pb"

	"go.dedis.ch/kyber/v3"
)

var (
	NumReplicas int32 = 4 //副本数量
	ReplicaID   int32 = 1 //副本ID
)
var LockedQC *pb.QC         //LockedQC
var PrepareQC *pb.QC        //PrepareQC
var curViewNumber int64 = 1 //当前视图编号
// var TempBlockMap = make(map[string]*pb.Block) //临时存储的区块，用于存储收到的Prepare消息中的区块

var PrivateKey = []byte("V5PGwk21S2gxQ2M2Madm079kH6bgvISCc8ypdRgDG7Y=") //私钥
var PublicKey = []byte("a18jcN2ymGwHW6sfM+36Z6WH6eEjmD1vyTHejDd/V4sIyd" +
	"MUSzID2emC0A95mVcj9nJzgPSYdzGkXmdzrlfb6hKB+flqtJPPA/gwDT2ym5vPVIEJfPz4W9" +
	"Xbp2kaU07LIgoGTXbzCory9Xw0MYu/zW04iKukoXq/FTd56YNCsiE=") //公钥

type KeyPair struct {
	PublicKey  string `json:"PublicKey"`
	PrivateKey string `json:"PrivateKey"`
}

var KeyPairs = []KeyPair{
	{
		PublicKey:  "a18jcN2ymGwHW6sfM+36Z6WH6eEjmD1vyTHejDd/V4sIydMUSzID2emC0A95mVcj9nJzgPSYdzGkXmdzrlfb6hKB+flqtJPPA/gwDT2ym5vPVIEJfPz4W9Xbp2kaU07LIgoGTXbzCory9Xw0MYu/zW04iKukoXq/FTd56YNCsiE=",
		PrivateKey: "V5PGwk21S2gxQ2M2Madm079kH6bgvISCc8ypdRgDG7Y=",
	},
	{
		PublicKey:  "ScfUmUSbfnRlWIJrOjeTWF5dN0k8glyIWCHV265sMS5rzJGx7SJMS5JsL+fcYIIE5ZWv8EWUy5xd3im5aGGLgC0V8TuN56seS0n4BIoemElR3DHAJJuJrV2XvdyJXB85LMRUCIVolpWLuq7xZZtJ4RNBVGgptt+Vkgx+aVMFyrs=",
		PrivateKey: "iGVAFluZ84ZfxrehT/WJ1DddjZutlMO4Bo8OUnE4NDM=",
	},
	{
		PublicKey:  "LUis6weGjZcbX6/tItwLai2qTz7mJw7f51V6jUWEHFtxK81IamCr77xX3JsD51RwLV0mwRQha7VyhyqLG5YWJTmDzqmUlw30RLZOK6RDB+A8WvDydWHjB8d9E31jJyNzc2ibLWY8UnOWWvqjJUN7",
		PrivateKey: "EfMDzg8Mnz6EkOeG2Viv7+w4KfpK1n/30o3NLNOd7iY=",
	},
	{
		PublicKey:  "KaFnrtCeRhQCW0XBQx9JMb0QN0KyezsaQC19/l9PrIRR3QGRBu2t3x1qjD95K388S9mPL+bkxn/AzIr8Dhq2LQwdeXZ9tBsCqoEVUfPJLWlGyksWhWtsjiP7AaPyqdYabFwYKAEsgf0ZYw0fGtIphYwQQeNtqb43+LSYyTpCuDQ=",
		PrivateKey: "OEdEfUl0xtozvgSoIivWcBwwAD3x7zg/R9X1SiSSmTg=",
	},
	{
		PublicKey:  "R6/0xrXPzNZvIxLU7rpIGNtDRegS/Pnx2AG2zkS2oW4irtp/TSH/ikwgZYLIy4+T+c//sds31sVAPeIATAjzOCYWR1BxnNn67klpweF0xF893P/AN8e2gzW/nXUC9nlOMTNgG5r7ODcrNnm7qIcYm8NXa8hrKzKP4/BBI3Ds/jI=",
		PrivateKey: "fOCzozx+9b6iPP/D6TiXokT4uwpWh/IhgkoEpGuGP9c=",
	},
	{
		PublicKey:  "eZpYENCrRx090YA7BHAdmMCOJ6UWTlOV/gFWyqdgh3Y72bVQXZUy0Bg+P7na0JXOVSNfNG2rhEA3bLV2gWEgQ0pU29vEnlyEe6j8n8/MqvPXTrEL+L3JcSNDan3yEJIaCxAvsZ93z9/RH1TUrwEtVhF9oDiKQ6Y0Voma/0VnAAY=",
		PrivateKey: "N/bg70Q9gK4/IxePFUgxayRd6/VDwngYtVGrc7wo7aM=",
	},
	{
		PublicKey:  "DiP7vyWWAsyhWLPyKLb6jAEYqmHHU7Sp2vJ4XnmMwptnw3S1qqEWiYDvlhmarygWM+K4gfTKP3op53PYVpiKZjBbC7WaMLrZOs1BVECh7dGdrqp6rQNGC35E2xuIKKC1UTAs9qV4RIudU849JU45tuextimnJfW/kWOD32Vm404=",
		PrivateKey: "DyGPC8AGDjbP2fiKU8WoAV9Bowl5CzbzqRtBR+nvKI8=",
	},
	{
		PublicKey:  "fqjr+LP9HIq7anib8y3AfP+ajXXbsxeiRdbkBtmX3YcZtEByGRhw1MHQy85oKvNpfEyp4x4qNAjmB/vu2yXC/QTz+cEXqZXRgM4Ct5EDWYzAl3lZeHiFMsxuTSrhxnS1im/+D8qmRjJeCLcaLSK80n8kzZhOOv1aQg+5G8cbgqM=",
		PrivateKey: "YIzuBsPOOHUCdlOD9PCpaSiPojrfXkl7g3jz0aB/LTw=",
	},
	{
		PublicKey:  "VPcBD3JDe9h7AGPqtYIb2Bc6JU8tLh/75odrJ1rTss5IBDiUJlhXQl2tLaKNXYQC8nOjJZnpZa0AB6h+yKkEQko8DluTQOCxta0/MLOlDikDIBm1z9bxkBJu69lDw3nZc0Ol83gljtn9L8CIadp7VUPRFRQS5yVUlqGkBCS/3OM=",
		PrivateKey: "J9VhDC6IhYYFAm9AbHm6+p+IDLMzW5/KLey/Xt2WN/0=",
	},
	{
		PublicKey:  "GcLGIftBwwVmazTXCl6sWx25xMgD0ozIk50q8IS308FNvFhnQnMWpP4VRAp3xKyxhfg0VsU5O/ncxTH+nUNZNX29HP9Lclvx8Z2OjXlwBZiu2pJ8cG/MPbSHUtsELUD+jbO8aNy5GXbXfCjLUM8lP8kpd2lyTDwcCzRwmIXaoUY=",
		PrivateKey: "O671m2q7jluBEW+u4DKrNaWuwSa8uSmHrCul+2X6cyI=",
	},
}

type Replica struct {
	// 副本的ID。
	ID uint32
	// 副本的公钥。
	PublicKey kyber.Point
}

type Replicas struct {
	REPS    []*Replica
	REPSMap map[uint32]*Replica
}

func NewReplicas() *Replicas {
	return &Replicas{
		REPS:    make([]*Replica, 0),
		REPSMap: make(map[uint32]*Replica),
	}
}
