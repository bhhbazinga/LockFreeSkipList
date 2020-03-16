# LockFreeSkipList
A set implemented based on lockfree skiplist.
## Feature
    * The time complexity of all operations included Add,Remove,Contains are log(n).
    * Thread-safe and Lock-free.
    * Support Multi-producer and Multi-consumer.
##Benchmark
```
n represents the amount of data.
go test bench=. -args -n=1000000
```
```
goos: darwin
goarch: amd64
pkg: LockFreeSkipList
BenchmarkRandomAdd-8                                   1        2055942811 ns/op
BenchmarkRandomRemove-8                                1        2093779104 ns/op
BenchmarkRandomAddAndRemoveAndContains-8               1        5792781871 ns/op
PASS
ok      LockFreeSkipList        12.056s
```
The above data was tested on my 2013 macbook-pro with Intel Core i7 4 cores 2.3 GHz. \
See [benchmark](lockfree_skiplist_test.go).
## API
```golang
func (sl *LockFreeSkipList) Add(value interface{})(success bool)
func (sl *LockFreeSkipList) Remove(value interface{})(success bool)
func (sl *LockFreeSkipList) Contains(value interface{})(contains bool)
func (sl *LockFreeSkipList) GetSize(value interface{})(size int32)
```