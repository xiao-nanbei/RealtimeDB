package rtdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/chenjiandongx/logger"
	"github.com/chenjiandongx/mandodb/pkg/mmap"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type Point struct {
	TimeStamp int64
	Value float64
}

type Tag struct {
	Name string
	Value string
}

type TagSet []Tag

type Row struct {
	Metric string
	Tags TagSet
	Point Point
}

const (
	separator    = "/-/"
	defaultQSize = 128
)

func JoinSeparator(a, b interface{}) string {
	return fmt.Sprintf("%v%s%v", a, separator, b)
}

func Dirname(a, b int64) string {
	return path.Join(GlobalOpts.DataPath, fmt.Sprintf("seg-%d-%d", a, b))
}

// ID 使用 hash 计算 Series 的唯一标识
func (r Row) ID() string {
	return JoinSeparator(xxhash.Sum64([]byte(r.Metric)), r.Tags.Hash())
}

type RTDB struct {
	segs *SegmentList
	mut sync.Mutex
	ctx context.Context
	cancel context.CancelFunc
	q chan []*Row
	wg sync.WaitGroup
}
var timerPool sync.Pool

func GetTimer(d time.Duration) *time.Timer{
	if v := timerPool.Get(); v != nil {
		t := v.(*time.Timer)
		/*
			if t.Reset(d) {
				panic("active timer trapped to the pool")
			}*/
		return t
	}
	return time.NewTimer(d)
}

func PutTimer(t *time.Timer) {
	if !t.Stop() {
		// Drain t.C if it wasn't obtained by the caller yet.
		select {
		case <-t.C:
		default:
		}
	}
	timerPool.Put(t)
}

func (rtdb *RTDB) InsertRows(rows []*Row) error {
	timer := GetTimer(GlobalOpts.WriteTimeout)
	select {
	case rtdb.q <- rows:
		PutTimer(timer)
	case <-timer.C:
		PutTimer(timer)
		return errors.New("failed to insert rows to database, write overloaded")
	}
	return nil
}

func (rtdb *RTDB) ingestRows(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rs := <-rtdb.q:
			head, err := rtdb.GetHeadPartition()
			if err != nil {
				logger.Errorf("failed to get head partition: %v", head)
				continue
			}
			head.InsertRows(rs)
		case <-time.After(1*time.Second):
			log.Println(time.Now())
		}
	}
}

func (rtdb *RTDB) GetHeadPartition() (Segment, error) {
	rtdb.mut.Lock()
	defer rtdb.mut.Unlock()

	if rtdb.segs.Head.Frozen() {
		head := rtdb.segs.Head

		go func() {
			rtdb.wg.Add(1)
			defer rtdb.wg.Done()

			rtdb.segs.Add(head)

			t0 := time.Now()
			dn := Dirname(head.MinTs(), head.MaxTs())

			if err := WriteToDisk(head.(*MemorySegment)); err != nil {
				logger.Errorf("failed to flush data to disk, %v", err)
				return
			}

			fname := path.Join(dn, "data")
			mf, err := mmap.OpenMmapFile(fname)
			if err != nil {
				logger.Errorf("failed to make a mmap file %s, %v", fname, err)
				return
			}

			rtdb.segs.Replace(head, newDiskSegment(mf, dn, head.MinTs(), head.MaxTs()))
			logger.Infof("write file %s take: %v", fname, time.Since(t0))
		}()

		rtdb.segs.Head = NewMemorySegment()
	}
	return rtdb.segs.Head, nil
}

type MetricRet struct {
	Tags TagSet
	Points []Point
}

/*
func (rtdb *RTDB) LoadAllDataToFiles(){
	rtdb.segs.Mut.Lock()
	defer rtdb.segs.Mut.Unlock()
	segs := make([]Segment, 0)
	iter := rtdb.segs.Lst.All()
	for iter.Next() {
		if iter.Value()==nil{
			break
		}
		seg := iter.Value().(Segment)
		segs = append(segs, seg)

	}
	segs = append(segs, rtdb.segs.Head)
	for _,seg:=range segs{
		seg = seg.Load()
		seg.QuerySeries()
	}
}
*/

func (rtdb *RTDB) QuerySeries(tms TagMatcherSet, start, end int64) ([]map[string]string, error) {
	tmp := make([]TagSet, 0)
	for _, segment := range rtdb.segs.Get(start, end) {
		segment = segment.Load()
		data, err := segment.QuerySeries(tms)
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, data...)
	}
	return rtdb.mergeQuerySeriesResult(tmp...), nil
}
func (rtdb *RTDB) QueryRange(metric string, tms TagMatcherSet, start, end int64) ([]MetricRet, error) {
	tms = tms.AddMetricName(metric)

	tmp := make([]MetricRet, 0)
	for _, segment := range rtdb.segs.Get(start, end) {
		segment = segment.Load()
		data, err := segment.QueryRange(tms, start, end)
		if err != nil {
			return nil, err
		}

		tmp = append(tmp, data...)
	}

	return rtdb.mergeQueryRangeResult(tmp...), nil
}
func (rtdb *RTDB) QueryNewPoint(metric string, tms TagMatcherSet) ([]Point, error){
	tms = tms.AddMetricName(metric)
	segment:=rtdb.segs.Head
	segment = segment.Load()
	data, err := segment.GetNewPoint(tms)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (rtdb *RTDB) mergeQueryRangeResult(ret ...MetricRet) []MetricRet {
	metrics := make(map[uint64]*MetricRet)
	for _, r := range ret {
		h := r.Tags.Hash()
		v, ok := metrics[h]
		if !ok {
			metrics[h] = &MetricRet{
				Tags: r.Tags,
				Points: r.Points,
			}
			continue
		}

		v.Points = append(v.Points, r.Points...)
	}

	items := make([]MetricRet, 0, len(metrics))
	for _, v := range metrics {
		sort.Slice(v.Points, func(i, j int) bool {
			return v.Points[i].TimeStamp < v.Points[j].TimeStamp
		})

		items = append(items, *v)
	}

	return items
}



func (rtdb *RTDB) mergeQuerySeriesResult(ret ...TagSet) []map[string]string {
	lbs := make(map[uint64]TagSet)
	for _, r := range ret {
		lbs[r.Hash()] = r
	}

	items := make([]map[string]string, 0)
	for _, lb := range lbs {
		items = append(items, lb.Map())
	}

	return items
}

func (rtdb *RTDB) QueryTagValues(tag string, start, end int64) []string {
	tmp := make(map[string]struct{})

	for _, segment := range rtdb.segs.Get(start, end) {
		segment = segment.Load()
		values := segment.QueryTagValues(tag)
		for i := 0; i < len(values); i++ {
			tmp[values[i]] = struct{}{}
		}
	}

	ret := make([]string, 0, len(tmp))
	for k := range tmp {
		ret = append(ret, k)
	}

	sort.Strings(ret)

	return ret
}

func (rtdb *RTDB) Close() {
	rtdb.wg.Wait()
	rtdb.cancel()

	it := rtdb.segs.Lst.All()
	for it.Next() {
		if it.Value()==nil{
			break
		}
		it.Value().(Segment).Close()
	}

	rtdb.segs.Head.Close()
}

func (rtdb *RTDB) removeExpires() {

	tick := time.Tick(5 * time.Minute)
	for {
		select {
		case <-rtdb.ctx.Done():
			return
		case <-tick:
			now := time.Now().UnixMilli()

			var removed []Segment
			it := rtdb.segs.Lst.All()
			for it.Next() {
				if it.Value()==nil{
					break
				}
				if now-it.Value().(Segment).MaxTs() > int64(GlobalOpts.Retention.Milliseconds()) {
					removed = append(removed, it.Value().(Segment))
				}
			}

			for _, r := range removed {
				rtdb.segs.Remove(r)
			}
		}
	}
}

func (rtdb *RTDB) loadFiles() {
	mkdir(GlobalOpts.DataPath)
	err := filepath.Walk(GlobalOpts.DataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read the dir: %s, err: %v", path, err)
		}

		if !info.IsDir() || !strings.HasPrefix(info.Name(), "seg-") {
			return nil
		}

		files, err := ioutil.ReadDir(filepath.Join(GlobalOpts.DataPath, info.Name()))
		if err != nil {
			return fmt.Errorf("failed to load data storage, err: %v", err)
		}

		diskseg := &diskSegment{}

		for _, file := range files {
			fn := filepath.Join(GlobalOpts.DataPath, info.Name(), file.Name())

			if file.Name() == "data" {
				mf, err := mmap.OpenMmapFile(fn)
				if err != nil {
					return fmt.Errorf("failed to open mmap file %s, err: %v", fn, err)
				}

				diskseg.dataFd = mf
				diskseg.dataFilename = fn
				diskseg.tagVs = NewTagValueSet()
			}

			if file.Name() == "meta.json" {
				bs, err := ioutil.ReadFile(fn)
				if err != nil {
					return fmt.Errorf("failed to read file: %s, err: %v", fn, err)
				}

				desc := Desc{}
				if err := json.Unmarshal(bs, &desc); err != nil {
					return fmt.Errorf("failed to unmarshal desc file: %v", err)
				}

				diskseg.minTs = desc.MinT
				diskseg.maxTs = desc.MaxT
			}
		}

		rtdb.segs.Add(diskseg)
		return nil
	})

	if err != nil {
		logger.Error(err)
	}
}

func OpenRTDB(opts ...Option) *RTDB {
	for _, opt := range opts {
		opt(GlobalOpts)
	}

	rtdb := &RTDB{
		segs: NewSegmentList(),
		q:    make(chan []*Row, defaultQSize),
	}

	rtdb.loadFiles()

	worker := runtime.GOMAXPROCS(-1)
	rtdb.ctx, rtdb.cancel = context.WithCancel(context.Background())

	for i := 0; i < worker; i++ {
		go rtdb.ingestRows(rtdb.ctx)
	}
	go rtdb.removeExpires()

	return rtdb
}
