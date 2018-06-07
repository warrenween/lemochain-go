package dpovp

import (
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"bytes"
)

type AddrNodeIDMapping struct{
	Addr common.Address
	Pubkey []byte
}

// Get all sorted nodes that who can produce blocks
func GetAllSortedCoreNodes() []AddrNodeIDMapping {
	// TODO
	result := make([]AddrNodeIDMapping, 0)
	var addr = common.HexToAddress(`0xb2a6935aedca3c64d1a98787f48ef4bc010b09d7`)
	var pubKey = common.Hex2Bytes(`8e88095f606e713bb52be8268ab7fe94f9656c0cf7a86d02dcaf141c7d54ed5ae9a3eafad69a2209a2fd645b33819df9ac46f4c251269671a847a7b01adf24b517`)
	result = append(result, AddrNodeIDMapping{addr, pubKey})

	addr = common.HexToAddress(`0xc848fe1d6b93c0ec640b3b2469c40378c8adbb5a`)
	pubKey = common.Hex2Bytes(`0cfb8b4d451fe60b86e4c258632e05e44c044d29bce69a3148b7949ace4320804b95f313c4e9cb69c93861fdbee62177d0d30bda1663c2f580bb6bdd196b244b`)
	result = append(result, AddrNodeIDMapping{addr, pubKey})

	addr = common.HexToAddress(`0x8353a1ce6b1a77a6863de1e0a764a6f3e58d3b0b`)
	pubKey = common.Hex2Bytes(`5a9e396255b7573f6be509ec2f0956fbc9b2dc2644a4e54c3ec02bc835770360c22aebca06431b7e8be5b2473c9addac18b2cb679cbc088274a5428df25d8849`)
	result = append(result, AddrNodeIDMapping{addr, pubKey})

	//addr = common.HexToAddress(`0xa289f069285341538d09951debe77b49078c1f67`)	// sman tmp
	//pubKey = common.Hex2Bytes(`dee071f32140f62caddbad45a181f7022878a92f87fd08b941423534b3e77a9cccf25bf3369897d698705e0a429011dc2c967a6a605f17535049078985429603`)
	//result = append(result, AddrNodeIDMapping{addr, pubKey})

	return result
}

// 获取主节点数量
func GetCoreNodesCount() int {
	nodes := GetAllSortedCoreNodes()
	return len(nodes)
}

// 获取节点索引 后期可优化下
func GetCoreNodeIndex(address *common.Address) int {
	nodes := GetAllSortedCoreNodes()
	for i := 0; i < len(nodes); i++ {
		if nodes[i].Addr == *address {
			return i
		}
	}
	return -1
}

// 根据pubkey获取节点索引
func GetCoreNodeIndexByPubkey(pubKey []byte) int{
	nodes := GetAllSortedCoreNodes()
	for i := 0; i < len(nodes); i++ {
		if bytes.Compare(nodes[i].Pubkey,  pubKey[1:]) == 0{
			return i
		}
	}
	return -1
}

// 通过出块者地址获取节点公钥
func GetPubkeyByAddress(address *common.Address) []byte{
	nodes := GetAllSortedCoreNodes()
	for i := 0; i < len(nodes); i++ {
		if nodes[i].Addr == *address {
			return nodes[i].Pubkey
		}
	}
	return nil
}

// 根据publick key 获取地址
func GetAddressByPubkey(pubKey []byte) common.Address {
	if len(pubKey) == 65 {
		pubKey = pubKey[1:]
	}
	nodes := GetAllSortedCoreNodes()
	for i := 0; i < len(nodes); i++ {
		if bytes.Compare(nodes[i].Pubkey, pubKey) == 0 {
			return nodes[i].Addr
		}
	}
	return common.Address{}
}