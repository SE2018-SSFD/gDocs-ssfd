## Lab 2 - Naive gDocs

```
518021910896 - 杨亘: {X}%
518021910270 - 蓝浩宁: {Y}%
518021911145 - 吴侃真: {Z}%
518021910917 - 莫戈泉: {Z}%
```
DFS部分，完整实现了文件系统基本接口, Chunk,  Fault , Consistency, Concurrency Control, Failure Recovery,  Checkpoint等所有功能(basic + advanced)，支持chunk多副本，chunkserver动态扩展、数据迁移、负载均衡，master借助kafka日志同步，集群之间借助zookeeper进行服务发现、leader选举以及心跳检测，也实现了backup-read,链式数据传播，checksum等额外功能。构造了一个强一致性，性能高效，可靠性稳定的分布式文件系统，并编写了详尽的测试，总代码量接近8000行。

gDocs部分，

