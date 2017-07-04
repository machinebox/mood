package textboxtally

import (
	"testing"

	"github.com/machinebox/sdk-go/textbox"
	"github.com/matryer/is"
)

func TestAddKeywords(t *testing.T) {
	is := is.New(t)
	tally := New()
	tally.Add(&textbox.Analysis{
		Keywords: []textbox.Keyword{
			{Keyword: "foo"},
			{Keyword: "bar"},
			{Keyword: "baz"},
		},
	})
	tally.Add(&textbox.Analysis{
		Keywords: []textbox.Keyword{
			{Keyword: "foo"},
			{Keyword: "bar"},
		},
	})
	is.Equal(tally.Keywords["foo"], 2)
	is.Equal(tally.Keywords["bar"], 2)
	is.Equal(tally.Keywords["baz"], 1)
}
