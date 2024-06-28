# MemDB: In-Memory Database Library

MemDB is a lightweight, in-memory database library for Go, designed with simplicity and flexibility in mind. It facilitates the creation and management of document collections, indexing for efficient querying, and complex searches based on filters.

## Features

- **In-Memory**: MemDB stores data entirely in memory for quick access.
- **Indexing**: Efficiently indexes documents using red-black trees, accelerating search operations.
- **Filtering**: Supports complex filter conditions for retrieving documents that match specific criteria.
- **Execution Plan**: Generates an execution plan based on filter conditions to optimize search performance.

## Benchmarks

These benchmarks evaluate the efficiency of MemDB operations under various scenarios:

```bash
goos: linux
goarch: amd64
pkg: github.com/siyul-park/uniflow/database/memdb
cpu: AMD EPYC 7282 16-Core Processor                
BenchmarkCollection_InsertOne-4            55045             20448 ns/op            4329 B/op         94 allocs/op
BenchmarkCollection_InsertMany-4              64          20625413 ns/op         4370966 B/op      94016 allocs/op
BenchmarkCollection_UpdateOne-4             1368            813229 ns/op           69702 B/op       5116 allocs/op
BenchmarkCollection_UpdateMany-4             121           8906252 ns/op          767224 B/op      28101 allocs/op
BenchmarkCollection_DeleteOne-4           100551             12179 ns/op             520 B/op         21 allocs/op
BenchmarkCollection_DeleteMany-4             246           4545208 ns/op          303756 B/op      13018 allocs/op
BenchmarkCollection_FindOne/With_Index-4                  294973              3601 ns/op             400 B/op         15 allocs/op
BenchmarkCollection_FindOne/Without_Index-4                 1880            669318 ns/op           65096 B/op       5014 allocs/op
BenchmarkCollection_FindMany/With_Index-4                 304732              3341 ns/op             392 B/op         14 allocs/op
BenchmarkCollection_FindMany/Without_Index-4                2116            663374 ns/op           65040 B/op       5013 allocs/op
```
