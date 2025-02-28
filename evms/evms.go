// Copyright 2019 Martin Holst Swende
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package evms

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// The Evm interface represents external EVM implementations, which can
// be e.g. docker instances or binaries
type Evm interface {
	// RunStateTest runs the statetest on the underlying EVM, and writes
	// the output to the given writer
	RunStateTest(path string, writer io.Writer, skipTrace bool) (cmd string, err error)
	// GetStateRoot runs the test and returns the stateroot
	GetStateRoot(path string) (root, command string, err error)
	// ParseStateRoot reads the stateroot from the combined output.
	ParseStateRoot([]byte) (string, error)
	// Copy takes the 'raw' output from the VM, and writes the
	// canonical output to the given writer
	Copy(out io.Writer, input io.Reader)
	//Open() // Preparare for execution
	Close() // Tear down processes
	Name() string
}

type stateRoot struct {
	StateRoot string `json:"stateRoot"`
}

// CompareFiles returns true if the files are equal, along with the number of line s
// compared
func CompareFiles(vms []Evm, readers []io.Reader) (bool, int) {
	var count = 0
	var scanners []*bufio.Scanner
	for _, r := range readers {
		scanners = append(scanners, bufio.NewScanner(r))
	}
	refOut := scanners[0]
	refVM := vms[0]
	for refOut.Scan() {
		for i, scanner := range scanners[1:] {
			scanner.Scan()
			if !bytes.Equal(refOut.Bytes(), scanner.Bytes()) {
				fmt.Printf("diff: \n%15v: %v\n%15v: %v\n",
					refVM.Name(),
					string(refOut.Bytes()),
					vms[i+1].Name(),
					string(scanner.Bytes()))
				return false, count
			}
		}
		count++
	}
	// The source is 'done', need to also check if the other scanners are done
	for i, scanner := range scanners[1:] {
		if scanner.Scan() {
			fmt.Printf("diff: \n%15v: %v\n%15v: %v\n",
				refVM.Name(),
				string("--  depleted --"),
				vms[i+1].Name(),
				string(scanner.Bytes()))
			return false, count
		}
	}
	return true, count
}
