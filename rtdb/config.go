package rtdb



import (

"time"

)

var GlobalOpts = &RtdbOptions{
	MetaSerializer:    NewBinaryMetaSerializer(),
	BytesCompressor:   NewNoopBytesCompressor(),
	SegmentDuration:   1600*time.Millisecond,
	Retention:         7 * 24 * time.Hour, // 7d
	WriteTimeout:      30 * time.Second,
	OnlyMemoryMode:    false,
	EnableOutdated:    true,
	MaxRowsPerSegment: 19960412,
	DataPath:          "./data",
	LoggerConfig:      nil,
	MemCompress:	   "gorilla",
}