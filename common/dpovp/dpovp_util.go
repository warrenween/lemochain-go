package dpovp

import (
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/crypto"
)

// 根据publick key 获取地址
func GetAddressByPubkey(keyBytes []byte) common.Address {
	pubKey := crypto.ToECDSAPub(keyBytes)
	addr := crypto.PubkeyToAddress(*pubKey)
	return addr
}

// Get all sorted nodes that who can produce blocks
func GetAllSortedCoreNodes() []common.Address {
	// TODO
	result := make([]common.Address, 1)
	str1 := `0x076dd80d5ac6324ded3c74668074e46ba3b73468`
	addr1 := common.HexToAddress(str1)
	//result = append(result, addr1)
	result[0] = addr1
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
