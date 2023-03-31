package rtdb

import (
	"os"
	"sync"

	"github.com/chenjiandongx/mandodb/pkg/sortedlist"
)

type SegmentType string

const (
	DiskSegmentType   SegmentType = "DISK"
	MemorySegmentType             = "MEMORY"
)

type Segment interface {
	InsertRows(row []*Row)
	GetNewPoint(tms TagMatcherSet) ([]Point, error)
	QueryRange(tms TagMatcherSet, start, end int64) ([]MetricRet, error)
	QuerySeries(tms TagMatcherSet) ([]TagSet, error)
	QueryTagValues(tag string) []string
	MinTs() int64
	MaxTs() int64
	Frozen() bool
	Close() error
	Cleanup() error
	Type() SegmentType
	Load() Segment
}

type Desc struct {
	SeriesCount     int64 `json:"seriesCount"`
	DataPointsCount int64 `json:"dataPointsCount"`
	MaxT            int64 `json:"maxRt"`
	MinT            int64 `json:"minRt"`
}

type SegmentList struct {
	Mut  sync.Mutex
	Head Segment
	Lst  sortedlist.List
}

func NewSegmentList() *SegmentList {
	return &SegmentList{Head: NewMemorySegment(), Lst: sortedlist.NewTree()}
}

func (sl *SegmentList) Get(start, end int64) []Segment {
	sl.Mut.Lock()
	defer sl.Mut.Unlock()

	segs := make([]Segment, 0)

	iter := sl.Lst.All()

	for iter.Next() {
		if iter.Value()==nil{
			break
		}
		seg := iter.Value().(Segment)
		if sl.Choose(seg, start, end) {
			segs = append(segs, seg)
		}
	}

	// 头部永远是最新的 所以放最后
	if sl.Choose(sl.Head, start, end) {
		segs = append(segs, sl.Head)
	}

	return segs
}

func (sl *SegmentList) Choose(seg Segment, start, end int64) bool {
	if seg.MinTs() < start && seg.MaxTs() > start {
		return true
	}

	if seg.MinTs() > start && seg.MaxTs() < end {
		return true
	}

	if seg.MinTs() < end && seg.MaxTs() > end {
		return true
	}

	return false
}

func (sl *SegmentList) Add(segment Segment) {
	sl.Mut.Lock()
	defer sl.Mut.Unlock()

	sl.Lst.Add(segment.MinTs(), segment)
}

func (sl *SegmentList) Remove(segment Segment) error {
	sl.Mut.Lock()
	defer sl.Mut.Unlock()

	if err := segment.Close(); err != nil {
		return err
	}

	if err := segment.Cleanup(); err != nil {
		return err
	}

	sl.Lst.Remove(segment.MinTs())
	return nil
}

func (sl *SegmentList) Replace(pre, nxt Segment) error {
	sl.Mut.Lock()
	defer sl.Mut.Unlock()

	if err := pre.Close(); err != nil {
		return err
	}

	if err := pre.Cleanup(); err != nil {
		return err
	}

	sl.Lst.Add(pre.MinTs(), nxt)
	return nil
}

const metricName = "metric"

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}