# MemDB: In-Memory Database Library

MemDB is a lightweight, in-memory database library designed for Go, offering simplicity and flexibility. It facilitates the creation and management of document collections, efficient indexing for rapid querying, and supports complex searches based on filters.

## Features

- **In-Memory Storage**: MemDB stores data entirely in memory for rapid access.
- **Indexing**: Utilizes red-black trees to efficiently index documents, speeding up search operations.
- **Filtering**: Supports complex filter conditions to retrieve documents matching specific criteria.
- **Execution Plan**: Generates optimized execution plans based on filter conditions to enhance search performance.

## Benchmarks

The benchmarks below demonstrate the efficiency of MemDB operations across various scenarios:

```bash
goos: linux
goarch: amd64
pkg: github.com/siyul-park/uniflow/database/memdb
cpu: AMD EPYC 7282 16-Core Processor

BenchmarkCollection_InsertOne-4             55,045        20,448 ns/op       4,329 B/op       94 allocs/op
BenchmarkCollection_InsertMany-4                64    20,625,413 ns/op   4,370,966 B/op   94,016 allocs/op
BenchmarkCollection_UpdateOne-4               1,368       813,229 ns/op      69,702 B/op    5,116 allocs/op
BenchmarkCollection_UpdateMany-4                121     8,906,252 ns/op     767,224 B/op   28,101 allocs/op
BenchmarkCollection_DeleteOne-4            100,551        12,179 ns/op         520 B/op       21 allocs/op
BenchmarkCollection_DeleteMany-4               246     4,545,208 ns/op     303,756 B/op   13,018 allocs/op
BenchmarkCollection_FindOne/With_Index-4   294,973         3,601 ns/op         400 B/op       15 allocs/op
BenchmarkCollection_FindOne/Without_Index-4   1,880      669,318 ns/op      65,096 B/op    5,014 allocs/op
BenchmarkCollection_FindMany/With_Index-4  304,732         3,341 ns/op         392 B/op       14 allocs/op
BenchmarkCollection_FindMany/Without_Index-4   2,116      663,374 ns/op      65,040 B/op    5,013 allocs/op
```

These benchmarks highlight MemDB's efficiency in handling various operations, showcasing its suitability for applications requiring fast in-memory data management and querying capabilities.