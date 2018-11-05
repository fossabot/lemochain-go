package types

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-go/common/crypto/sha3"
	"io"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-go/common/rlp"
)

//go:generate gencodec -type Event -field-override eventMarshaling -out gen_event_json.go

var (
	TopicContractCreation = rlpHash("Contract creation")
	TopicRunFail          = rlpHash("Contract run fail")
)

// Event represents a contract event event. These Events are generated by the LOG opcode and
// stored/indexed by the node.
type Event struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address common.Address `json:"address" gencodec:"required"`
	// list of topics provided by the contract.
	Topics []common.Hash `json:"topics" gencodec:"required"`
	// supplied by the contract, usually ABI-encoded
	Data []byte `json:"data" gencodec:"required"`

	// Derived fields. These fields are filled in by the node
	// but not secured by consensus.
	// block in which the transaction was included
	BlockHeight uint32 `json:"blockHeight"`
	// hash of the transaction
	TxHash common.Hash `json:"transactionHash" gencodec:"required"`
	// index of the transaction in the block
	TxIndex uint `json:"transactionIndex" gencodec:"required"`
	// hash of the block in which the transaction was included
	BlockHash common.Hash `json:"blockHash"`
	// index of the event in the receipt
	Index uint `json:"eventIndex" gencodec:"required"`

	// The Removed field is true if this event was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive Events through a filter query.
	Removed bool `json:"removed"`
}

type eventMarshaling struct {
	Data        hexutil.Bytes
	BlockHeight hexutil.Uint64Hex
	TxIndex     hexutil.Uint64Hex
	Index       hexutil.Uint64Hex
}

type rlpEvent struct {
	Address common.Address
	Topics  []common.Hash
	Data    []byte
}

type rlpStorageEvent struct {
	Address     common.Address
	Topics      []common.Hash
	Data        []byte
	BlockHeight uint32
	TxHash      common.Hash
	TxIndex     uint
	BlockHash   common.Hash
	Index       uint
}

// Hash returns the keccak256 hash of its RLP encoding.
func (l *Event) Hash() (h common.Hash) {
	hw := sha3.NewKeccak256()
	// this will call EncodeRLP
	rlp.Encode(hw, l)
	hw.Sum(h[:0])
	return h
}

// EncodeRLP implements rlp.Encoder.
func (l *Event) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, rlpEvent{Address: l.Address, Topics: l.Topics, Data: l.Data})
}

// DecodeRLP implements rlp.Decoder.
func (l *Event) DecodeRLP(s *rlp.Stream) error {
	var dec rlpEvent
	err := s.Decode(&dec)
	if err == nil {
		l.Address, l.Topics, l.Data = dec.Address, dec.Topics, dec.Data
	}
	return err
}

func (l *Event) String() string {
	return fmt.Sprintf(`event: %x %x %x %x %d %x %d`, l.Address, l.Topics, l.Data, l.TxHash, l.TxIndex, l.BlockHash, l.Index)
}

// EventForStorage is a wrapper around a Event that flattens and parses the entire content of
// a event including non-consensus fields.
type EventForStorage Event

// EncodeRLP implements rlp.Encoder.
func (l *EventForStorage) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, rlpStorageEvent{
		Address:     l.Address,
		Topics:      l.Topics,
		Data:        l.Data,
		BlockHeight: l.BlockHeight,
		TxHash:      l.TxHash,
		TxIndex:     l.TxIndex,
		BlockHash:   l.BlockHash,
		Index:       l.Index,
	})
}

// DecodeRLP implements rlp.Decoder.
func (l *EventForStorage) DecodeRLP(s *rlp.Stream) error {
	var dec rlpStorageEvent
	err := s.Decode(&dec)
	if err == nil {
		*l = EventForStorage{
			Address:     dec.Address,
			Topics:      dec.Topics,
			Data:        dec.Data,
			BlockHeight: dec.BlockHeight,
			TxHash:      dec.TxHash,
			TxIndex:     dec.TxIndex,
			BlockHash:   dec.BlockHash,
			Index:       dec.Index,
		}
	}
	return err
}
