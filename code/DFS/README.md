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
