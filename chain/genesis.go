package chain

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-go/chain/account"
	"github.com/LemoFoundationLtd/lemochain-go/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-go/common/log"
	"github.com/LemoFoundationLtd/lemochain-go/store/protocol"
	"math/big"
	"net"
	"time"
)

// DefaultDeputyNodes
var DefaultDeputyNodes = deputynode.DeputyNodes{
	&deputynode.DeputyNode{
		MinerAddress: decodeMinerAddress("Lemo83GN72GYH2NZ8BA729Z9TCT7KQ5FC3CR6DJG"),
		NodeID:       common.FromHex("0x5e3600755f9b512a65603b38e30885c98cbac70259c3235c9b3f42ee563b480edea351ba0ff5748a638fe0aeff5d845bf37a3b437831871b48fd32f33cd9a3c0"),
		IP:           net.ParseIP("127.0.0.1"),
		Port:         7001,
		Rank:         0,
		Votes:        50000,
	},
	&deputynode.DeputyNode{
		MinerAddress: decodeMinerAddress("Lemo83JW7TBPA7P2P6AR9ZC2WCQJYRNHZ4NJD4CY"),
		NodeID:       common.FromHex("0xddb5fc36c415799e4c0cf7046ddde04aad6de8395d777db4f46ebdf258e55ee1d698fdd6f81a950f00b78bb0ea562e4f7de38cb0adf475c5026bb885ce74afb0"),
		IP:           net.ParseIP("127.0.0.1"),
		Port:         7002,
		Rank:         1,
		Votes:        40000,
	},
	&deputynode.DeputyNode{
		MinerAddress: decodeMinerAddress("Lemo842BJZ4DKCC764C63Y6A943775JH6NQ3Z33Y"),
		NodeID:       common.FromHex("0x7739f34055d3c0808683dbd77a937f8e28f707d5b1e873bbe61f6f2d0347692f36ef736f342fb5ce4710f7e337f062cc2110d134b63a9575f78cb167bfae2f43"),
		IP:           net.ParseIP("127.0.0.1"),
		Port:         7003,
		Rank:         2,
		Votes:        30000,
	},
	&deputynode.DeputyNode{
		MinerAddress: decodeMinerAddress("Lemo837QGPS3YNTYNF53CD88WA5DR3ABNA95W2DG"),
		NodeID:       common.FromHex("0x34f0df789b46e9bc09f23d5315b951bc77bbfeda653ae6f5aab564c9b4619322fddb3b1f28d1c434250e9d4dd8f51aa8334573d7281e4d63baba913e9fa6908f"),
		IP:           net.ParseIP("127.0.0.1"),
		Port:         7004,
		Rank:         3,
		Votes:        20000,
	},
	&deputynode.DeputyNode{
		MinerAddress: decodeMinerAddress("Lemo83HKZK68JQZDRGS5PWT2ZBSKR5CRADCSJB9B"),
		NodeID:       common.FromHex("0x5b980ffb1b463fce4773a22ebf376c07c6207023b016b36ccfaba7be1cd1ab4a91737741cd43b7fcb10879e0fcf314d69fa953daec0f02be0f8f9cedb0cb3797"),
		IP:           net.ParseIP("127.0.0.1"),
		Port:         7005,
		Rank:         4,
		Votes:        10000,
	},
}

func decodeMinerAddress(input string) common.Address {
	if address, err := common.StringToAddress(input); err == nil {
		return address
	}
	panic(fmt.Sprintf("deputy nodes have invalid miner address: %s", input))
}

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis_json.go

type Genesis struct {
	Time        uint32                 `json:"timestamp"     gencodec:"required"`
	ExtraData   []byte                 `json:"extraData"`
	GasLimit    uint64                 `json:"gasLimit"      gencodec:"required"`
	Founder     common.Address         `json:"founder"       gencodec:"required"`
	DeputyNodes deputynode.DeputyNodes `json:"deputyNodes"   gencodec:"required"`
}

type genesisSpecMarshaling struct {
	Time        hexutil.Uint32
	ExtraData   hexutil.Bytes
	GasLimit    hexutil.Uint64
	DeputyNodes []*deputynode.DeputyNode
}

// DefaultGenesisBlock default genesis block
func DefaultGenesisBlock() *Genesis {
	timeSpan, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-08-30 12:00:00", time.UTC)
	address := decodeMinerAddress("Lemo83GN72GYH2NZ8BA729Z9TCT7KQ5FC3CR6DJG")
	return &Genesis{
		Time:        uint32(timeSpan.Unix()),
		ExtraData:   []byte(""),
		GasLimit:    105000000,
		Founder:     address,
		DeputyNodes: DefaultDeputyNodes,
	}
}

// SetupGenesisBlock setup genesis block
func SetupGenesisBlock(db protocol.ChainDB, genesis *Genesis) (common.Hash, error) {
	if genesis == nil {
		log.Info("Writing default genesis block.")
		genesis = DefaultGenesisBlock()
	}
	if len(genesis.DeputyNodes) == 0 {
		panic("default deputy nodes can't be empty")
	}

	if len(genesis.ExtraData) > 256 {
		panic("genesis block's extraData length larger than 256")
	}

	// check genesis block's time
	if int64(genesis.Time) > time.Now().Unix() {
		panic("Genesis block's time can't be larger than current time.")
	}
	// check deputy nodes
	for _, deputy := range genesis.DeputyNodes {
		if err := deputy.Check(); err != nil {
			panic("genesis deputy nodes check error")
		}
	}

	am := account.NewManager(common.Hash{}, db)
	block := genesis.ToBlock()
	genesis.setBalance(am)
	if err := am.Finalise(); err != nil {
		return common.Hash{}, fmt.Errorf("setup genesis block failed: %v", err)
	}
	block.Header.VersionRoot = am.GetVersionRoot()
	logs := am.GetChangeLogs()
	block.SetChangeLogs(logs)
	block.Header.LogRoot = types.DeriveChangeLogsSha(logs)
	hash := block.Hash()
	if err := db.SetBlock(hash, block); err != nil {
		return common.Hash{}, fmt.Errorf("setup genesis block failed: %v", err)
	}
	if err := am.Save(hash); err != nil {
		return common.Hash{}, fmt.Errorf("setup genesis block failed: %v", err)
	}
	if err := db.SetStableBlock(hash); err != nil {
		return common.Hash{}, fmt.Errorf("setup genesis block failed: %v", err)
	}
	return block.Hash(), nil
}

// ToBlock
func (g *Genesis) ToBlock() *types.Block {
	head := &types.Header{
		ParentHash:   common.Hash{},
		MinerAddress: g.Founder,
		TxRoot:       common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"), // empty merkle
		EventRoot:    common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"), // empty merkle
		Height:       0,
		GasLimit:     g.GasLimit,
		Extra:        g.ExtraData,
		Time:         g.Time,
		DeputyRoot:   types.DeriveDeputyRootSha(g.DeputyNodes).Bytes(),
	}
	block := types.NewBlock(head, nil, nil, nil, nil)
	block.SetDeputyNodes(g.DeputyNodes)
	return block
}

func (g *Genesis) setBalance(am *account.Manager) {
	total, _ := new(big.Int).SetString("1600000000000000000000000000", 10) // 1.6 billion
	am.GetAccount(g.Founder).SetBalance(total)
}
