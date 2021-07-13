#go test tests/* -coverprofile=coverage.out -coverpkg=./...
#go tool cover -func=coverage.out
#sleep 2
#go tool cover -html=coverage.out

#go test -v tests/system_test/* -test.run TestLogin -coverpkg=./...

# system test
export CGO_CPPFLAGS="-Wno-error -Wno-nullability-completeness -Wno-expansion-to-defined -Wno-builtin-requires-header"
#go test ../tests/concurrency_test/*.go -benchtime=40s -bench=. -coverprofile=coverage.out -v
#go test ../tests/concurrency_test -coverprofile=coverage.out -coverpkg=../... -v
go test ../tests/dfs_test -coverprofile=coverage.out -coverpkg=../... -v
#go test ../tests/lib_test/sheetCache_test -coverprofile=coverage.out -coverpkg=../... -v -race
#go test ../tests/lib_test/algorithm_test/lru_test* -coverpkg=../... -v
##go test ../tests/lib_test/algorithm_test/lru_test* -coverpkg=../... -coverprofile=coverage.out -v
#go tool cover -func=coverage.out
#go tool cover -html=coverage.out