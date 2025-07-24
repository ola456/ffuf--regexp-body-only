package filter

import (
	"testing"
)

func TestNewFilterByName(t *testing.T) {
	scf, _ := NewFilterByName("status", "200", false)
	if _, ok := scf.(*StatusFilter); !ok {
		t.Errorf("Was expecting statusfilter")
	}

	szf, _ := NewFilterByName("size", "200", false)
	if _, ok := szf.(*SizeFilter); !ok {
		t.Errorf("Was expecting sizefilter")
	}

	wf, _ := NewFilterByName("word", "200", false)
	if _, ok := wf.(*WordFilter); !ok {
		t.Errorf("Was expecting wordfilter")
	}

	lf, _ := NewFilterByName("line", "200", false)
	if _, ok := lf.(*LineFilter); !ok {
		t.Errorf("Was expecting linefilter")
	}

	ref, _ := NewFilterByName("regexp", "200", false)
	if _, ok := ref.(*RegexpFilter); !ok {
		t.Errorf("Was expecting regexpfilter")
	}

	tf, _ := NewFilterByName("time", "200", false)
	if _, ok := tf.(*TimeFilter); !ok {
		t.Errorf("Was expecting timefilter")
	}
}

func TestNewFilterByNameError(t *testing.T) {
	_, err := NewFilterByName("status", "invalid", false)
	if err == nil {
		t.Errorf("Was expecing an error")
	}
}

func TestNewFilterByNameNotFound(t *testing.T) {
	_, err := NewFilterByName("nonexistent", "invalid", false)
	if err == nil {
		t.Errorf("Was expecing an error with invalid filter name")
	}
}
