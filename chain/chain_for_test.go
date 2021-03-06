package chain

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-go/chain/account"
	"github.com/LemoFoundationLtd/lemochain-go/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-go/chain/params"
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/common/flag"
	"github.com/LemoFoundationLtd/lemochain-go/common/log"
	"github.com/LemoFoundationLtd/lemochain-go/store"
	"github.com/LemoFoundationLtd/lemochain-go/store/protocol"
	"math/big"
	"time"
)

type blockInfo struct {
	hash        common.Hash
	parentHash  common.Hash
	height      uint32
	author      common.Address
	versionRoot common.Hash
	txRoot      common.Hash
	logRoot     common.Hash
	txList      []*types.Transaction
	gasLimit    uint64
	time        uint32
	deputyRoot  []byte
	deputyNodes deputynode.DeputyNodes
}

var (
	chainID         uint16 = 200
	bigNumber, _           = new(big.Int).SetString("1000000000000000000000", 10) // 1 thousand
	testSigner             = types.DefaultSigner{}
	testPrivate, _         = crypto.HexToECDSA("432a86ab8765d82415a803e29864dcfc1ed93dac949abf6f95a583179f27e4bb") // secp256k1.V = 1
	testAddr               = crypto.PubkeyToAddress(testPrivate.PublicKey)                                         // 0x0107134B9CdD7D89F83eFa6175F9b3552F29094c
	defaultAccounts        = []common.Address{
		common.HexToAddress("0x10000"), common.HexToAddress("0x20000"), testAddr,
	}
	defaultBlocks     = make([]*types.Block, 0)
	defaultBlockInfos = []blockInfo{
		// genesis block must no transactions
		{
			hash:        common.HexToHash("0xfceb0a339a5be9689a3db2b7df2a05feba5b1f6e1ce385f14764d139556b9426"),
			height:      0,
			author:      defaultAccounts[0],
			versionRoot: common.HexToHash("0x6eea9449a171035539c71d2895830afc061d0777da6e86735d9899c888d953c1"),
			txRoot:      common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"), // empty merkle
			logRoot:     common.HexToHash("0xb0d3749ecc3a7a0db6368284320863a3d2fa963b2c33b41c6ebf8632cd84bda9"),
			time:        1538209751,
		},
		// block 1 is stable block
		{
			hash:        common.HexToHash("0xfabb0fff8f44b829f937070de3d8625972900867a6dce77b24c0b5ccd1d591ca"),
			height:      1,
			author:      common.HexToAddress("0x20000"),
			versionRoot: common.HexToHash("0xfa470f96f703f04ba85d39bf7f64aa43e8b352e541fe36ea9f053336dfd1b22a"),
			txRoot:      common.HexToHash("0x44bb706ae39d4e24df62fee83d9daa0e0a8e422e9b3f6d48243e2511cb055479"),
			logRoot:     common.HexToHash("0xb6062a94cf5ce70ec3910072da83eca56ba584aba8273e4d0837eb03c090e473"),
			txList: []*types.Transaction{
				// testAddr -> defaultAccounts[0] 1
				signTransaction(types.NewTransaction(defaultAccounts[0], common.Big1, 2000000, common.Big2, []byte{12}, chainID, 1538210391, "aa", "aaa"), testPrivate),
				// testAddr -> defaultAccounts[1] 1
				makeTransaction(testPrivate, defaultAccounts[1], common.Big1, common.Big2, 1538210491, 2000000),
			},
			gasLimit: 20000000,
			time:     1538209755,
		},
		// block 2 is not stable block
		{
			hash:        common.HexToHash("0x18229ab6caf0a9f2972eaa97ff4ede9189e0f6d96bbef3a972795313d92319b2"),
			height:      2,
			author:      defaultAccounts[0],
			versionRoot: common.HexToHash("0x407e1d64578f852d97ab3b5a655ece80ca1422ef68ba0145756f35a4b7fb141b"),
			txRoot:      common.HexToHash("0xf499c76d7f3693700d1cab1618d7c62c619686578ce982375161747927117a8f"),
			logRoot:     common.HexToHash("0x40ab1668ecddd288ded6cec4135dadb071beee4e2826640e96c416894aaa6469"),
			txList: []*types.Transaction{
				// testAddr -> defaultAccounts[0] 2
				makeTransaction(testPrivate, defaultAccounts[0], bigNumber, common.Big2, 1538210395, 2000000),
			},
			time:     1538209758,
			gasLimit: 20000000,
		},
		// block 3 is not store in db
		{
			hash:        common.HexToHash("0xfd283c1db324237dcf224ccae7e68bca17f1f806d68a0208458188cca5f20d23"),
			height:      3,
			author:      defaultAccounts[0],
			versionRoot: common.HexToHash("0x26a8540cd0e07e18d08ad484cfb44b0fd40dc5f197678897fc80076dbfa942b4"),
			txRoot:      common.HexToHash("0x3f4519bff066109e8b47621fc8fa8bc791abec347478cc8f0d084012fc83f76d"),
			logRoot:     common.HexToHash("0xa92582c4c7627d130a38c94e568b8c671817fe742930636c7b0190b7f7344064"),
			txList: []*types.Transaction{
				// testAddr -> defaultAccounts[0] 2
				makeTransaction(testPrivate, defaultAccounts[0], common.Big2, common.Big2, 1538210398, 30000),
				// testAddr -> defaultAccounts[1] 2
				makeTransaction(testPrivate, defaultAccounts[1], common.Big2, common.Big3, 1538210425, 30000),
			},
			time:     1538209761,
			gasLimit: 20000000,
		},
	}
)

func init() {
	store.ClearData()
	log.Setup(log.LevelInfo, false, false)
}

// newChain creates chain for test
func newChain() *BlockChain {
	db := newDB()
	bc, err := NewBlockChain(chainID, NewDpovp(10*1000, db), db, flag.CmdFlags{})
	if err != nil {
		panic(err)
	}
	return bc
}

// newDB creates db for test chain module
func newDB() protocol.ChainDB {
	db, err := store.NewCacheChain(store.GetStorePath())
	if err != nil {
		panic(err)
	}

	for i, _ := range defaultBlockInfos {
		if i > 0 {
			defaultBlockInfos[i].parentHash = defaultBlocks[i-1].Hash()
		}
		newBlock := makeBlock(db, defaultBlockInfos[i], i < 3)
		defaultBlocks = append(defaultBlocks, newBlock)
	}
	// err = db.SetStableBlock(defaultBlockInfos[1].hash)
	err = db.SetStableBlock(defaultBlocks[1].Hash())
	if err != nil {
		panic(err)
	}
	return db
}

func makeBlock(db protocol.ChainDB, info blockInfo, save bool) *types.Block {
	manager := account.NewManager(info.parentHash, db)
	// sign transactions
	var err error
	var gasUsed uint64 = 0
	txRoot := types.DeriveTxsSha(info.txList)
	if txRoot != info.txRoot {
		if info.txRoot != (common.Hash{}) {
			fmt.Printf("%d txRoot hash error. except: %s, got: %s\n", info.height, info.txRoot.Hex(), txRoot.Hex())
		}
		info.txRoot = txRoot
	}

	var deputyRoot []byte
	if len(info.deputyNodes) > 0 {
		deputyRoot = types.DeriveDeputyRootSha(info.deputyNodes).Bytes()
	}
	if bytes.Compare(deputyRoot, info.deputyRoot) != 0 {
		if len(info.deputyNodes) > 0 || len(info.deputyRoot) != 0 {
			fmt.Printf("%d deputyRoot error. except: %s, got: %s\n", info.height, common.ToHex(info.deputyRoot), common.ToHex(deputyRoot))
		}
		info.deputyRoot = deputyRoot
	}

	// genesis coin
	if info.height == 0 {
		owner := manager.GetAccount(testAddr)
		// 1 million
		balance, _ := new(big.Int).SetString("1000000000000000000000000", 10)
		owner.SetBalance(balance)
	}
	// account
	salary := new(big.Int)
	for _, tx := range info.txList {
		gas := params.TxGas + params.TxDataNonZeroGas*uint64(len(tx.Data()))
		fromAddr, err := tx.From()
		if err != nil {
			panic(err)
		}
		from := manager.GetAccount(fromAddr)
		fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), tx.GasPrice())
		cost := new(big.Int).Add(tx.Amount(), fee)
		to := manager.GetAccount(*tx.To())
		// make sure the change log has right order
		if fromAddr.Hex() < tx.To().Hex() {
			from.SetBalance(new(big.Int).Sub(from.GetBalance(), cost))
			to.SetBalance(new(big.Int).Add(to.GetBalance(), tx.Amount()))
		} else {
			to.SetBalance(new(big.Int).Add(to.GetBalance(), tx.Amount()))
			from.SetBalance(new(big.Int).Sub(from.GetBalance(), cost))
		}
		gasUsed += gas
		salary.Add(salary, fee)
		from.(*account.SafeAccount).AppendTx(tx.Hash())
		to.(*account.SafeAccount).AppendTx(tx.Hash())
	}
	if salary.Cmp(new(big.Int)) != 0 {
		miner := manager.GetAccount(info.author)
		miner.SetBalance(new(big.Int).Add(miner.GetBalance(), salary))
	}
	err = manager.Finalise()
	if err != nil {
		panic(err)
	}
	// header
	if manager.GetVersionRoot() != info.versionRoot {
		if info.versionRoot != (common.Hash{}) {
			fmt.Printf("%d version root error. except: %s, got: %s\n", info.height, info.versionRoot.Hex(), manager.GetVersionRoot().Hex())
		}
		info.versionRoot = manager.GetVersionRoot()
	}
	changeLogs := manager.GetChangeLogs()
	// fmt.Printf("%d changeLogs %v\n", info.height, changeLogs)
	logRoot := types.DeriveChangeLogsSha(changeLogs)
	if logRoot != info.logRoot {
		if info.logRoot != (common.Hash{}) {
			fmt.Printf("%d change logs root error. except: %s, got: %s\n", info.height, info.logRoot.Hex(), logRoot.Hex())
		}
		info.logRoot = logRoot
	}
	if info.time == 0 {
		info.time = uint32(time.Now().Unix())
	}
	if info.gasLimit == 0 {
		info.gasLimit = 1000000
	}
	header := &types.Header{
		ParentHash:   info.parentHash,
		MinerAddress: info.author,
		VersionRoot:  info.versionRoot,
		TxRoot:       info.txRoot,
		LogRoot:      info.logRoot,
		Bloom:        types.CreateBloom(nil),
		EventRoot:    common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"), // empty merkle
		Height:       info.height,
		GasLimit:     info.gasLimit,
		GasUsed:      gasUsed,
		Time:         info.time,
		Extra:        []byte{},
	}
	if len(info.deputyRoot) > 0 {
		header.DeputyRoot = make([]byte, len(info.deputyRoot))
		copy(header.DeputyRoot[:], info.deputyRoot[:])
	}

	blockHash := header.Hash()
	if blockHash != info.hash {
		if info.hash != (common.Hash{}) {
			fmt.Printf("%d block hash error. except: %s, got: %s\n", info.height, info.hash.Hex(), blockHash.Hex())
		}
		info.hash = blockHash
	}
	// block
	block := &types.Block{
		Txs:         info.txList,
		ChangeLogs:  changeLogs,
		DeputyNodes: info.deputyNodes,
	}
	block.SetHeader(header)
	if save {
		err = db.SetBlock(blockHash, block)
		if err != nil && err != store.ErrExist {
			panic(err)
		}
		err = manager.Save(blockHash)
		if err != nil {
			panic(err)
		}
	}
	return block
}

func makeTx(fromPrivate *ecdsa.PrivateKey, to common.Address, amount *big.Int) *types.Transaction {
	return makeTransaction(fromPrivate, to, amount, common.Big1, uint64(time.Now().Unix()+300), 1000000)
}

func makeTransaction(fromPrivate *ecdsa.PrivateKey, to common.Address, amount, gasPrice *big.Int, expiration uint64, gasLimit uint64) *types.Transaction {
	tx := types.NewTransaction(to, amount, gasLimit, gasPrice, []byte{}, chainID, expiration, "", "")
	return signTransaction(tx, fromPrivate)
}

func signTransaction(tx *types.Transaction, private *ecdsa.PrivateKey) *types.Transaction {
	tx, err := types.SignTx(tx, testSigner, private)
	if err != nil {
		panic(err)
	}
	return tx
}

// h returns hash for test
func h(i int64) common.Hash { return common.HexToHash(fmt.Sprintf("0xa%x", i)) }

// b returns block hash for test
func b(i int64) common.Hash { return common.HexToHash(fmt.Sprintf("0xb%x", i)) }

// c returns code hash for test
func c(i int64) common.Hash { return common.HexToHash(fmt.Sprintf("0xc%x", i)) }

// k returns storage key hash for test
func k(i int64) common.Hash { return common.HexToHash(fmt.Sprintf("0xd%x", i)) }

// t returns transaction hash for test
func th(i int64) common.Hash { return common.HexToHash(fmt.Sprintf("0xe%x", i)) }
