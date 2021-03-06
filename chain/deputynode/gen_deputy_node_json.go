// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package deputynode

import (
	"encoding/json"
	"errors"
	"net"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
)

var _ = (*deputyNodeMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (d DeputyNode) MarshalJSON() ([]byte, error) {
	type DeputyNode struct {
		MinerAddress common.Address `json:"minerAddress"   gencodec:"required"`
		NodeID       hexutil.Bytes  `json:"nodeID"         gencodec:"required"`
		IP           hexutil.IP     `json:"ip"             gencodec:"required"`
		Port         hexutil.Uint32 `json:"port"           gencodec:"required"`
		Rank         hexutil.Uint32 `json:"rank"           gencodec:"required"`
		Votes        hexutil.Uint32 `json:"votes"          gencodec:"required"`
	}
	var enc DeputyNode
	enc.MinerAddress = d.MinerAddress
	enc.NodeID = d.NodeID
	enc.IP = hexutil.IP(d.IP)
	enc.Port = hexutil.Uint32(d.Port)
	enc.Rank = hexutil.Uint32(d.Rank)
	enc.Votes = hexutil.Uint32(d.Votes)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (d *DeputyNode) UnmarshalJSON(input []byte) error {
	type DeputyNode struct {
		MinerAddress *common.Address `json:"minerAddress"   gencodec:"required"`
		NodeID       *hexutil.Bytes  `json:"nodeID"         gencodec:"required"`
		IP           *hexutil.IP     `json:"ip"             gencodec:"required"`
		Port         *hexutil.Uint32 `json:"port"           gencodec:"required"`
		Rank         *hexutil.Uint32 `json:"rank"           gencodec:"required"`
		Votes        *hexutil.Uint32 `json:"votes"          gencodec:"required"`
	}
	var dec DeputyNode
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.MinerAddress == nil {
		return errors.New("missing required field 'minerAddress' for DeputyNode")
	}
	d.MinerAddress = *dec.MinerAddress
	if dec.NodeID == nil {
		return errors.New("missing required field 'nodeID' for DeputyNode")
	}
	d.NodeID = *dec.NodeID
	if dec.IP == nil {
		return errors.New("missing required field 'ip' for DeputyNode")
	}
	d.IP = net.IP(*dec.IP)
	if dec.Port == nil {
		return errors.New("missing required field 'port' for DeputyNode")
	}
	d.Port = uint32(*dec.Port)
	if dec.Rank == nil {
		return errors.New("missing required field 'rank' for DeputyNode")
	}
	d.Rank = uint32(*dec.Rank)
	if dec.Votes == nil {
		return errors.New("missing required field 'votes' for DeputyNode")
	}
	d.Votes = uint32(*dec.Votes)
	return nil
}
