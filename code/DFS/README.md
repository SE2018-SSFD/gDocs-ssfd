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
3. goroutine是协程，和主线程是等价的
```go
package doc
func IsClosed(c chan int) bool {
	select {
	case <-c:
		return true
	default:
	}
	return false
}
```
4. Tricky pointer
```go
// when reply is a function param...
*reply = util.GetFileMetaRet{
	Exist: false,
	IsDir: false,
	Size: -1,
}//correct
reply = &util.GetFileMetaRet{
Exist: false,
IsDir: false,
Size: -1,
}//incorrect,doesn't change the value
```
5. defer  
defer语句需要执行到才有效果（执行前return了则无效果），执行时已经算好了该语句的值

6. DFS删除也是二次删除法，删除时只是改变映射，这样避免了不少同步错误（比如拿着大锁删除映射并删除整个filestate，但此时有人要修改文件则内存错误
## lock
1. 全局拿锁顺序一致 从大锁->小锁
2. 一个锁保护一个共享变量

## TODO
日志 租约 小粒度锁 多client
全局ckp，同步 等等