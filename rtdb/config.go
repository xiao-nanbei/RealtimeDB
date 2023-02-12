package rtdb



import (

"time"

)

var GlobalOpts = &RtdbOptions{
	MetaSerializer:    NewBinaryMetaSerializer(),
	BytesCompressor:   NewNoopBytesCompressor(),
	SegmentDuration:   2 * time.Hour,
	Retention:         7 * 24 * time.Hour, // 7d
	WriteTimeout:      30 * time.Second,
	OnlyMemoryMode:    false,
	EnableOutdated:    true,
	MaxRowsPerSegment: 19960412,
	DataPath:          ".",
	LoggerConfig:      nil,
}