package filter

import (
	"strings"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestNewRegexpFilter(t *testing.T) {
	f, _ := NewRegexpFilter("s([a-z]+)arch", false)
	statusRepr := f.Repr()
	if !strings.Contains(statusRepr, "s([a-z]+)arch") {
		t.Errorf("Status filter was expected to have a regexp value")
	}
}

func TestNewRegexpFilterError(t *testing.T) {
	_, err := NewRegexpFilter("r((", false)
	if err == nil {
		t.Errorf("Was expecting an error from errenous input data")
	}
}

func TestRegexpFiltering(t *testing.T) {
	f, _ := NewRegexpFilter("s([a-z]+)arch", false)
	for i, test := range []struct {
		input  string
		output bool
	}{
		{"search", true},
		{"text and search", true},
		{"sbarch in beginning", true},
		{"midd scarch le", true},
		{"s1arch", false},
		{"invalid", false},
	} {
		inp := make(map[string][]byte)
		resp := ffuf.Response{
			Data: []byte(test.input),
			Request: &ffuf.Request{
				Input: inp,
			},
		}
		filterReturn, _ := f.Filter(&resp)
		if filterReturn != test.output {
			t.Errorf("Filter test %d: Was expecing filter return value of %t but got %t", i, test.output, filterReturn)
		}
	}
}
