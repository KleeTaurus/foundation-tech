# 代码示例

* `benchmark.go` 给予 iohub/ahocorasick 的性能测试，写入500万数据后测试查找的速度，其测试结果小于50us。
* `blacklist.go` 基于 iohub/ahocorasick 实现的用户黑白名单服务，提供一次性构建初始化黑名单列表，同时支持用户自定义数据信息的存储，对于构建完成的树该库还提供了追加、更新和删除节点的接口。
