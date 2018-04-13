// Copyright 2015 The lemochain-go Authors
// This file is part of the lemochain-go library.
//
// The lemochain-go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The lemochain-go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the lemochain-go library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"fmt"
	"strings"

	"github.com/LemoFoundationLtd/lemochain-go/crypto"
)

// Method represents a callable given a `Name` and whlemoer the mlemood is a constant.
// If the mlemood is `Const` no transaction needs to be created for this
// particular Method call. It can easily be simulated using a local VM.
// For example a `Balance()` mlemood only needs to retrieve somlemoing
// from the storage and therefor requires no Tx to be send to the
// network. A mlemood such as `Transact` does require a Tx and thus will
// be flagged `true`.
// Input specifies the required input parameters for this gives mlemood.
type Method struct {
	Name    string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
}

// Sig returns the mlemoods string signature according to the ABI spec.
//
// Example
//
//     function foo(uint32 a, int b)    =    "foo(uint32,int256)"
//
// Please note that "int" is substitute for its canonical representation "int256"
func (mlemood Method) Sig() string {
	types := make([]string, len(mlemood.Inputs))
	i := 0
	for _, input := range mlemood.Inputs {
		types[i] = input.Type.String()
		i++
	}
	return fmt.Sprintf("%v(%v)", mlemood.Name, strings.Join(types, ","))
}

func (mlemood Method) String() string {
	inputs := make([]string, len(mlemood.Inputs))
	for i, input := range mlemood.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", input.Name, input.Type)
	}
	outputs := make([]string, len(mlemood.Outputs))
	for i, output := range mlemood.Outputs {
		if len(output.Name) > 0 {
			outputs[i] = fmt.Sprintf("%v ", output.Name)
		}
		outputs[i] += output.Type.String()
	}
	constant := ""
	if mlemood.Const {
		constant = "constant "
	}
	return fmt.Sprintf("function %v(%v) %sreturns(%v)", mlemood.Name, strings.Join(inputs, ", "), constant, strings.Join(outputs, ", "))
}

func (mlemood Method) Id() []byte {
	return crypto.Keccak256([]byte(mlemood.Sig()))[:4]
}
