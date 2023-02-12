package rtdb

import (
	"github.com/chenjiandongx/logger"
	"time"
)

type RtdbOptions struct {
	MetaSerializer    MetaSerializer
	BytesCompressor   BytesCompressor
	Retention         time.Duration
	SegmentDuration   time.Duration
	WriteTimeout      time.Duration
	OnlyMemoryMode    bool
	EnableOutdated    bool
	MaxRowsPerSegment int64
	DataPath          string
	LoggerConfig      *logger.Options
}

type Option func(c *RtdbOptions)

func WithMetaSerializerType(t MetaSerializerType) Option {
	return func(c *RtdbOptions) {
		switch t {
		default: // binary
			c.MetaSerializer = NewBinaryMetaSerializer()
		}
	}
}

// WithMetaBytesCompressorType 设置字节数据的压缩算法
// 目前提供了
// * 不压缩: NoopBytesCompressor（默认）
// * ZSTD: ZstdBytesCompressor
// * Snappy: SnappyBytesCompressor
func WithMetaBytesCompressorType(t BytesCompressorType) Option {
	return func(c *RtdbOptions) {
		switch t {
		case ZstdBytesCompressor:
			c.BytesCompressor = NewZstdBytesCompressor()
		case SnappyBytesCompressor:
			c.BytesCompressor = NewSnappyBytesCompressor()
		default: // noop
			c.BytesCompressor = NewNoopBytesCompressor()
		}
	}
}

// WithOnlyMemoryMode 设置是否默认只存储在内存中
// 默认为 false
func WithOnlyMemoryMode(memoryMode bool) Option {
	return func(c *RtdbOptions) {
		c.OnlyMemoryMode = memoryMode
	}
}

// WithEnabledOutdated 设置是否支持乱序写入 此特性会增加资源开销 但会提升数据完整性
// 默认为 true
func WithEnabledOutdated(outdated bool) Option {
	return func(c *RtdbOptions) {
		c.EnableOutdated = outdated
	}
}

// WithMaxRowsPerSegment 设置单 Segment 最大允许存储的点数
// 默认为 19960412（夹杂私货 🐶）
func WithMaxRowsPerSegment(n int64) Option {
	return func(c *RtdbOptions) {
		c.MaxRowsPerSegment = n
	}
}

// WithDataPath 设置 Segment 持久化存储文件夹
// 默认为 "."
func WithDataPath(d string) Option {
	return func(c *RtdbOptions) {
		c.DataPath = d
	}
}

// WithRetention 设置 Segment 持久化数据保存时长
// 默认为 7d
func WithRetention(t time.Duration) Option {
	return func(c *RtdbOptions) {
		c.Retention = t
	}
}

// WithWriteTimeout 设置写入超时阈值
// 默认为 30s
func WithWriteTimeout(t time.Duration) Option {
	return func(c *RtdbOptions) {
		c.WriteTimeout = t
	}
}

// WithLoggerConfig 设置日志配置项
func WithLoggerConfig(opt *logger.Options) Option {
	return func(c *RtdbOptions) {
		if opt != nil {
			c.LoggerConfig = opt
			logger.SetOptions(*opt)
		}
	}
}
