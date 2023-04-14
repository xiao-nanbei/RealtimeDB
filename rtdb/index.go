package rtdb

import (
	"github.com/RoaringBitmap/roaring"
	"strings"
	"sync"
)

type memorySidSet struct {
	set map[string]struct{}
	mut sync.Mutex
}

func newMemorySidSet() *memorySidSet {
	return &memorySidSet{set: make(map[string]struct{})}
}

func (mss *memorySidSet) Add(a string) {
	mss.mut.Lock()
	defer mss.mut.Unlock()

	mss.set[a] = struct{}{}
}

func (mss *memorySidSet) Size() int {
	mss.mut.Lock()
	defer mss.mut.Unlock()

	return len(mss.set)
}

func (mss *memorySidSet) Copy() *memorySidSet {
	mss.mut.Lock()
	defer mss.mut.Unlock()

	newset := newMemorySidSet()
	for k := range mss.set {
		newset.set[k] = struct{}{}
	}

	return newset
}

func (mss *memorySidSet) Intersection(other *memorySidSet) {
	mss.mut.Lock()
	defer mss.mut.Unlock()

	for k := range mss.set {
		_, ok := other.set[k]
		if !ok {
			delete(mss.set, k)
		}
	}
}

func (mss *memorySidSet) Union(other *memorySidSet) {
	mss.mut.Lock()
	defer mss.mut.Unlock()

	for k := range other.set {
		mss.set[k] = struct{}{}
	}
}

func (mss *memorySidSet) List() []string {
	mss.mut.Lock()
	defer mss.mut.Unlock()

	keys := make([]string, 0, len(mss.set))
	for k := range mss.set {
		keys = append(keys, k)
	}

	return keys
}

type memoryIndexMap struct {
	idx map[string]*memorySidSet
	mut sync.Mutex
}

func newMemoryIndexMap() *memoryIndexMap {
	return &memoryIndexMap{idx: make(map[string]*memorySidSet)}
}

func (mim *memoryIndexMap) Range(f func(k string, v *memorySidSet)) {
	mim.mut.Lock()
	defer mim.mut.Unlock()

	for k, sids := range mim.idx {
		f(k, sids)
	}
}

func (mim *memoryIndexMap) UpdateIndex(sid string, tags TagSet) {
	mim.mut.Lock()
	defer mim.mut.Unlock()

	for _, tag := range tags {
		key := tag.MarshalName()
		if _, ok := mim.idx[key]; !ok {
			mim.idx[key] = newMemorySidSet()
		}
		mim.idx[key].Add(sid)
	}
}

func (mim *memoryIndexMap) MatchSids(tvs *TagValueSet, tms TagMatcherSet) []string {
	mim.mut.Lock()
	defer mim.mut.Unlock()

	sids := newMemorySidSet()
	var got bool
	for i := len(tms) - 1; i >= 0; i-- {
		tmp := newMemorySidSet()
		vs := tvs.Match(tms[i])
		for _, v := range vs {
			midx := mim.idx[JoinSeparator(tms[i].Name, v)]
			if midx == nil || midx.Size() <= 0 {
				continue
			}

			tmp.Union(midx.Copy())
		}

		if tmp == nil || tmp.Size() <= 0 {
			return nil
		}

		if !got {
			sids = tmp
			got = true
			continue
		}

		sids.Intersection(tmp.Copy())
	}

	return sids.List()
}

// Disk Index 负责管理磁盘的索引存储和搜索

type diskSidSet struct {
	set *roaring.Bitmap
	mut sync.Mutex
}

func newDiskSidSet() *diskSidSet {
	return &diskSidSet{set: roaring.New()}
}

func (dss *diskSidSet) Add(a uint32) {
	dss.mut.Lock()
	defer dss.mut.Unlock()

	dss.set.Add(a)
}

type DiskIndexMap struct {
	tag2sids   map[string]*diskSidSet
	tagOrdered map[int]string

	mut sync.Mutex
}

func NewDiskIndexMap(swt []SeriesWithTag) *DiskIndexMap {
	dim := &DiskIndexMap{
		tag2sids:   make(map[string]*diskSidSet),
		tagOrdered: make(map[int]string),
	}

	for i := range swt {
		row := swt[i]
		dim.tag2sids[row.Name] = newDiskSidSet()
		for _, sid := range swt[i].Sids {
			dim.tag2sids[row.Name].Add(sid)
		}
		dim.tagOrdered[i] = row.Name
	}

	return dim
}

func (dim *DiskIndexMap) MatchTags(lids ...uint32) []Tag {
	ret := make([]Tag, 0, len(lids))
	for _, lid := range lids {
		tagPair := dim.tagOrdered[int(lid)]
		kv := strings.SplitN(tagPair, separator, 2)
		if len(kv) != 2 {
			continue
		}

		ret = append(ret, Tag{
			Name:  kv[0],
			Value: kv[1],
		})
	}

	return ret
}

func (dim *DiskIndexMap) MatchSids(tvs *TagValueSet, lms TagMatcherSet) []uint32 {
	dim.mut.Lock()
	defer dim.mut.Unlock()

	lst := make([]*roaring.Bitmap, 0)
	for i := len(lms) - 1; i >= 0; i-- {
		tmp := make([]*roaring.Bitmap, 0)
		vs := tvs.Match(lms[i])

		for _, v := range vs {
			didx := dim.tag2sids[JoinSeparator(lms[i].Name, v)]
			if didx == nil || didx.set.IsEmpty() {
				continue
			}

			tmp = append(tmp, didx.set)
		}

		union := roaring.ParOr(4, tmp...)
		if union.IsEmpty() {
			return nil
		}

		lst = append(lst, union)
	}

	return roaring.ParAnd(4, lst...).ToArray()
}
