// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"

	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/common/hexutil"
	"github.com/tomochain/tomochain/rlp"
)

//go:generate gencodec -type Receipt -field-override receiptMarshaling -out gen_receipt_json.go

var (
	receiptStatusFailedRLP     = []byte{}
	receiptStatusSuccessfulRLP = []byte{0x01}
)

const (
	// ReceiptStatusFailed is the status code of a transaction if execution failed.
	ReceiptStatusFailed = uint(0)

	// ReceiptStatusSuccessful is the status code of a transaction if execution succeeded.
	ReceiptStatusSuccessful = uint(1)
)

// Receipt represents the results of a transaction.
type Receipt struct {
	// Consensus fields
	PostState         []byte `json:"root"`
	Status            uint   `json:"status"`
	CumulativeGasUsed uint64 `json:"cumulativeGasUsed" gencodec:"required"`
	Bloom             Bloom  `json:"logsBloom"         gencodec:"required"`
	Logs              []*Log `json:"logs"              gencodec:"required"`

	// Implementation fields (don't reorder!)
	TxHash          common.Hash     `json:"transactionHash" gencodec:"required"`
	ContractAddress common.Address  `json:"contractAddress"`
	GasUsed         uint64          `json:"gasUsed" gencodec:"required"`
	IsSponsoredTx   *bool           `json:"isSponsoredTx,omitempty"` // nil = old receipt, non-nil = new receipt
	Payer           *common.Address `json:"payer,omitempty"`         // nil = old receipt, non-nil = new receipt
}

type receiptMarshaling struct {
	PostState         hexutil.Bytes
	Status            hexutil.Uint
	CumulativeGasUsed hexutil.Uint64
	GasUsed           hexutil.Uint64
	IsSponsoredTx     *bool
	Payer             *common.Address
}

// receiptRLP is the consensus encoding of a receipt.
type receiptRLP struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	Bloom             Bloom
	Logs              []*Log
}

type receiptStorageRLP struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	Bloom             Bloom
	TxHash            common.Hash
	ContractAddress   common.Address
	Logs              []*LogForStorage
	GasUsed           uint64
	IsSponsoredTx     *bool           // nil = old receipt, non-nil = new receipt
	Payer             *common.Address // nil = old receipt, non-nil = new receipt
}

type oldReceiptStorageRLP struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	Bloom             Bloom
	TxHash            common.Hash
	ContractAddress   common.Address
	Logs              []*LogForStorage
	GasUsed           uint64
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, failed bool, cumulativeGasUsed uint64) *Receipt {
	r := &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: cumulativeGasUsed}
	if failed {
		r.Status = ReceiptStatusFailed
	} else {
		r.Status = ReceiptStatusSuccessful
	}
	return r
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (r *Receipt) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &receiptRLP{r.statusEncoding(), r.CumulativeGasUsed, r.Bloom, r.Logs})
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (r *Receipt) DecodeRLP(s *rlp.Stream) error {
	var dec receiptRLP
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if err := r.setStatus(dec.PostStateOrStatus); err != nil {
		return err
	}
	r.CumulativeGasUsed, r.Bloom, r.Logs = dec.CumulativeGasUsed, dec.Bloom, dec.Logs
	return nil
}

func (r *Receipt) setStatus(postStateOrStatus []byte) error {
	switch {
	case bytes.Equal(postStateOrStatus, receiptStatusSuccessfulRLP):
		r.Status = ReceiptStatusSuccessful
	case bytes.Equal(postStateOrStatus, receiptStatusFailedRLP):
		r.Status = ReceiptStatusFailed
	case len(postStateOrStatus) == len(common.Hash{}):
		r.PostState = postStateOrStatus
	default:
		return fmt.Errorf("invalid receipt status %x", postStateOrStatus)
	}
	return nil
}

func (r *Receipt) statusEncoding() []byte {
	if len(r.PostState) == 0 {
		if r.Status == ReceiptStatusFailed {
			return receiptStatusFailedRLP
		}
		return receiptStatusSuccessfulRLP
	}
	return r.PostState
}

// Size returns the approximate memory used by all internal contents. It is used
// to approximate and limit the memory consumption of various caches.
func (r *Receipt) Size() common.StorageSize {
	size := common.StorageSize(unsafe.Sizeof(*r)) + common.StorageSize(len(r.PostState))

	size += common.StorageSize(len(r.Logs)) * common.StorageSize(unsafe.Sizeof(Log{}))
	for _, log := range r.Logs {
		size += common.StorageSize(len(log.Topics)*common.HashLength + len(log.Data))
	}
	return size
}

// String implements the Stringer interface.
func (r *Receipt) String() string {
	if len(r.PostState) == 0 {
		return fmt.Sprintf("receipt{status=%d cgas=%v bloom=%x logs=%v}", r.Status, r.CumulativeGasUsed, r.Bloom, r.Logs)
	}
	return fmt.Sprintf("receipt{med=%x cgas=%v bloom=%x logs=%v}", r.PostState, r.CumulativeGasUsed, r.Bloom, r.Logs)
}

// ReceiptForStorage is a wrapper around a Receipt that flattens and parses the
// entire content of a receipt, as opposed to only the consensus fields originally.
type ReceiptForStorage Receipt

// EncodeRLP implements rlp.Encoder, and flattens all content fields of a receipt
// into an RLP stream.
func (r *ReceiptForStorage) EncodeRLP(w io.Writer) error {
	// For old receipts (nil pointers), encode without IsSponsoredTx and Payer
	// For new receipts (non-nil pointers), include them
	if r.IsSponsoredTx == nil && r.Payer == nil {
		// Old receipt format - encode without IsSponsoredTx and Payer
		oldEnc := &oldReceiptStorageRLP{
			PostStateOrStatus: (*Receipt)(r).statusEncoding(),
			CumulativeGasUsed: r.CumulativeGasUsed,
			Bloom:             r.Bloom,
			TxHash:            r.TxHash,
			ContractAddress:   r.ContractAddress,
			Logs:              make([]*LogForStorage, len(r.Logs)),
			GasUsed:           r.GasUsed,
		}
		for i, log := range r.Logs {
			oldEnc.Logs[i] = (*LogForStorage)(log)
		}
		return rlp.Encode(w, oldEnc)
	}

	// New receipt format - encode with IsSponsoredTx and Payer
	// Both fields must be non-nil for new format
	enc := &receiptStorageRLP{
		PostStateOrStatus: (*Receipt)(r).statusEncoding(),
		CumulativeGasUsed: r.CumulativeGasUsed,
		Bloom:             r.Bloom,
		TxHash:            r.TxHash,
		ContractAddress:   r.ContractAddress,
		Logs:              make([]*LogForStorage, len(r.Logs)),
		GasUsed:           r.GasUsed,
		IsSponsoredTx:     r.IsSponsoredTx,
		Payer:             r.Payer,
	}
	for i, log := range r.Logs {
		enc.Logs[i] = (*LogForStorage)(log)
	}
	return rlp.Encode(w, enc)
}

// DecodeRLP implements rlp.Decoder, and loads both consensus and implementation
// fields of a receipt from an RLP stream.
// Supports backward compatibility: tries new format first, falls back to old format.
func (r *ReceiptForStorage) DecodeRLP(s *rlp.Stream) error {
	// Get raw bytes first so we can try both formats
	rawBytes, err := s.Raw()
	if err != nil {
		return err
	}

	// Try to decode as new format first (with IsSponsoredTx and Payer)
	var dec receiptStorageRLP
	newStream := rlp.NewStream(bytes.NewReader(rawBytes), 0)
	if err := newStream.Decode(&dec); err == nil {
		// Successfully decoded as new format
		// Set status and assign fields
		if err := (*Receipt)(r).setStatus(dec.PostStateOrStatus); err != nil {
			return err
		}
		// Assign the consensus fields
		r.CumulativeGasUsed, r.Bloom = dec.CumulativeGasUsed, dec.Bloom
		r.Logs = make([]*Log, len(dec.Logs))
		for i, log := range dec.Logs {
			r.Logs[i] = (*Log)(log)
		}
		// Assign the implementation fields
		r.TxHash, r.ContractAddress, r.GasUsed = dec.TxHash, dec.ContractAddress, dec.GasUsed
		r.IsSponsoredTx = dec.IsSponsoredTx
		r.Payer = dec.Payer
		return nil
	}

	// New format failed, try old format
	var oldDec oldReceiptStorageRLP
	oldStream := rlp.NewStream(bytes.NewReader(rawBytes), 0)
	if err := oldStream.Decode(&oldDec); err != nil {
		// Both formats failed, return the original error
		return err
	}

	// Successfully decoded as old format, convert to new format structure
	dec = receiptStorageRLP{
		PostStateOrStatus: oldDec.PostStateOrStatus,
		CumulativeGasUsed: oldDec.CumulativeGasUsed,
		Bloom:             oldDec.Bloom,
		TxHash:            oldDec.TxHash,
		ContractAddress:   oldDec.ContractAddress,
		Logs:              oldDec.Logs,
		GasUsed:           oldDec.GasUsed,
		IsSponsoredTx:     nil,
		Payer:             nil,
	}

	// Set status and assign fields
	if err := (*Receipt)(r).setStatus(dec.PostStateOrStatus); err != nil {
		return err
	}
	// Assign the consensus fields
	r.CumulativeGasUsed, r.Bloom = dec.CumulativeGasUsed, dec.Bloom
	r.Logs = make([]*Log, len(dec.Logs))
	for i, log := range dec.Logs {
		r.Logs[i] = (*Log)(log)
	}
	// Assign the implementation fields
	r.TxHash, r.ContractAddress, r.GasUsed = dec.TxHash, dec.ContractAddress, dec.GasUsed
	r.IsSponsoredTx = nil
	r.Payer = nil
	return nil
}

// Receipts is a wrapper around a Receipt array to implement DerivableList.
type Receipts []*Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }

// GetRlp returns the RLP encoding of one receipt from the list.
func (r Receipts) GetRlp(i int) []byte {
	bytes, err := rlp.EncodeToBytes(r[i])
	if err != nil {
		panic(err)
	}
	return bytes
}
