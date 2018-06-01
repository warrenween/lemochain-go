package dpovp

import (
	"crypto/ecdsa"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/crypto"
)

type AddrNodeIDMapping struct{
	Addr common.Address
	Pubkey []byte
}

// 根据publick key 获取地址
func GetAddressByPubkey(pubKey *ecdsa.PublicKey) common.Address {
	addr := crypto.PubkeyToAddress(*pubKey)
	return addr
}

// Get all sorted nodes that who can produce blocks
func GetAllSortedCoreNodes() []AddrNodeIDMapping {
	// TODO
	result := make([]AddrNodeIDMapping, 1)
	addr1 := common.HexToAddress(`0xb2a6935aedca3c64d1a98787f48ef4bc010b09d7`)
	pubkey1:=[]byte(`fa609f0ff5d528f74e013be7a07a4c1365ba5e288ef5ae043037f9f3edaa02e72dbb2d860e01d67a99c289e307250cea4918b8c18377b7fad5d015d90ea90b2a`)
	tmp1 := AddrNodeIDMapping{addr1,pubkey1 }

	//addr2 := common.HexToAddress(`0x2fb44abed468e558ffcb2ef3de03c7746038be04`)
	//pubkey2:= []byte(`03b7e35a9dcb07e2e90a4eddfd711d203e9cdd84d8f1dc2b495edbdf264caf37b19e2e2ee43aa59287f3a8a45e212e24ffc091c71f6ddadbd0bc2baae9c8e00a`)
	//tmp2 := AddrNodeIDMapping{addr2,pubkey2 }

	result[0] = tmp1
	//result[1] = tmp2
	return result
}

// 获取主节点数量
func GetCorNodesCount() int {
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