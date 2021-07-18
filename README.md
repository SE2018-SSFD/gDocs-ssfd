## Lab 2 - Naive gDocs
```
518021910896 - 杨亘: 25%
518021910270 - 蓝浩宁: 25%
518021911145 - 吴侃真: 25%
518021910917 - 莫戈泉: 25%
```

本项目分为前端-->后端-->DFS三个部分。

gDocs前端需要负责后端数据的转换渲染，以及不同用户间协同编辑时的状态同步。gDocs的网页前端主要使用React编写，部分利用了Antd的组件库，整体风格与腾讯文档相似，分为登录/注册界面，主界面和文档界面，操作逻辑清晰明了。达成了除slides外的所有基本要求与bonus。

后端部分，在实现了要求的基本功能和附加功能的基础上，还额外的通过一致性哈希等技术实现了具有高可用性、一致性和可拓展性的分布式后端；实现了LRU驱逐机制的分布式缓存；实现了具有高性能的完全无锁的生产者消费者WebSocket架构。代码通过了详尽的单元测试、系统测试、端到端测试、压力测试，以及对WebSocket的性能测试。

DFS部分，完整实现了文件系统基本接口, Chunk, Fault , Consistency, Concurrency Control, Failure Recovery, Checkpoint等所有功能(basic + advanced)，支持chunk多副本，chunkserver动态扩展、数据迁移、负载均衡，master借助kafka日志同步，集群之间借助zookeeper进行服务发现、leader选举以及心跳检测，也实现了backup-read,链式数据传播，checksum等额外功能。构造了一个强一致性，性能高效，可靠性稳定的分布式文件系统，并编写了详尽的测试，总代码量接近8000行。
