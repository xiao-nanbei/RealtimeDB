# DataBrainsDB 时序数据库设计与实现


## 一.整体架构

### 数据结构设计

```
series1: {"metric": "cpu.busy", "host": "localhost", "iface": "eth0"}
series2: {"metric": "cpu.busy", "host": "localhost", "iface": "eth1"}
```
```go
Point: 
type Point struct {
	TimeStamp int64
	Value float64//为了简化实现,我们先选择float64作为唯一的value类型
}
Tag: 
type Tag struct {
	Name string
	Value string
}//这里的tag采用了influxdb的设计思路
//example:
//name:host
//value:linux-001
type Row struct {
	Metric string
	Tags TagSet
	Point Point
}
```

### 持久化设计



## 优化

### 压缩算法优化
#### 磁盘压缩（二次压缩）优化
1. ZstdBytesCompressor 使用 ZSTD 算法压缩
2. ppyBytesCompressor 使用 Snappy 算法压缩
3. Simple8bBytesCompressor
4. GzipBytesCompressor
5. ZipBytesCompressor
#### 内存压缩优化
1. delta-of-delta压缩算法
2. bitmap压缩算法
3. delta&simple8b算法
##### 压缩算法比较
##### 自适应压缩算法
avgFre = dataPointsCount/SeriesCount/Duration
if avgFre>=0.7{
    bitmap()
}else if avgFre>=0.3{
    simple8b()
}else {
    gorilla()
}
##### 性能比较

### 索引优化
AVLTree->RBTree

按照理论来说，红黑是用非严格的平衡来换取增删节点时候旋转次数的降低，任何不平衡都会在三次旋转之内解决，而AVL是严格平衡树，因此在增加或者删除节点的时候，根据不同情况，旋转的次数比红黑树要多。 所以红黑树的插入效率更高，但是具体的实验结果却是二者的时间开销相差不大。



## 配置使用

1. 安装grpc
```shell
git clone https://github.com/grpc/grpc.git
cd grpc
git submodule update --init  //更新第三方源码
mkdir -p cmake/build
cd cmake/build
cmake ../..
make
sudo make install
```
2.安装protobuf
```shell
cd grpc/third_party/protobuf/
./autogen.sh
./configure
make
sudo make install
```

3.启动server

```shell
cd RealtimeDB
go build server.go
#server默认使用8086端口，更改端口请在server.go里面更改
./server
```

4.启动client_cpp

```shell
cd client_cpp
mkdir build
cd build
cmake ..
make
#这里的test_client只是一个demo，接口实现在client.cpp中
./test_client
```

### API调用

```
service Greeter {
    rpc WritePoints (WritePointsRequest) returns (WritePointsResponse) {}
    rpc QuerySeries (QuerySeriesRequest) returns (QuerySeriesResponse) {}
    rpc Config (ConfigRequest) returns (ConfigResponse) {}
    rpc QueryRange (QueryRangeRequest) returns (QueryRangeResponse) {}
    rpc QueryTagValues (QueryTagValuesRequest) returns (QueryTagValuesResponse) {}
    rpc QueryNewPoint (QueryNewPointRequest) returns (QueryNewPointResponse) {}
    rpc QuerySeriesAllData (QuerySeriesAllDataRequest) returns (QuerySeriesAllDataResponse) {}
    rpc QueryAllData (google.protobuf.Empty) returns (QueryAllDataResponse) {}
}
```
#### Config
每一台client试图连接server的时候，会先调用config函数来进行配置，每一个客户端配置一个数据库实例。
```
//在client中设定了localhost:8086
ClientConfig client(grpc::CreateChannel("localhost:8086",grpc::InsecureChannelCredentials()));
//向server发送请求，绑定“example”数据库实例
client.Config("example");
```
#### WritePoints
```c++
//绑定client配置存根
writePoints.stub_=client.stub_;
//写入一行数据
writePoints.WritePoints(row);
//row示例：
//{"metric":"cpu.busy","host":"host0","core":"core0","value":0,"timestamp":1600000000000}
//即{"metric":"metric_value_","tag1":"tag1_value_",...,"value":"value_","timestamp":"timestamp_value_"}
```
#### Query
```c++
ClientQuerySeries//查询曲线，查询包含某几个metric和tag键值对的所有曲线，如{"metric":"cpu.busy","host":"host2"}
ClientQueryNewPoint//查询某series的最新point，输入为一条曲线的metric，tags，如{"metric":"cpu.busy","host":"host2"}
ClientQueryRange//按照时间戳范围查询某series的数据，输入为一条曲线的metric，tags，start，end，如{"metric":"cpu.busy","host":"host1","start":1600000000001,"end":1600001111000}
ClientQueryTagValues//查询数据库中某个tag的所有value，如{"host"}
ClientQuerySeriesAllData//查询某一曲线的所有point，输入同ClientQueryNewPoint
ClientQueryAllData//load所有数据
```

以上查询支持正则，不支持聚合，示例如下：

```c++
{"metric":"cpu.busy","host":"host0","core":"core0","value":0,"timestamp":1600000000000}
{"metric":"cpu.busy","host":"host0","core":"core1","value":0,"timestamp":1600000000000}
在以上数据中，tag-host和tag-core不一样，因此是10条不同的series，以下是正则用法：
std::string tag=R"({"metric":"cpu.busy","host":"host."})";
qsd_.QuerySeriesAllData(tag);
//res:{"Tags":[{"Name":"core","Value":"core1"},{"Name":"host","Value":"host0"},{"Name":"metric","Value":"cpu.busy"}],"Points":[{"TimeStamp":1600000000000,"Value":0}]},{"Tags":[{"Name":"core","Value":"core0"},{"Name":"host","Value":"host0"},{"Name":"metric","Value":"cpu.busy"}],"Points":[{"TimeStamp":1600000000000,"Value":0}]}]
```

以下是一个demo：

```c++
//建立连接，发送配置请求   
ClientConfig client(grpc::CreateChannel("localhost:8086",grpc::InsecureChannelCredentials()));
//发送配置请求
client.Config("test");
//建立写请求，注意，建立了配置请求之后不能随意新建后续api中的stub_，只能使用配置请求的stub_，这是因为在config请求后server会确定唯一ip-database键值对
ClientWritePoints writePoints;
writePoints.stub_=std::move(client.stub_);
std::fstream f;
f.open("./data4.txt",std::ios::in);
  std::string row;
  while(getline(f,row))
  {
      //发送文件中的数据到server
      writePoints.WritePoints(row);
  }
f.close();
//建立查询请求
ClientQuerySeriesAllData qsd_;
qsd_.stub_=std::move(writePoints.stub_);
std::string tag=R"({"metric":"cpu.busy","host":"host2"})";
//发送查询请求
qsd_.QuerySeriesAllData(tag);
std::cout<<qsd_.qsde_.reply()<<std::endl;
```

如有不详细的地方，可以查看tests文件夹