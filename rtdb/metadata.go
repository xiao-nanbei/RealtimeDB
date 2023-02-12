package rtdb
import (
	"sort"
)

type MetaSeries struct {
	Sid         string
	StartOffset uint64
	EndOffset   uint64
	Tags      []uint32
}

type SeriesWithTag struct {
	Name string
	Sids []uint32
}

// Metadata 描述了 Segment 的相关元数据
type Metadata struct {
	MinTs  int64
	MaxTs  int64
	Series []MetaSeries
	Tags []SeriesWithTag // labels -> sid

	SidRelatedTags []TagSet
}

type MetaSerializerType int8

type MetaSerializer interface {
	Marshal(Metadata) ([]byte, error)
	Unmarshal([]byte, *Metadata) error
}

// MarshalMeta 负责序列化 Meta 数据
func MarshalMeta(meta Metadata) ([]byte, error) {
	return GlobalOpts.MetaSerializer.Marshal(meta)
}

// UnmarshalMeta 负责反序列化 Meta 数据
func UnmarshalMeta(data []byte, meta *Metadata) error {
	return GlobalOpts.MetaSerializer.Unmarshal(data, meta)
}

const (
	endOfBlock uint16 = 0xffff
	uint16Size        = 2
	uint32Size        = 4
	uint64Size        = 8

	magic = "https://github.com/chenjiandongx/mandodb"
)

type binaryMetaSerializer struct{}

func NewBinaryMetaSerializer() MetaSerializer {
	return &binaryMetaSerializer{}
}

func (s *binaryMetaSerializer) Marshal(meta Metadata) ([]byte, error) {
	encf := newEncbuf()

	// labels block
	labelOrdered := make(map[string]int)
	for idx, row := range meta.Tags {
		labelOrdered[row.Name] = idx
		encf.MarshalUint16(uint16(len(row.Name)))
		encf.MarshalString(row.Name)
		encf.MarshalUint32(uint32(len(row.Sids)))
		encf.MarshalUint32(row.Sids...)
	}
	encf.MarshalUint16(endOfBlock)

	// series block
	for idx, series := range meta.Series {
		encf.MarshalUint16(uint16(len(series.Sid)))
		encf.MarshalString(series.Sid)
		encf.MarshalUint64(series.StartOffset, series.EndOffset)

		rl := meta.SidRelatedTags[idx]
		encf.MarshalUint32(uint32(rl.Len()))

		lids := make([]uint32, 0, rl.Len())
		for _, lb := range rl {
			lids = append(lids, uint32(labelOrdered[lb.MarshalName()]))
		}

		sort.Slice(lids, func(i, j int) bool {
			return lids[i] < lids[j]
		})
		encf.MarshalUint32(lids...)
	}
	encf.MarshalUint16(endOfBlock)

	encf.MarshalUint64(uint64(meta.MinTs))
	encf.MarshalUint64(uint64(meta.MaxTs))
	encf.MarshalString(magic)

	return ByteCompress(encf.Bytes()), nil
}

func (s *binaryMetaSerializer) Unmarshal(data []byte, meta *Metadata) error {
	data, err := ByteDecompress(data)
	if err != nil {
		return ErrInvalidSize
	}

	if len(data) < len(magic) {
		return ErrInvalidSize
	}

	decf := newDecbuf()
	// 检验数据完整性
	if decf.UnmarshalString(data[len(data)-len(magic):]) != magic {
		return ErrInvalidSize
	}

	offset := 0
	labels := make([]SeriesWithTag, 0)
	for {
		var labelName string

		labelLen := decf.UnmarshalUint16(data[offset : offset+uint16Size])
		offset += uint16Size

		if labelLen == endOfBlock {
			break
		}

		labelName = decf.UnmarshalString(data[offset : offset+int(labelLen)])
		offset += int(labelLen)

		sidCnt := decf.UnmarshalUint32(data[offset : offset+uint32Size])
		offset += uint32Size

		sidLst := make([]uint32, sidCnt)
		for i := 0; i < int(sidCnt); i++ {
			sidLst[i] = decf.UnmarshalUint32(data[offset : offset+uint32Size])
			offset += uint32Size
		}
		labels = append(labels, SeriesWithTag{Name: labelName, Sids: sidLst})
	}
	meta.Tags = labels

	rows := make([]MetaSeries, 0)
	for {
		series := MetaSeries{}

		sidLen := decf.UnmarshalUint16(data[offset : offset+uint16Size])
		offset += uint16Size

		if sidLen == endOfBlock {
			break
		}

		series.Sid = decf.UnmarshalString(data[offset : offset+int(sidLen)])
		offset += int(sidLen)

		series.StartOffset = decf.UnmarshalUint64(data[offset : offset+uint64Size])
		offset += uint64Size

		series.EndOffset = decf.UnmarshalUint64(data[offset : offset+uint64Size])
		offset += uint64Size

		labelCnt := decf.UnmarshalUint32(data[offset : offset+uint32Size])
		offset += uint32Size

		labelLst := make([]uint32, labelCnt)
		for i := 0; i < int(labelCnt); i++ {
			labelLst[i] = decf.UnmarshalUint32(data[offset : offset+uint32Size])
			offset += uint32Size
		}
		series.Tags = labelLst
		rows = append(rows, series)
	}
	meta.Series = rows

	meta.MinTs = int64(decf.UnmarshalUint64(data[offset : offset+uint64Size]))
	offset += uint64Size

	meta.MaxTs = int64(decf.UnmarshalUint64(data[offset : offset+uint64Size]))
	offset += uint64Size

	return decf.Err()
}