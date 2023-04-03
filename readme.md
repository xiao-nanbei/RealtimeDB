# DataBrainsDB 时序数据库设计与实现
##一.整体架构

##数据结构设计
```
series1: {"metric": "cpu.busy", "host": "localhost", "iface": "eth0"}
series2: {"metric": "cpu.busy", "host": "localhost", "iface": "eth1"}
```
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
为什么
##### 性能比较

### 索引优化
AVLTree->RBTree

按照理论来说，红黑是用非严格的平衡来换取增删节点时候旋转次数的降低，任何不平衡都会在三次旋转之内解决，而AVL是严格平衡树，因此在增加或者删除节点的时候，根据不同情况，旋转的次数比红黑树要多。 所以红黑树的插入效率更高，但是具体的实验结果却是二者的时间开销相差不大。



##配置使用
1. 安装grpc
```
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
```
cd grpc/third_party/protobuf/
./autogen.sh
./configure
make
sudo make install
```
3.