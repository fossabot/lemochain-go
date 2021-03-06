package types

import (
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var (
	logType1 = ChangeLogType(1)
	logType2 = ChangeLogType(2)
)

func getAccountData() *AccountData {
	return &AccountData{
		Address:       common.HexToAddress("0x10000"),
		Balance:       big.NewInt(100),
		CodeHash:      common.HexToHash("0x1d5f11eaa13e02cdca886181dc38ab4cb8cf9092e86c000fb42d12c8b504500e"),
		StorageRoot:   common.HexToHash("0xcbeb7c7e36b846713bc99b8fa527e8d552e31bfaa1ac0f2b773958cda3aba3ed"),
		NewestRecords: map[ChangeLogType]VersionRecord{logType1: {100, 10}, logType2: {101, 11}},
		TxHashList:    []common.Hash{common.HexToHash("0x11"), common.HexToHash("0x22")},
	}
}

func TestAccountData_EncodeRLP_DecodeRLP(t *testing.T) {
	account := getAccountData()

	data, err := rlp.EncodeToBytes(account)
	assert.NoError(t, err)

	// decode correct data
	decoded := new(AccountData)
	err = rlp.DecodeBytes(data, decoded)
	assert.NoError(t, err)
	assert.Equal(t, account, decoded)
	assert.Equal(t, uint32(100), decoded.NewestRecords[logType1].Version)
	assert.Equal(t, uint32(10), decoded.NewestRecords[logType1].Height)
	assert.Equal(t, 2, len(decoded.TxHashList))
	assert.Equal(t, common.HexToHash("0x22"), decoded.TxHashList[1])

	// decode incorrect data
	decoded = new(AccountData)
	err = rlp.DecodeBytes(common.Hex2Bytes("f8a594000000000000000000000000000000000001000064a01d5f11eaa13e02cdca886181dc38ab4cb8cf9092e86c000fb42d12c8b504500ea0cbeb7c7e36b846713bc99b8fa527e8d552e31bfaa1ac0f2b773958cda3aba3edf842a00000000000000000000000000000000000000000000000000000000000000011b00000000000000000000000000000000000000000000000000000000000000022c8c301640ac302650b"), decoded)
	assert.Error(t, err)
}

func TestAccountData_Copy(t *testing.T) {
	account := getAccountData()

	cpy := account.Copy()
	assert.Equal(t, account, cpy)
	account.Balance.SetInt64(101)
	assert.NotEqual(t, account.Balance, cpy.Balance)
	account.NewestRecords[logType1] = VersionRecord{Version: 101, Height: 11}
	assert.NotEqual(t, account.NewestRecords[logType1].Version, cpy.NewestRecords[logType1].Version)
	account.TxHashList[0] = common.HexToHash("0x10")
	assert.NotEqual(t, account.TxHashList[0], cpy.TxHashList[0])
}

func TestAccountData_MarshalJSON_UnmarshalJSON(t *testing.T) {
	account := getAccountData()

	data, err := account.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"address":"Lemo8888888888888888888888888888883CPHBJ","balance":"100","codeHash":"0x1d5f11eaa13e02cdca886181dc38ab4cb8cf9092e86c000fb42d12c8b504500e","root":"0xcbeb7c7e36b846713bc99b8fa527e8d552e31bfaa1ac0f2b773958cda3aba3ed","records":{"1":{"version":"100","height":"10"},"2":{"version":"101","height":"11"}}}`, string(data))

	decode := new(AccountData)
	err = decode.UnmarshalJSON(data)
	assert.NoError(t, err)
	account.TxHashList = nil
	assert.Equal(t, account, decode)
}
