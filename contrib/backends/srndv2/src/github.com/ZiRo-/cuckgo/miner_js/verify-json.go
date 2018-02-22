/*
The MIT License (MIT)

Copyright (c) 2016 ZiRo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"crypto/sha256"
	"encoding/json"
)

type CuckooJSON struct {
	Parameter map[string]uint64 `json:"parameters"`
	InputData []byte            `json:"header"` //sha256 bytes
	Cycle     []uint64          `json:"cycle"`
}

func DecodeCuckooJSON(jsonBlob []byte) (*CuckooJSON, error) {
	res := new(CuckooJSON)
	err := json.Unmarshal(jsonBlob, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func EncodeCuckooJSON(cuck CuckooJSON) ([]byte, error) {
	return json.Marshal(cuck)
}

func VerifyJSON(cuck CuckooJSON) bool {
	easy := cuck.Parameter["easiness"]

	if easy <= 0 || easy >= SIZE {
		return false
	}

	if len(cuck.InputData) != sha256.Size {
		return false
	}

	if uint64(len(cuck.Cycle)) != PROOFSIZE {
		return false
	}

	var sha [sha256.Size]byte
	copy(sha[:], cuck.InputData)

	c := NewCuckooSHA(sha)
	return c.Verify(cuck.Cycle, easy)
}
