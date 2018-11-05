// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
)

var _ = (*eventMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (e Event) MarshalJSON() ([]byte, error) {
	type Event struct {
		Address     common.Address    `json:"address" gencodec:"required"`
		Topics      []common.Hash     `json:"topics" gencodec:"required"`
		Data        hexutil.Bytes     `json:"data" gencodec:"required"`
		BlockHeight hexutil.Uint64Hex `json:"blockHeight"`
		TxHash      common.Hash       `json:"transactionHash" gencodec:"required"`
		TxIndex     hexutil.Uint64Hex `json:"transactionIndex" gencodec:"required"`
		BlockHash   common.Hash       `json:"blockHash"`
		Index       hexutil.Uint64Hex `json:"eventIndex" gencodec:"required"`
		Removed     bool              `json:"removed"`
	}
	var enc Event
	enc.Address = e.Address
	enc.Topics = e.Topics
	enc.Data = e.Data
	enc.BlockHeight = hexutil.Uint64Hex(e.BlockHeight)
	enc.TxHash = e.TxHash
	enc.TxIndex = hexutil.Uint64Hex(e.TxIndex)
	enc.BlockHash = e.BlockHash
	enc.Index = hexutil.Uint64Hex(e.Index)
	enc.Removed = e.Removed
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (e *Event) UnmarshalJSON(input []byte) error {
	type Event struct {
		Address     *common.Address    `json:"address" gencodec:"required"`
		Topics      []common.Hash      `json:"topics" gencodec:"required"`
		Data        *hexutil.Bytes     `json:"data" gencodec:"required"`
		BlockHeight *hexutil.Uint64Hex `json:"blockHeight"`
		TxHash      *common.Hash       `json:"transactionHash" gencodec:"required"`
		TxIndex     *hexutil.Uint64Hex `json:"transactionIndex" gencodec:"required"`
		BlockHash   *common.Hash       `json:"blockHash"`
		Index       *hexutil.Uint64Hex `json:"eventIndex" gencodec:"required"`
		Removed     *bool              `json:"removed"`
	}
	var dec Event
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Address == nil {
		return errors.New("missing required field 'address' for Event")
	}
	e.Address = *dec.Address
	if dec.Topics == nil {
		return errors.New("missing required field 'topics' for Event")
	}
	e.Topics = dec.Topics
	if dec.Data == nil {
		return errors.New("missing required field 'data' for Event")
	}
	e.Data = *dec.Data
	if dec.BlockHeight != nil {
		e.BlockHeight = uint32(*dec.BlockHeight)
	}
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionHash' for Event")
	}
	e.TxHash = *dec.TxHash
	if dec.TxIndex == nil {
		return errors.New("missing required field 'transactionIndex' for Event")
	}
	e.TxIndex = uint(*dec.TxIndex)
	if dec.BlockHash != nil {
		e.BlockHash = *dec.BlockHash
	}
	if dec.Index == nil {
		return errors.New("missing required field 'eventIndex' for Event")
	}
	e.Index = uint(*dec.Index)
	if dec.Removed != nil {
		e.Removed = *dec.Removed
	}
	return nil
}
