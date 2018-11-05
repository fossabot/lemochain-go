// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package node

import (
	"encoding/json"
	"errors"

	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
)

var _ = (*ConfigFromFileMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (c ConfigFromFile) MarshalJSON() ([]byte, error) {
	type ConfigFromFile struct {
		ChainID   hexutil.Uint64 `json:"chainID"     gencodec:"required"`
		SleepTime hexutil.Uint64 `json:"sleepTime"   gencodec:"required"`
		Timeout   hexutil.Uint64 `json:"timeout"     gencodec:"required"`
	}
	var enc ConfigFromFile
	enc.ChainID = hexutil.Uint64(c.ChainID)
	enc.SleepTime = hexutil.Uint64(c.SleepTime)
	enc.Timeout = hexutil.Uint64(c.Timeout)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (c *ConfigFromFile) UnmarshalJSON(input []byte) error {
	type ConfigFromFile struct {
		ChainID   *hexutil.Uint64 `json:"chainID"     gencodec:"required"`
		SleepTime *hexutil.Uint64 `json:"sleepTime"   gencodec:"required"`
		Timeout   *hexutil.Uint64 `json:"timeout"     gencodec:"required"`
	}
	var dec ConfigFromFile
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.ChainID == nil {
		return errors.New("missing required field 'chainID' for ConfigFromFile")
	}
	c.ChainID = uint64(*dec.ChainID)
	if dec.SleepTime == nil {
		return errors.New("missing required field 'sleepTime' for ConfigFromFile")
	}
	c.SleepTime = uint64(*dec.SleepTime)
	if dec.Timeout == nil {
		return errors.New("missing required field 'timeout' for ConfigFromFile")
	}
	c.Timeout = uint64(*dec.Timeout)
	return nil
}
