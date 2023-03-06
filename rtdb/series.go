package rtdb


import (
	"RealtimeDB/gorilla"
	"github.com/chenjiandongx/mandodb/pkg/sortedlist"

	"math"
	"sort"
	"sync"
	"sync/atomic"
)

type TszStore struct {
	block *gorilla.Series
	lock  sync.Mutex
	maxTs int64
	count int64
}

func (store *TszStore) Append(point *Point) *Point {
	store.lock.Lock()
	defer store.lock.Unlock()

	if store.maxTs >= point.TimeStamp {
		return point
	}
	store.maxTs = point.TimeStamp

	// 懒加载的方式初始化
	if store.count <= 0 {
		store.block = gorilla.New(uint64(point.TimeStamp))
	}

	store.block.Push(uint64(point.TimeStamp), point.Value)
	store.maxTs = point.TimeStamp

	store.count++
	return nil
}

func (store *TszStore) GetNewPoint() Point{

	t,val:=store.block.GetNewT()

	return Point{TimeStamp: int64(t),Value: val}
}
func (store *TszStore) Get(start, end int64) []Point {
	points := make([]Point, 0)

	it := store.block.Iter()
	for it.Next() {
		ts, val := it.Values()
		if ts > uint64(end) {
			break
		}

		if ts >= uint64(start) {
			points = append(points, Point{TimeStamp: int64(ts), Value: val})
		}
	}

	return points
}

func (store *TszStore) All() []Point {
	return store.Get(math.MinInt64, math.MaxInt64)
}

func (store *TszStore) Count() int {
	return int(atomic.LoadInt64(&store.count))
}

func (store *TszStore) Bytes() []byte {
	return store.block.Bytes()
}

func (store *TszStore) MergeOutdatedList(lst sortedlist.List) *TszStore {
	if lst == nil {
		return store
	}

	news := &TszStore{}
	tmp := store.All()
	it := lst.All()
	for it.Next() {
		dp := it.Value().(Point)
		tmp = append(tmp, Point{TimeStamp: dp.TimeStamp, Value: dp.Value})
	}

	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].TimeStamp < tmp[j].TimeStamp
	})

	for i := 0; i < len(tmp); i++ {
		news.Append(&tmp[i])
	}

	return news
}

type memorySeries struct {
	tags TagSet
	*TszStore
}

func newSeries(row *Row) *memorySeries {
	return &memorySeries{tags: row.Tags, TszStore: &TszStore{}}
}
