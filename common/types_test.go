// Copyright 2015 The go-ethereum Authors
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

package common

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func TestBytesConversion(t *testing.T) {
	bytes := []byte{5}
	hash := BytesToHash(bytes)

	var exp Hash
	exp[31] = 5

	if hash != exp {
		t.Errorf("expected %x got %x", exp, hash)
	}
}

func TestIsHexAddress(t *testing.T) {
	tests := []struct {
		str string
		exp bool
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"0X5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"0XAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1", false},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beae", false},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed11", false},
		{"0xxaaeb6053f3e94c9b9a09f33669435e7ef1beaed", false},
	}

	for _, test := range tests {
		if result := IsHexAddress(test.str); result != test.exp {
			t.Errorf("IsHexAddress(%s) == %v; expected %v",
				test.str, result, test.exp)
		}
	}
}

func TestHashJsonValidation(t *testing.T) {
	var tests = []struct {
		Prefix string
		Size   int
		Error  string
	}{
		{"", 62, "json: cannot unmarshal hex string without 0x prefix into Go value of type common.Hash"},
		{"0x", 66, "hex string has length 66, want 64 for common.Hash"},
		{"0x", 63, "json: cannot unmarshal hex string of odd length into Go value of type common.Hash"},
		{"0x", 0, "hex string has length 0, want 64 for common.Hash"},
		{"0x", 64, ""},
		{"0X", 64, ""},
	}
	for _, test := range tests {
		input := `"` + test.Prefix + strings.Repeat("0", test.Size) + `"`
		var v Hash
		err := json.Unmarshal([]byte(input), &v)
		if err == nil {
			if test.Error != "" {
				t.Errorf("%s: error mismatch: have nil, want %q", input, test.Error)
			}
		} else {
			if err.Error() != test.Error {
				t.Errorf("%s: error mismatch: have %q, want %q", input, err, test.Error)
			}
		}
	}
}

func TestAddressUnmarshalJSON(t *testing.T) {
	var tests = []struct {
		Input     string
		ShouldErr bool
		Output    *big.Int
	}{
		{"", true, nil},
		{`""`, true, nil},
		{`"0x"`, true, nil},
		{`"0x00"`, true, nil},
		{`"0xG000000000000000000000000000000000000000"`, true, nil},
		{`"0x0000000000000000000000000000000000000000"`, false, big.NewInt(0)},
		{`"0x0000000000000000000000000000000000000010"`, false, big.NewInt(16)},
	}
	for i, test := range tests {
		var v Address
		err := json.Unmarshal([]byte(test.Input), &v)
		if err != nil && !test.ShouldErr {
			t.Errorf("test #%d: unexpected error: %v", i, err)
		}
		if err == nil {
			if test.ShouldErr {
				t.Errorf("test #%d: expected error, got none", i)
			}
			if v.Big().Cmp(test.Output) != 0 {
				t.Errorf("test #%d: address mismatch: have %v, want %v", i, v.Big(), test.Output)
			}
		}
	}
}

func TestAddressHexChecksum(t *testing.T) {
	var tests = []struct {
		Input  string
		Output string
	}{
		// Test cases from https://github.com/ethereum/EIPs/blob/master/EIPS/eip-55.md#specification
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"},
		{"0xfb6916095ca1df60bb79ce92ce3ea74c37c5d359", "0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359"},
		{"0xdbf03b407c01e7cd3cbea99509d93f8dddc8c6fb", "0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB"},
		{"0xd1220a0cf47c7b9be7a2e6ba89f429762e7b9adb", "0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"},
		// Ensure that non-standard length input values are handled correctly
		{"0xa", "0x000000000000000000000000000000000000000A"},
		{"0x0a", "0x000000000000000000000000000000000000000A"},
		{"0x00a", "0x000000000000000000000000000000000000000A"},
		{"0x000000000000000000000000000000000000000a", "0x000000000000000000000000000000000000000A"},
	}
	for i, test := range tests {
		output := HexToAddress(test.Input).Hex()
		if output != test.Output {
			t.Errorf("test #%d: failed to match when it should (%s != %s)", i, output, test.Output)
		}
	}
}

func BenchmarkAddressHex(b *testing.B) {
	testAddr := HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	for n := 0; n < b.N; n++ {
		testAddr.Hex()
	}
}

func TestRemoveItemInArray(t *testing.T) {
	array := []Address{HexToAddress("0x0000003"), HexToAddress("0x0000001"), HexToAddress("0x0000002"), HexToAddress("0x0000003")}
	remove := []Address{HexToAddress("0x0000002"), HexToAddress("0x0000004"), HexToAddress("0x0000003")}
	array = RemoveItemFromArray(array, remove)
	if len(array) != 1 {
		t.Error("fail remove item from array address")
	}
}

func TestGetPen(t *testing.T) {
	penHeaders := "0x255fca57249b5747e11be0b4fc9ad2d0ecfcda881c4006de9cb80bae42c8921933285b819f3bcf0a615ad0d8e40d709af9b015be131e4e81637208c3c7db2f9977488dca3a17ce3b83e417c03468dbf68cb1883564db887a1d2e52311bed01664f921c21"
	ls := FromHex(penHeaders)
	fmt.Println("-> ls", ls)
	penlist := ExtractAddressFromBytes(ls)
	for i := range penlist {
		fmt.Println("-> pen", penlist[i].String())
	}
}

const (
	extraVanity = 32 // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = 65 // Fixed number of extra-data suffix bytes reserved for signer seal
)

func TestGetM1s(t *testing.T) {
	extra := "0xda8302040084746f6d6f89676f312e31382e31308664617277696e00000000008cb1883564db887a1d2e52311bed01664f921c21c7db2f9977488dca3a17ce3b83e417c03468dbf62c182a3dba2755fa630c3413d38d320ff1588b8d71da1c0977d305dd85fa290b6cba332c0d6f0e841c892f1029b592ffbfc0fb61f65f252439f7fa06f76162a601"
	ls := FromHex(extra)
	extraSuffix := len(ls) - extraSeal
	fmt.Println("extraSuffix", extraSuffix)
	masternodesFromCheckpointHeader := ExtractAddressFromBytes(ls[extraVanity:extraSuffix])
	fmt.Println("masternodes", masternodesFromCheckpointHeader)
	for i := range masternodesFromCheckpointHeader {
		fmt.Println("M1", masternodesFromCheckpointHeader[i].String(), masternodesFromCheckpointHeader[i])
	}
}

func TestVerifyCheckpoint(t *testing.T) {
	signers := []Address{
		HexToAddress("0x255fCa57249b5747e11Be0b4fC9Ad2D0eCFCDa88"),
		HexToAddress("0x1c4006De9CB80BAe42c8921933285B819F3bCF0A"),
		HexToAddress("0x118Db4a6718E79A8cA2A05A5def0e6AfeaAF24F4"),
		HexToAddress("0x615ad0D8e40D709aF9B015BE131e4e81637208C3"),
		HexToAddress("0x596571D3f8B5903d908A7BD6bB9e96BCdE691581"),
		HexToAddress("0x322a8D78955774256A3174DB0ABB9cfc75211759"),
		HexToAddress("0xcE55BF99666FBBA399260c53A05863F8Adc7b121"),
		HexToAddress("0xC7db2f9977488DCa3A17cE3B83E417C03468DBf6"),
		HexToAddress("0x8Cb1883564Db887A1D2e52311beD01664f921c21"),
	}

	//newSigners := make([]Address, len(signers))
	//copy(newSigners, signers)

	cmd := []Address{
		HexToAddress("0x596571d3f8b5903d908a7bd6bb9e96bcde691581"),
		HexToAddress("0x118db4a6718e79a8ca2a05a5def0e6afeaaf24f4"),
		HexToAddress("0x322a8d78955774256a3174db0abb9cfc75211759"),
		HexToAddress("0xce55bf99666fbba399260c53a05863f8adc7b121"),
	}
	pen := RemoveItemFromArray(signers, cmd)
	for i := range pen {
		fmt.Println("-> pen", pen[i].String())
	}
	for i := range signers {
		fmt.Println("-> signers", signers[i].String())
	}
	//for i := range newSisigners {
	//	fmt.Println("-> signers", newSigners[i].String())
	//}
}
