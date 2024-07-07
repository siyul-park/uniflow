# MemDB: 인메모리 데이터베이스 라이브러리

MemDB는 Go 언어로 작성된 가벼운 인메모리 데이터베이스 라이브러리로, 간단하고 유연한 데이터 관리를 제공합니다. 문서 컬렉션의 생성 및 관리, 빠른 쿼리를 위한 효율적인 인덱싱, 복잡한 조건에 기반한 검색 지원을 특징으로 합니다.

## 주요 기능

- **인메모리 저장**: MemDB는 데이터를 전적으로 메모리에 저장하여 빠른 접근을 가능하게 합니다.
- **인덱싱**: 레드-블랙 트리를 활용하여 문서를 효율적으로 인덱싱하여 검색 작업을 가속화합니다.
- **필터링**: 특정 조건을 충족하는 문서를 검색하기 위해 복잡한 필터 조건을 지원합니다.
- **실행 계획**: 필터 조건에 기반한 최적화된 실행 계획을 생성하여 검색 성능을 향상시킵니다.

## 벤치마크

아래는 다양한 시나리오에서 MemDB 작업의 효율성을 보여주는 벤치마크입니다:

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

이 벤치마크는 MemDB가 다양한 작업을 효율적으로 처리하는 능력을 보여줍니다. 이는 메모리 내 데이터 관리와 빠른 쿼리 기능이 필요한 애플리케이션에 적합함을 시사합니다.
