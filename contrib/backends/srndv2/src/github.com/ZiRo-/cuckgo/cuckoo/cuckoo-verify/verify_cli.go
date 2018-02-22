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
	"encoding/base64"
	"fmt"
	"github.com/ZiRo-/cuckgo/cuckoo"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(3)
	}

	data, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(2)
	}
	cuck, err := cuckoo.DecodeCuckooJSON(data)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(2)
	}
	//fmt.Println(cuck)

	if cuckoo.VerifyJSON(*cuck) {
		fmt.Println("Valid")
		os.Exit(0)
	} else {
		fmt.Println("Invalid")
		os.Exit(1)
	}
}
