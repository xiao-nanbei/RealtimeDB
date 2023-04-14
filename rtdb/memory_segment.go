package rtdb

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"sync"
	"sync/atomic"

	"RealtimeDB/sortedlist"
)

type MemorySegment struct {
	Once     sync.Once
	Segment  sync.Map
	IndexMap *memoryIndexMap
	TagVs  *TagValueSet

	Outdated    map[string]sortedlist.List
	OutdatedMut sync.Mutex

	MinT int64
	MaxT int64

	SeriesCount     int64
	DataPointsCount int64
}

func NewMemorySegment() Segment {
	return &MemorySegment{
		IndexMap: newMemoryIndexMap(),
		TagVs:  NewTagValueSet(),
		Outdated: make(map[string]sortedlist.List),
		MinT:    math.MaxInt64,
		MaxT:    math.MinInt64,
	}
}

func (ms *MemorySegment) getOrCreateSeries(row *Row) *memorySeries {
	v, ok := ms.Segment.Load(row.ID())
	if ok {
		return v.(*memorySeries)
	}

	atomic.AddInt64(&ms.SeriesCount, 1)
	newSeries := newSeries(row)
	ms.Segment.Store(row.ID(), newSeries)

	return newSeries
}

func (ms *MemorySegment) MinTs() int64 {
	return atomic.LoadInt64(&ms.MinT)
}

func (ms *MemorySegment) MaxTs() int64 {
	return atomic.LoadInt64(&ms.MaxT)
}

func (ms *MemorySegment) Frozen() bool {
	if GlobalOpts.OnlyMemoryMode {
		return false
	}
	return ms.MaxTs()-ms.MinTs() >= GlobalOpts.SegmentDuration.Milliseconds()-1
}

func (ms *MemorySegment) Type() SegmentType {
	return MemorySegmentType
}

func (ms *MemorySegment) Close() error {
	if ms.DataPointsCount == 0 || GlobalOpts.OnlyMemoryMode {
		return nil
	}
	return WriteToDisk(ms)
}

func (ms *MemorySegment) Cleanup() error {
	return nil
}

func (ms *MemorySegment) Load() Segment {
	return ms
}

func (ms *MemorySegment) InsertRows(rows []*Row) {
	for _, row := range rows {
		ms.TagVs.Set(metricName, row.Metric)
		for _, tag := range row.Tags {
			ms.TagVs.Set(tag.Name, tag.Value)
		}

		row.Tags = row.Tags.AddMetricName(row.Metric)
		row.Tags.Sorted()
		series := ms.getOrCreateSeries(row)
		dp := series.Append(&row.Point)
		if dp != nil {
			ms.OutdatedMut.Lock()
			if _, ok := ms.Outdated[row.ID()]; !ok {
				ms.Outdated[row.ID()] = sortedlist.NewTree()
			}
			ms.Outdated[row.ID()].Add(row.Point.TimeStamp, row.Point)
			ms.OutdatedMut.Unlock()
		}

		if atomic.LoadInt64(&ms.MinT) >= row.Point.TimeStamp {
			atomic.StoreInt64(&ms.MinT, row.Point.TimeStamp)
		}
		if atomic.LoadInt64(&ms.MaxT) <= row.Point.TimeStamp {
			atomic.StoreInt64(&ms.MaxT, row.Point.TimeStamp)
		}
		atomic.AddInt64(&ms.DataPointsCount, 1)
		ms.IndexMap.UpdateIndex(row.ID(), row.Tags)
	}
}

func (ms *MemorySegment) QueryTagValues(tag string) []string {
	return ms.TagVs.Get(tag)
}

func (ms *MemorySegment) QuerySeries(tms TagMatcherSet) ([]TagSet, error) {

	matchSids := ms.IndexMap.MatchSids(ms.TagVs, tms)
	ret := make([]TagSet, 0)
	for _, sid := range matchSids {
		b, _ := ms.Segment.Load(sid)
		series := b.(*memorySeries)

		ret = append(ret, series.tags)
	}

	return ret, nil
}

func (ms *MemorySegment) GetNewPoint(tms TagMatcherSet) (Point, error){
	matchSids := ms.IndexMap.MatchSids(ms.TagVs, tms)
	point:=Point{TimeStamp: -1,Value: -1}
	for _, sid := range matchSids {
		b, _ := ms.Segment.Load(sid)
		series := b.(*memorySeries)

		point = series.GetNewPoint()
	}
	return point,nil
}
func (ms *MemorySegment) QueryRange(tms TagMatcherSet, start, end int64) ([]MetricRet, error) {
	matchSids := ms.IndexMap.MatchSids(ms.TagVs, tms)
	ret := make([]MetricRet, 0, len(matchSids))
	for _, sid := range matchSids {
		b, _ := ms.Segment.Load(sid)
		series := b.(*memorySeries)

		points := series.Get(start, end)

		ms.OutdatedMut.Lock()
		v, ok := ms.Outdated[sid]
		if ok {
			iter := v.Range(start, end)
			for iter.Next() {
				points = append(points, iter.Value().(Point))
			}
		}
		ms.OutdatedMut.Unlock()

		ret = append(ret, MetricRet{
			Tags: series.tags,
			Points: points,
		})
	}

	return ret, nil
}

func (ms *MemorySegment) Marshal() ([]byte, []byte, error) {
	sids := make(map[string]uint32)

	startOffset := 0
	size := 0

	dataBuf := make([]byte, 0)

	// TOC 占位符 用于后面标记 dataBytes / metaBytes 长度
	dataBuf = append(dataBuf, make([]byte, uint64Size*2)...)
	meta := Metadata{MinTs: ms.MinT, MaxTs: ms.MaxT}

	// key: sid
	// value: series entity
	ms.Segment.Range(func(key, value interface{}) bool {
		sid := key.(string)
		sids[sid] = uint32(size)
		size++

		series := value.(*memorySeries)
		meta.SidRelatedTags = append(meta.SidRelatedTags, series.tags)

		ms.OutdatedMut.Lock()
		v, ok := ms.Outdated[sid]
		ms.OutdatedMut.Unlock()

		var dataBytes []byte
		if ok {
			dataBytes = ByteCompress(series.MergeOutdatedList(v).Bytes())
		} else {
			dataBytes = ByteCompress(series.Bytes())
		}

		dataBuf = append(dataBuf, dataBytes...)
		endOffset := startOffset + len(dataBytes)
		meta.Series = append(meta.Series, MetaSeries{
			Sid:         key.(string),
			StartOffset: uint64(startOffset),
			EndOffset:   uint64(endOffset),
		})
		startOffset = endOffset

		return true
	})

	tagIdx := make([]SeriesWithTag, 0)

	// key: Tag.MarshalName()
	// value: sids...
	ms.IndexMap.Range(func(key string, value *memorySidSet) {
		l := make([]uint32, 0)
		for _, s := range value.List() {
			l = append(l, sids[s])
		}

		sort.Slice(l, func(i, j int) bool {
			return l[i] < l[j]
		})
		tagIdx = append(tagIdx, SeriesWithTag{Name: key, Sids: l})
	})
	meta.Tags = tagIdx

	metaBytes, err := MarshalMeta(meta)
	if err != nil {
		return nil, nil, err
	}
	metalen := len(metaBytes)

	desc := &Desc{
		SeriesCount:     ms.SeriesCount,
		DataPointsCount: ms.DataPointsCount,
		MaxT:           ms.MaxT,
		MinT:           ms.MinT,
	}

	descBytes, _ := json.MarshalIndent(desc, "", "    ")

	dataLen := len(dataBuf) - (uint64Size * 2)
	dataBuf = append(dataBuf, metaBytes...)

	// TOC 写入
	encf := newEncbuf()
	encf.MarshalUint64(uint64(dataLen))
	dataLenBs := encf.Bytes()
	copy(dataBuf[:uint64Size], dataLenBs[:uint64Size])

	encf.Reset()

	encf.MarshalUint64(uint64(metalen))
	metaLenBs := encf.Bytes()
	copy(dataBuf[uint64Size:uint64Size*2], metaLenBs[:uint64Size])

	return dataBuf, descBytes, nil
}

func mkdir(d string) {
	if _, err := os.Stat(d); !os.IsNotExist(err) {
		return
	}

	if err := os.MkdirAll(d, os.ModePerm); err != nil {

			panic(fmt.Sprintf("BUG: failed to create dir: %s", d))


	}
}

func WriteToDisk(segment *MemorySegment) error {
	dataBytes, descBytes, err := segment.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal segment: %s", err.Error())
	}

	writeFile := func(f string, data []byte) error {
		if isFileExist(f) {
			return fmt.Errorf("%s file is already exists", f)
		}

		fd, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer fd.Close()

		_, err = fd.Write(data)
		return err
	}

	dn := Dirname(segment.MinTs(), segment.MaxTs())
	mkdir(dn)

	if err := writeFile(path.Join(dn, "data"), dataBytes); err != nil {
		return err
	}

	// 这里的 meta.json 只是描述了一些简单的信息 并非全局定义的 MetaData
	if err := writeFile(path.Join(dn, "meta.json"), descBytes); err != nil {
		return err
	}
	log.Println("write to disk")
	return nil
}