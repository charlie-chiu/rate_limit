# 啟動
`go run cmd/server/main.go`
listen on 127.0.0.1:80

# 測試
## with browser
很直覺的，用任何 browser 直接打開 `127.0.0.1` 測試，可能會因為 GET /favicon.ico 產生多一個 request

## with go test library
`go test -v acceptance_test.go -race`

用go test library 完成，並使用 data race detector


## with apache bench
`ab -n 310 -c 5 -k  127.0.0.1/`
依需求，會有如下結果

Complete requests:      310
Non-2xx responses:      10

5條連線 * 接受60個requests = 完成300個 request

10 個超過 limit 的 request return 249

共完成 310 個 request 

# 設計說明
## 演算法
最直覺的作法是記錄每個 request 的時間，收到 request 時再 count 60 秒內的 request 數量，但這需要大量的計算

所以實作上每個 ip 依秒來儲存 request 數，會失去一些精確度但減少很多重複計算的工

## storage
使用簡單的 map variable 記錄 request 數量，保持能達成需求所需要的最小限度的 code ，速度快、成本也低

如果需求有改變，例如說時間從60秒改為一天，或是 request 量是對客戶計價的，可以改用 redis 之類的 in-memory db


# 可調校
- 舊的 request 數量沒有清理機制，會一直累加
- mutex.Lock / Unlock 會對整個 map 動作，但不同的 ip 可以分開處理
- go test 的花費時間很長，因記錄的粒度是每秒，測試整個流程就需要2秒以上的時間，可以改用 dependency injection 處理 
