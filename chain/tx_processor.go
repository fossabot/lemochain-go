package chain

import (
	"errors"
	"github.com/LemoFoundationLtd/lemochain-go/chain/account"
	"github.com/LemoFoundationLtd/lemochain-go/chain/params"
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/chain/vm"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/log"
	"math"
	"math/big"
	"sync"
)

var (
	ErrInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
	ErrInvalidTxInBlock          = errors.New("block contains invalid transaction")
)

type TxProcessor struct {
	chain *BlockChain
	am    *account.Manager
	cfg   *vm.Config // configuration of vm

	lock sync.Mutex
}

type ApplyTxsResult struct {
	Txs     types.Transactions // The transactions executed indeed. These transactions will be packaged in a block
	Events  []*types.Event     // contract events
	Bloom   types.Bloom        // used to search contract events
	GasUsed uint64             // gas used by all transactions
}

func NewTxProcessor(bc *BlockChain) *TxProcessor {
	debug := bc.Flags().Bool(common.Debug)
	cfg := &vm.Config{
		Debug: debug,
	}
	return &TxProcessor{
		chain: bc,
		am:    bc.am,
		cfg:   cfg,
	}
}

// Process processes all transactions in a block. Change accounts' data and execute contract codes.
func (p *TxProcessor) Process(block *types.Block) (*types.Header, error) {
	p.lock.Lock()
	p.lock.Unlock()
	var (
		gp          = new(types.GasPool).AddGas(block.GasLimit())
		gasUsed     = uint64(0)
		totalGasFee = new(big.Int)
		header      = block.Header
		txs         = block.Txs
	)
	p.am.Reset(header.ParentHash)
	// genesis
	if header.Height == 0 {
		log.Warn("It is not necessary to process genesis block.")
		return header, nil
	}
	// Iterate over and process the individual transactions
	for i, tx := range txs {
		gas, err := p.applyTx(gp, header, tx, uint(i), block.Hash())
		if err != nil {
			log.Info("Invalid transaction", "hash", tx.Hash(), "err", err)
			return nil, ErrInvalidTxInBlock
		}
		gasUsed = gasUsed + gas
		fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), tx.GasPrice())
		totalGasFee.Add(totalGasFee, fee)
	}
	p.chargeForGas(totalGasFee, header.MinerAddress)

	return p.FillHeader(header.Copy(), txs, gasUsed)
}

// ApplyTxs picks and processes transactions from miner's tx pool.
func (p *TxProcessor) ApplyTxs(header *types.Header, txs types.Transactions) (*types.Header, types.Transactions, types.Transactions, error) {
	p.lock.Lock()
	p.lock.Unlock()
	gp := new(types.GasPool).AddGas(header.GasLimit)
	gasUsed := uint64(0)
	totalGasFee := new(big.Int)
	selectedTxs := make(types.Transactions, 0)
	invalidTxs := make(types.Transactions, 0)

	p.am.Reset(header.ParentHash)

	for _, tx := range txs {
		// If we don't have enough gas for any further transactions then we're done
		if gp.Gas() < params.TxGas {
			log.Info("Not enough gas for further transactions", "gp", gp)
			break
		}
		// Start executing the transaction
		snap := p.am.Snapshot()

		gas, err := p.applyTx(gp, header, tx, uint(len(selectedTxs)), common.Hash{})
		if err != nil {
			p.am.RevertToSnapshot(snap)
			if err == types.ErrGasLimitReached {
				// block is full
				log.Info("Not enough gas for further transactions", "gp", gp, "lastTxGasLimit", tx.GasLimit())
			} else {
				// Strange error, discard the transaction and get the next in line.
				log.Info("Skipped invalid transaction", "hash", tx.Hash(), "err", err)
				invalidTxs = append(invalidTxs, tx)
			}
			continue
		}
		selectedTxs = append(selectedTxs, tx)

		gasUsed = gasUsed + gas
		fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), tx.GasPrice())
		totalGasFee.Add(totalGasFee, fee)
	}
	p.chargeForGas(totalGasFee, header.MinerAddress)

	newHeader, err := p.FillHeader(header.Copy(), selectedTxs, gasUsed)
	return newHeader, selectedTxs, invalidTxs, err
}

// applyTx processes transaction. Change accounts' data and execute contract codes.
func (p *TxProcessor) applyTx(gp *types.GasPool, header *types.Header, tx *types.Transaction, txIndex uint, blockHash common.Hash) (uint64, error) {
	senderAddr, err := tx.From()
	if err != nil {
		return 0, err
	}
	var (
		// Create a new context to be used in the EVM environment
		context = NewEVMContext(tx, header, txIndex, tx.Hash(), blockHash, p.chain)
		// Create a new environment which holds all relevant information
		// about the transaction and calling mechanisms.
		vmEnv            = vm.NewEVM(context, p.am, *p.cfg)
		sender           = p.am.GetAccount(senderAddr)
		contractCreation = tx.To() == nil
		restGas          = tx.GasLimit()
		mergeFrom        = len(p.am.GetChangeLogs())
	)
	err = p.buyGas(gp, tx)
	if err != nil {
		return 0, err
	}
	restGas, err = p.payIntrinsicGas(tx, restGas)
	if err != nil {
		return 0, err
	}

	// vm errors do not effect consensus and are therefor not assigned to err,
	// except for insufficient balance error.
	var (
		vmErr         error
		recipientAddr common.Address
	)
	if contractCreation {
		_, recipientAddr, restGas, vmErr = vmEnv.Create(sender, tx.Data(), restGas, tx.Amount())
	} else {
		recipientAddr = *tx.To()
		_, restGas, vmErr = vmEnv.Call(sender, recipientAddr, tx.Data(), restGas, tx.Amount())
	}
	if vmErr != nil {
		log.Info("VM returned with error", "err", vmErr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmErr == vm.ErrInsufficientBalance {
			return 0, vmErr
		}
	}
	p.refundGas(gp, tx, restGas)
	p.am.SaveTxInAccount(senderAddr, recipientAddr, tx.Hash())
	// Merge change logs by transaction will save more transaction execution detail than by block
	p.am.MergeChangeLogs(mergeFrom)
	mergeFrom = len(p.am.GetChangeLogs())

	return tx.GasLimit() - restGas, nil
}

func (p *TxProcessor) buyGas(gp *types.GasPool, tx *types.Transaction) error {
	// ignore the error because it is checked in applyTx
	senderAddr, _ := tx.From()
	sender := p.am.GetAccount(senderAddr)

	maxFee := new(big.Int).Mul(new(big.Int).SetUint64(tx.GasLimit()), tx.GasPrice())
	if sender.GetBalance().Cmp(maxFee) < 0 {
		return ErrInsufficientBalanceForGas
	}
	if err := gp.SubGas(tx.GasLimit()); err != nil {
		return err
	}
	sender.SetBalance(new(big.Int).Sub(sender.GetBalance(), maxFee))
	return nil
}

func (p *TxProcessor) payIntrinsicGas(tx *types.Transaction, restGas uint64) (uint64, error) {
	gas, err := IntrinsicGas(tx.Data(), tx.To() == nil)
	if err != nil {
		return restGas, err
	}
	if restGas < gas {
		return restGas, vm.ErrOutOfGas
	}
	return restGas - gas, nil
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, vm.ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, vm.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}

func (p *TxProcessor) refundGas(gp *types.GasPool, tx *types.Transaction, restGas uint64) {
	// ignore the error because it is checked in applyTx
	senderAddr, _ := tx.From()
	sender := p.am.GetAccount(senderAddr)

	// Return LEMO for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(restGas), tx.GasPrice())
	sender.SetBalance(new(big.Int).Add(sender.GetBalance(), remaining))

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	gp.AddGas(restGas)
}

// chargeForGas change the gas to miner
func (p *TxProcessor) chargeForGas(charge *big.Int, minerAddress common.Address) {
	if charge.Cmp(new(big.Int)) != 0 {
		miner := p.am.GetAccount(minerAddress)
		miner.SetBalance(new(big.Int).Add(miner.GetBalance(), charge))
	}
}

// FillHeader creates a new header then fills it with the result of transactions process
func (p *TxProcessor) FillHeader(header *types.Header, txs types.Transactions, gasUsed uint64) (*types.Header, error) {
	if len(txs) > 0 {
		log.Infof("process %d transactions", len(txs))
	}
	events := p.am.GetEvents()
	header.Bloom = types.CreateBloom(events)
	header.EventRoot = types.DeriveEventsSha(events)
	header.GasUsed = gasUsed
	header.TxRoot = types.DeriveTxsSha(txs)
	// Pay miners at the end of their tenure. This method increases miners' balance.
	p.chain.engine.Finalize(header, p.am)
	// Update version trie, storage trie.
	err := p.am.Finalise()
	if err != nil {
		// Access trie node fail.
		return nil, err
	}
	header.VersionRoot = p.am.GetVersionRoot()
	changeLogs := p.am.GetChangeLogs()
	header.LogRoot = types.DeriveChangeLogsSha(changeLogs)
	return header, nil
}
