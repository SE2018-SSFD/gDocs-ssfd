## Naive gDocs

Give a brief introduction to your project here.

## File Tree


## Packages
### logrus 
a powerful open-source go logger  
https://github.com/sirupsen/logrus/
### rpc
go RPC package  
```go
// m should at least has one method observing the RPC method standard, or the Register will failed 
// non-RPC method will report "method xxx has x input parameters; needs exactly three". It is normal.
var m class
rpc.Register(&m)
```

## 笔记
零散的记录一点实现过程中的思考，最后再统一
1. 回收站功能：后端在收到第二次delete时才向DFS发Delete操作
2. 为了保证并发写，master管理namespace时，应该对写请求加读锁
