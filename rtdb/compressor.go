package rtdb

import (
	"encoding/binary"
	"github.com/golang/snappy"
	"github.com/jwilder/encoding/simple8b"
	"github.com/klauspost/compress/zstd"
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
	srcUint64:=make([]uint64,0)
	for i:=0;i<len(src)/8;i++{
		srcUint64=append(srcUint64,binary.LittleEndian.Uint64(src[i*8:(i+1)*8]))
	}
	all, err := simple8b.EncodeAll(srcUint64)
	if err != nil {
		return nil
	}
	var res = make([]byte, 0)
	for i:=0;i<len(all);i++{
		var buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf,all[i])
		res=append(res,buf...)
	}
	return res
}

func (c *simple8bBytesCompressor) Decompress(src []byte) ([]byte, error) {
	srcUint64:=make([]uint64,0)
	for i:=0;i<len(src)/8;i++{
		srcUint64=append(srcUint64,binary.LittleEndian.Uint64(src[i*8:(i+1)*8]))
	}
	dst:=make([]uint64,0)
	_, err := simple8b.DecodeAll(dst,srcUint64)
	if err != nil {
		return nil, err
	}
	var res = make([]byte, 0)
	for i:=0;i<len(dst);i++{
		var buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf,dst[i])
		res=append(res,buf...)
	}
	return res,nil
}