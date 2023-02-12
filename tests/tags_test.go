package tests
import (
	"RealtimeDB/rtdb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTags_String(t *testing.T) {
	cases := []struct {
		tags   rtdb.TagSet
		expected string
	}{
		{
			tags: rtdb.TagSet{
				{
					Name:  "t1",
					Value: "t1",
				},
				{
					Name:  "t2",
					Value: "t2",
				},
			},
			expected: "{t1=\"t1\", t2=\"t2\"}",
		},
		{
			tags:   rtdb.TagSet{},
			expected: "{}",
		},
		{
			tags:   nil,
			expected: "{}",
		},
	}
	for _, c := range cases {
		str := c.tags.String()
		require.Equal(t, c.expected, str)
	}
}

func TestTags_Has(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "foo",
			expected: false,
		},
		{
			input:    "aaa",
			expected: true,
		},
	}

	labelsSet := rtdb.TagSet{
		{
			Name:  "aaa",
			Value: "111",
		},
		{
			Name:  "bbb",
			Value: "222",
		},
	}

	for i, test := range tests {
		got := labelsSet.Has(test.input)
		require.Equal(t, test.expected, got, "unexpected comparison result for test case %d", i)
	}
}

func TestTags_Hash(t *testing.T) {
	lbls := rtdb.TagSet{
		{Name: "foo", Value: "bar"},
		{Name: "baz", Value: "qux"},
	}
	require.Equal(t, lbls.Hash(), lbls.Hash())
	require.NotEqual(t, lbls.Hash(), rtdb.TagSet{lbls[1], lbls[0]}.Hash(), "unordered labels match.")
	require.NotEqual(t, lbls.Hash(), rtdb.TagSet{lbls[0]}.Hash(), "different labels match.")
}


func TestTags_WithoutEmpty(t *testing.T) {
	for _, test := range []struct {
		input    rtdb.TagSet
		expected rtdb.TagSet
	}{
		{
			input: rtdb.TagSet{
				{Name: "foo"},
				{Name: "bar"},
			},
			expected: rtdb.TagSet{},
		},
		{
			input: rtdb.TagSet{
				{Name: "foo"},
				{Name: "bar"},
				{Name: "baz"},
			},
			expected: rtdb.TagSet{},
		},
		{
			input: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "hostname", Value: "localhost"},
				{Name: "job", Value: "check"},
			},
			expected: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "hostname", Value: "localhost"},
				{Name: "job", Value: "check"},
			},
		},
		{
			input: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "hostname", Value: "localhost"},
				{Name: "bar"},
				{Name: "job", Value: "check"},
			},
			expected: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "hostname", Value: "localhost"},
				{Name: "job", Value: "check"},
			},
		},
		{
			input: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "foo"},
				{Name: "hostname", Value: "localhost"},
				{Name: "bar"},
				{Name: "job", Value: "check"},
			},
			expected: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "hostname", Value: "localhost"},
				{Name: "job", Value: "check"},
			},
		},
		{
			input: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "foo"},
				{Name: "baz"},
				{Name: "hostname", Value: "localhost"},
				{Name: "bar"},
				{Name: "job", Value: "check"},
			},
			expected: rtdb.TagSet{
				{Name: "__name__", Value: "test"},
				{Name: "hostname", Value: "localhost"},
				{Name: "job", Value: "check"},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			require.Equal(t, test.expected, test.input.Filter())
		})
	}
}