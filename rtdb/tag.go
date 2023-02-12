package rtdb

import (
	"bytes"
	"github.com/cespare/xxhash"
	"regexp"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"
)
// TagMatcher Tag 匹配器 支持正则
type TagMatcher struct {
	Name   string
	Value  string
	IsRegx bool
}
var TagBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

func UnmarshalTagName(s string) (string, string) {
	pair := strings.SplitN(s, separator, 2)
	if len(pair) != 2 {
		return "", ""
	}

	return pair[0], pair[1]
}
func (t Tag) MarshalName() string {
	return JoinSeparator(t.Name, t.Value)
}

type TagValueSet struct {
	mut    sync.Mutex
	values map[string]map[string]struct{}
}
func NewTagValueSet() *TagValueSet {
	return &TagValueSet{
		values: map[string]map[string]struct{}{},
	}
}
func (tvs *TagValueSet) Set(tag, value string) {
	tvs.mut.Lock()
	defer tvs.mut.Unlock()

	if _, ok := tvs.values[tag]; !ok {
		tvs.values[tag] = make(map[string]struct{})
	}

	tvs.values[tag][value] = struct{}{}
}
func (tvs *TagValueSet) Get(tag string) []string {
	tvs.mut.Lock()
	defer tvs.mut.Unlock()

	ret := make([]string, 0)
	vs, ok := tvs.values[tag]
	if !ok {
		return ret
	}

	for k := range vs {
		ret = append(ret, k)
	}

	return ret
}

type FastRegexMatcher struct {
	re       *regexp.Regexp
	prefix   string
	suffix   string
	contains string
}

func NewFastRegexMatcher(v string) (*FastRegexMatcher, error) {
	re, err := regexp.Compile("^(?:" + v + ")$")
	if err != nil {
		return nil, err
	}

	parsed, err := syntax.Parse(v, syntax.Perl)
	if err != nil {
		return nil, err
	}

	m := &FastRegexMatcher{
		re: re,
	}

	if parsed.Op == syntax.OpConcat {
		m.prefix, m.suffix, m.contains = OptimizeConcatRegex(parsed)
	}

	return m, nil
}

// optimizeConcatRegex returns literal prefix/suffix text that can be safely
// checked against the label value before running the regexp matcher.
func OptimizeConcatRegex(r *syntax.Regexp) (prefix, suffix, contains string) {
	sub := r.Sub

	// We can safely remove begin and end text matchers respectively
	// at the beginning and end of the regexp.
	if len(sub) > 0 && sub[0].Op == syntax.OpBeginText {
		sub = sub[1:]
	}
	if len(sub) > 0 && sub[len(sub)-1].Op == syntax.OpEndText {
		sub = sub[:len(sub)-1]
	}

	if len(sub) == 0 {
		return
	}

	// Given Prometheus regex matchers are always anchored to the begin/end
	// of the text, if the first/last operations are literats, we can safely
	// treat them as prefix/suffix.
	if sub[0].Op == syntax.OpLiteral && (sub[0].Flags&syntax.FoldCase) == 0 {
		prefix = string(sub[0].Rune)
	}
	if last := len(sub) - 1; sub[last].Op == syntax.OpLiteral && (sub[last].Flags&syntax.FoldCase) == 0 {
		suffix = string(sub[last].Rune)
	}

	// If contains any literal which is not a prefix/suffix, we keep the
	// 1st one. We do not keep the whole list of literats to simplify the
	// fast path.
	for i := 1; i < len(sub)-1; i++ {
		if sub[i].Op == syntax.OpLiteral && (sub[i].Flags&syntax.FoldCase) == 0 {
			contains = string(sub[i].Rune)
			break
		}
	}

	return
}

func (m *FastRegexMatcher) MatchString(s string) bool {
	if m.prefix != "" && !strings.HasPrefix(s, m.prefix) {
		return false
	}

	if m.suffix != "" && !strings.HasSuffix(s, m.suffix) {
		return false
	}

	if m.contains != "" && !strings.Contains(s, m.contains) {
		return false
	}
	return m.re.MatchString(s)
}

// Match 主要用于匹配 Tags 组合 支持正则匹配
func (tvs *TagValueSet) Match(matcher TagMatcher) []string {
	ret := make([]string, 0)
	if matcher.IsRegx {
		pattern, err := NewFastRegexMatcher(matcher.Value)
		if err != nil {
			return []string{matcher.Value}
		}

		for _, v := range tvs.Get(matcher.Name) {
			if pattern.MatchString(v) {
				ret = append(ret, v)
			}
		}

		return ret
	}

	return []string{matcher.Value}
}

// filter 过滤空 kv 和重复数据
func (ts TagSet) Filter() TagSet {
	mark := make(map[string]struct{})
	var size int
	for _, v := range ts {
		_, ok := mark[v.Name]
		if v.Name != "" && v.Value != "" && !ok {
			ts[size] = v // 复用原来的 slice
			size++
		}
		mark[v.Name] = struct{}{}
	}

	return ts[:size]
}
// Map 将 Tag 列表转换成 map
func (ts TagSet) Map() map[string]string {
	m := make(map[string]string)
	for _, tag := range ts {
		m[tag.Name] = tag.Value
	}

	return m
}
func (ts TagSet) Len() int           { return len(ts) }
func (ts TagSet) Less(i, j int) bool { return ts[i].Name < ts[j].Name }
func (ts TagSet) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

func (ts TagSet) AddMetricName(metric string) TagSet {
	tags := ts.Filter()
	tags = append(tags, Tag{
		Name:  metricName,
		Value: metric,
	})
	return tags
}
func (ts TagSet) Sorted() {
	sort.Sort(ts)
}
// Hash 哈希计算 TagSet 唯一标识符
func (ts TagSet) Hash() uint64 {
	b := TagBufPool.Get().([]byte)

	const sep = '\xff'
	for _, v := range ts {
		b = append(b, v.Name...)
		b = append(b, sep)
		b = append(b, v.Value...)
		b = append(b, sep)
	}
	h := xxhash.Sum64(b)

	b = b[:0]
	TagBufPool.Put(b) // 复用 buffer

	return h
}

func (ts TagSet) Has(name string) bool {
	for _, tag := range ts {
		if tag.Name == name {
			return true
		}
	}

	return false
}

func (ts TagSet) String() string {
	var b bytes.Buffer

	b.WriteByte('{')
	for i, t := range ts {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.WriteString(t.Name)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(t.Value))
	}
	b.WriteByte('}')
	return b.String()
}
func (tms TagMatcherSet) AddMetricName(metric string) TagMatcherSet {
	tags := tms.Filter()

	newt := TagMatcher{
		Name:  metricName,
		Value: metric,
	}
	tags = append(tags, newt)
	return tags
}
type TagMatcherSet []TagMatcher

// filter 过滤空 kv 和重复数据
func (tms TagMatcherSet) Filter() TagMatcherSet {
	mark := make(map[string]struct{})
	var size int
	for _, v := range tms {
		_, ok := mark[v.Name]
		if v.Name != "" && v.Value != "" && !ok {
			tms[size] = v // 复用原来的 slice
			size++
		}
		mark[v.Name] = struct{}{}
	}

	return tms[:size]
}
