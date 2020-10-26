# 啟動
`go run cmd/server/main.go`
listen on 127.0.0.1:80

# testing
## with go test
`go test -v acceptance_test.go -race`
用go test library 完成，並使用 data race detector

## with apache bench
`ab -n 310 -c 5 -k  127.0.0.1/`
依需求，會有如下結果

Complete requests:      310
Non-2xx responses:      10

# storage 說明
使用簡單的 map[string]int 記錄 request 數量

- 保持能達成需求所需要的最小限度的 code ，速度快、成本也低
- 萬一出了問題需要重置，request 數量丟失不會造成致命的影響
