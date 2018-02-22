
Cuckoo Cycle
============
# Go implementation
This is a Go implementation of the cuckoo Proof of Work algorith "cuckoo cycle" by John Tromp.
This implementation includes the verification function as a library, and a simple multi-threaded miner.

# Installation
To install this miner just run
`go get github.com/ZiRo-/cuckgo/miner`

# Usage

```
Usage of miner:
  -e float
    	easiness in percentage (default 50)
  -m int
    	maximum number of solutions (default 8)
  -t int
    	number of miner threads (default 4)
```
# Algorithm

For details about the algorith, see: https://github.com/tromp/cuckoo

# Benchmarks
#### N = 100, Easiness = 55, Average time = 0.9 s, CPU = i5-4210U
![N = 100, Easiness = 55, Average time = 0.9 s, CPU = i5-4210U] (https://raw.githubusercontent.com/ZiRo-/cuckgo/master/bench/bench_N_100_e_55_avg_0.9.png)

#### N = 100, Easiness = 50, Average time = 2.3 s, CPU = i5-4210U
![N = 100, Easiness = 50, Average time = 2.3 s, CPU = i5-4210U] (https://raw.githubusercontent.com/ZiRo-/cuckgo/master/bench/bench_N_100_e_50_avg_2.3.png)

#### N = 300, Easiness = 46, Average time = 56.2 s, CPU = i5-4210U
![N = 300, Easiness = 46, Average time = 56.2 s, CPU = i5-4210U] (https://raw.githubusercontent.com/ZiRo-/cuckgo/master/bench/bench_N_300_e_46_avg_56.2.png)

# JavaScript version
Using [GopherJS](https://github.com/gopherjs/gopherjs) the Go implementation can be transpiled to JavaScript,
which is also included in this repo. This file currently exports one function:
`cuckoo["mine_cuckoo"](easiness)`
which returns the same base64 representation of a proof as the Go CLI miner.
