package rtdb

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/snappy"
	"github.com/jwilder/encoding/simple8b"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"io"
)

// BytesCompressorType 代表字节数据压缩算法类型
type BytesCompressorType int8

const (
	// NoopBytesCompressor 不压缩
	NoopBytesCompressor BytesCompressorType = iota

	// ZstdBytesCompressor 使用 ZSTD 算法压缩
	ZstdBytesCompressor

	// SnappyBytesCompressor 使用 Snappy 算法压缩
	SnappyBytesCompressor
	// 使用simple8b算法压缩
	Simple8bBytesCompressor
	GzipBytesCompressor
	ZipBytesCompressor
)

// BytesCompressor 数据压缩器抽象接口
type BytesCompressor interface {
	Compress(src []byte) []byte
	Decompress(src []byte) ([]byte, error)
}

// ByteCompress 数据压缩
func ByteCompress(src []byte) []byte {
	return GlobalOpts.BytesCompressor.Compress(src)
}

// ByteDecompress 数据解压缩
func ByteDecompress(src []byte) ([]byte, error) {
	return GlobalOpts.BytesCompressor.Decompress(src)
}

type noopBytesCompressor struct{}

func NewNoopBytesCompressor() BytesCompressor {
	return &noopBytesCompressor{}
}

func (c *noopBytesCompressor) Compress(src []byte) []byte {
	return src
}

func (c *noopBytesCompressor) Decompress(src []byte) ([]byte, error) {
	return src, nil
}

type zstdBytesCompressor struct{}

func NewZstdBytesCompressor() BytesCompressor {
	return &zstdBytesCompressor{}
}

func (c *zstdBytesCompressor) Compress(src []byte) []byte {
	var encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
	return encoder.EncodeAll(src, make([]byte, 0, len(src)))
}

func (c *zstdBytesCompressor) Decompress(src []byte) ([]byte, error) {
	var decoder, _ = zstd.NewReader(nil)
	return decoder.DecodeAll(src, nil)
}

type snappyBytesCompressor struct{}

func NewSnappyBytesCompressor() BytesCompressor {
	return &snappyBytesCompressor{}
}

func (c *snappyBytesCompressor) Compress(src []byte) []byte {
	return snappy.Encode(nil, src)
}

func (c *snappyBytesCompressor) Decompress(src []byte) ([]byte, error) {
	return snappy.Decode(nil, src)
}

type simple8bBytesCompressor struct {}

func NewSimple8bBytesCompressor() BytesCompressor {
	return &simple8bBytesCompressor{}
}
func (c *simple8bBytesCompressor) Compress(src []byte) []byte {
	//fmt.Printf("the lens is %d\n", len(src))
	srcUint64:=make([]uint64,len(src))
	for i:=0;i<len(src);i++{
		srcUint64[i]=uint64(src[i])
	}
	all, err := simple8b.EncodeAll(srcUint64)
	if err != nil {
		return nil
	}
	var res = make([]byte, 0)
	for i:=0;i<len(all);i++{
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf,all[i])
		res=append(res,buf...)
	}
	return res
}

func (c *simple8bBytesCompressor) Decompress(src []byte) ([]byte, error) {
	srcUint64:=make([]uint64,len(src)/8)
	for i:=0;i<len(src)/8;i++{
		srcUint64[i]=binary.BigEndian.Uint64(src[i*8:(i+1)*8])
	}
	dst:=make([]uint64,0)
	_, err := simple8b.DecodeAll(dst,srcUint64)
	if err != nil {
		return nil, err
	}
	var res = make([]byte, len(dst))
	for i:=0;i<len(dst);i++{
		res[i]=byte(dst[i])
	}
	return res,nil
}
type gzipBytesCompressor struct {}

func NewGzipBytesCompressor() BytesCompressor {
	return &gzipBytesCompressor{}
}
func (c *gzipBytesCompressor) Compress(src []byte) []byte {
	var in bytes.Buffer
	g:=gzip.NewWriter(&in) //面向api编程调用压缩算法的一个api
	//参数就是指向某个数据缓冲区默认压缩等级是DefaultCompression 在这里还有另一个api可以调用调整压缩级别
	//gzip.NewWirterLevel(&in,gzip.BestCompression) NoCompression（对应的int 0）、
	//BestSpeed（1）、DefaultCompression（-1）、HuffmanOnly（-2）BestCompression（9）这几个级别也可以
	//这样写gzip.NewWirterLevel(&in,0)
	//这里的异常返回最好还是处理下，我这里纯属省事
	g.Write(src)
	g.Close()
	return in.Bytes()
}

func (c *gzipBytesCompressor) Decompress(src []byte) ([]byte, error) {
	var out bytes.Buffer
	var in bytes.Buffer
	in.Write(src)
	r,_:=gzip.NewReader(&in)
	r.Close() //这句放在后面也没有问题，不写也没有任何报错
	//机翻注释：关闭关闭读者。它不会关闭底层的io.Reader。为了验证GZIP校验和，读取器必须完全使用，直到io.EOF。

	io.Copy(&out,r)  //这里我看了下源码不是太明白，
	//我个人想法是这样的，Reader本身就是go中表示一个压缩文件的形式，r转化为[]byte就是一个符合压缩文件协议的压缩文件
	return out.Bytes(), nil
}
type zipBytesCompressor struct {}

func NewZipBytesCompressor() BytesCompressor {
	return &zipBytesCompressor{}
}
func (c *zipBytesCompressor) Compress(src []byte) []byte {
	var in bytes.Buffer
	z:=zlib.NewWriter(&in)
	z.Write(src)
	z.Close()
	return  in.Bytes()
}

func (c *zipBytesCompressor) Decompress(src []byte) ([]byte, error) {
	var out bytes.Buffer
	var in bytes.Buffer
	in.Write(src)
	r,_:=zlib.NewReader(&in)
	r.Close()
	io.Copy(&out,r)
	return out.Bytes(),nil
}