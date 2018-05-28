package dpovp

import (
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/log"
	"crypto/ecdsa"
)

// 根据publick key 获取地址
func GetAddressByPubkey(pubKey *ecdsa.PublicKey) common.Address {
	//switch len(keyBytes[:]) {
	//case 64:
	//	// add 'uncompressed key' flag
	//	keyBytes = append([]byte{0x04}, keyBytes...)
	//default:
	//}
	//pubKey := crypto.ToECDSAPub(keyBytes)
	addr := crypto.PubkeyToAddress(*pubKey)
	str := common.ToHex(addr[:])
	log.Info(`sman address:%s`, str)
	return addr
}

// Get all sorted nodes that who can produce blocks
func GetAllSortedCoreNodes() []common.Address {
	// TODO
	result := make([]common.Address, 2)
	str1 := `0x076dd80d5ac6324ded3c74668074e46ba3b73468`
	addr1 := common.HexToAddress(str1)
	str2 := `0x2fb44abed468e558ffcb2ef3de03c7746038be04`
	addr2 := common.HexToAddress(str2)
	//result = append(result, addr1)
	result[1] = addr1
	result[0] = addr2
	return result
}

// 获取主节点数量
func GetCorNodesCount() int {
	nodes := GetAllSortedCoreNodes()
	return len(nodes)
}

// 获取节点索引 后期可优化下
func GetCoreNodeIndex(address common.Address) int {
	nodes := GetAllSortedCoreNodes()
	for i := 0; i < len(nodes); i++ {
		if nodes[i] == address {
			return i
		}
	}
	return -1
}
