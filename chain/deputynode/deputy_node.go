package deputynode

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/common/crypto/sha3"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-go/common/log"
	"github.com/LemoFoundationLtd/lemochain-go/common/rlp"
	"net"
	"sync"
)

const (
	SnapshotBlockInterval = 100000
	TransitionPeriod      = 1000
)

var (
	selfNodeKey *ecdsa.PrivateKey
)

func GetSelfNodeKey() *ecdsa.PrivateKey {
	return selfNodeKey
}

func GetSelfNodeID() []byte {
	return (crypto.FromECDSAPub(&selfNodeKey.PublicKey))[1:]
}

func SetSelfNodeKey(key *ecdsa.PrivateKey) {
	selfNodeKey = key
}

//go:generate gencodec -type DeputyNode --field-override deputyNodeMarshaling -out gen_deputy_node_json.go

// DeputyNode
type DeputyNode struct {
	MinerAddress common.Address `json:"minerAddress"   gencodec:"required"`
	NodeID       []byte         `json:"nodeID"         gencodec:"required"`
	IP           net.IP         `json:"ip"             gencodec:"required"` // ip
	Port         uint32         `json:"port"           gencodec:"required"` // 端口
	Rank         uint32         `json:"rank"           gencodec:"required"` // 排名 从0开始
	Votes        uint32         `json:"votes"          gencodec:"required"` // 得票数
}

func (d *DeputyNode) Hash() (h common.Hash) {
	data := []interface{}{
		d.MinerAddress,
		d.NodeID,
		d.IP,
		d.Port,
		d.Rank,
		d.Votes,
	}
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, data)
	hw.Sum(h[:0])
	return h
}

func (d *DeputyNode) Check() error {
	if len(d.NodeID) != 64 {
		log.Errorf("incorrect field: 'NodeID'.value: %s", common.ToHex(d.NodeID))
		return errors.New("incorrect field: 'NodeID'")
	}
	if d.Port > 65535 {
		log.Errorf("incorrect field: 'port'. value: %d", d.Port)
		return errors.New("max deputy node's port is 65535")
	}
	return nil
}

type DeputyNodes []*DeputyNode

type deputyNodeMarshaling struct {
	NodeID hexutil.Bytes
	IP     hexutil.IP
	Port   hexutil.Uint32
	Rank   hexutil.Uint32
	Votes  hexutil.Uint32
}

type DeputyNodesRecord struct {
	height uint32
	nodes  DeputyNodes
}

// Manager 代理节点管理器
type Manager struct {
	DeputyNodesList []*DeputyNodesRecord // key：节点列表生效开始高度 value：节点列表
	lock            sync.Mutex
}

// Add 投票结束 统计结果通过add函数缓存起来
func (d *Manager) Add(height uint32, nodes DeputyNodes) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.DeputyNodesList = append(d.DeputyNodesList, &DeputyNodesRecord{height: height, nodes: nodes})
}

var deputyNodeManger *Manager
var once sync.Once

func Instance() *Manager {
	once.Do(func() {
		deputyNodeManger = &Manager{
			DeputyNodesList: make([]*DeputyNodesRecord, 0, 1),
		}
	})
	return deputyNodeManger
}

// getDeputiesByHeight 通过height获取对应的节点列表
func (d *Manager) getDeputiesByHeight(height uint32) DeputyNodes {
	d.lock.Lock()
	defer d.lock.Unlock()
	var nodes DeputyNodes
	for i := 0; i < len(d.DeputyNodesList)-1; i++ {
		if d.DeputyNodesList[i].height <= height && d.DeputyNodesList[i+1].height > height {
			nodes = d.DeputyNodesList[i].nodes
			break
		}
	}
	if nodes == nil {
		nodes = d.DeputyNodesList[len(d.DeputyNodesList)-1].nodes
	}
	return nodes
}

// getDeputyNodeCount 获取共识节点数量
func (d *Manager) GetDeputiesCount() int {
	return len(d.DeputyNodesList[0].nodes)
}

// getNodeByAddress 获取address对应的节点
func (d *Manager) GetDeputyByAddress(height uint32, addr common.Address) *DeputyNode {
	nodes := d.getDeputiesByHeight(height)
	for _, node := range nodes {
		if node.MinerAddress == addr {
			return node
		}
	}
	return nil
}

// getNodeByNodeID 根据nodeid获取对应的节点
func (d *Manager) GetDeputyByNodeID(height uint32, nodeID []byte) *DeputyNode {
	nodes := d.getDeputiesByHeight(height)
	for _, node := range nodes {
		if bytes.Compare(node.NodeID, nodeID) == 0 {
			return node
		}
	}
	return nil
}

// 获取最新块的出块者序号与本节点序号差
func (d *Manager) GetSlot(height uint32, firstAddress, nextAddress common.Address) int {
	firstNode := d.GetDeputyByAddress(height, firstAddress)
	nextNode := d.GetDeputyByAddress(height, nextAddress)
	if ((height == 1) || (height > SnapshotBlockInterval && height%SnapshotBlockInterval == TransitionPeriod+1)) && nextNode != nil {
		return int(nextNode.Rank + 1)
	}
	if firstNode == nil || nextNode == nil {
		return -1
	}
	nodeCount := d.GetDeputiesCount()
	// 只有一个主节点
	if nodeCount == 1 {
		log.Debug("getSlot: only one star node")
		return 1
	}
	return (int(nextNode.Rank) - int(firstNode.Rank) + nodeCount) % nodeCount
}

// TimeToHandOutRewards 是否该发出块奖励了
func (d *Manager) TimeToHandOutRewards(height uint32) bool {
	// d.lock.Lock()
	// defer d.lock.Unlock()
	for i := 1; i < len(d.DeputyNodesList); i++ {
		if d.DeputyNodesList[i].height+1000+1 == height {
			return true
		}
	}
	return false
}

// IsSelfDeputyNode
func (d *Manager) IsSelfDeputyNode(height uint32) bool {
	node := d.GetDeputyByNodeID(height, GetSelfNodeID())
	return node != nil
}

func (d *Manager) Clear() {
	d.DeputyNodesList = make([]*DeputyNodesRecord, 0, 1)
}
