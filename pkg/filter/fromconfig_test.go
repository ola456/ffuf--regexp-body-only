package filter

import (
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// TestFromConfig_DefaultStatusMatcher covers the one behavior main used to drive
// via flag.Visit: whether the default status matcher is installed. It is now
// testable without the flag package.
func TestFromConfig_DefaultStatusMatcher(t *testing.T) {
	opts := ffuf.NewConfigOptions()

	mm, err := FromConfig(opts, true)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if _, ok := mm.GetMatchers()["status"]; !ok {
		t.Error("default status matcher should be installed when addDefaultStatusMatcher=true")
	}

	mm2, err := FromConfig(opts, false)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if _, ok := mm2.GetMatchers()["status"]; ok {
		t.Error("no status matcher should be installed when addDefaultStatusMatcher=false and none configured")
	}
}

// TestFromConfig_FiltersAndMatchers checks that configured filters/matchers are
// installed from the ConfigOptions values.
func TestFromConfig_FiltersAndMatchers(t *testing.T) {
	opts := ffuf.NewConfigOptions()
	opts.Filter.Size = "100"
	opts.Matcher.Regexp = "admin"

	mm, err := FromConfig(opts, false)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if _, ok := mm.GetFilters()["size"]; !ok {
		t.Error("size filter should be installed from opts.Filter.Size")
	}
	if _, ok := mm.GetMatchers()["regexp"]; !ok {
		t.Error("regexp matcher should be installed from opts.Matcher.Regexp")
	}
}

func TestFromConfig_RegexpBodyOnly(t *testing.T) {
	for _, bodyOnly := range []bool{false, true} {
		opts := ffuf.NewConfigOptions()
		opts.Filter.Regexp = "secret"
		opts.Matcher.Regexp = "secret"
		opts.Filter.RegexpBodyOnly = bodyOnly

		mm, err := FromConfig(opts, false)
		if err != nil {
			t.Fatalf("FromConfig: %v", err)
		}
		responses := []struct {
			name string
			resp ffuf.Response
			want bool
		}{
			{"header", ffuf.Response{Headers: map[string][]string{"X-Probe": {"secret"}}, Data: []byte("public"), Request: &ffuf.Request{}}, !bodyOnly},
			{"body", ffuf.Response{Headers: map[string][]string{"X-Probe": {"public"}}, Data: []byte("secret"), Request: &ffuf.Request{}}, true},
		}
		for _, tc := range responses {
			for name, provider := range map[string]ffuf.FilterProvider{
				"filter":  mm.GetFilters()["regexp"],
				"matcher": mm.GetMatchers()["regexp"],
			} {
				got, err := provider.Filter(&tc.resp)
				if err != nil || got != tc.want {
					t.Errorf("bodyOnly=%t %s %s: got %t, %v; want %t", bodyOnly, name, tc.name, got, err, tc.want)
				}
			}
		}

		// Interactive and per-domain filters added later use the same setting.
		if err := mm.AddFilter("regexp", "secret", true); err != nil {
			t.Fatal(err)
		}
		interactiveFilter := mm.GetFilters()["regexp"]
		mm.RemoveFilter("regexp")
		if err := mm.AddPerDomainFilter("example.org", "regexp", "secret"); err != nil {
			t.Fatal(err)
		}
		for name, provider := range map[string]ffuf.FilterProvider{
			"interactive": interactiveFilter,
			"per-domain":  mm.FiltersForDomain("example.org")["regexp"],
		} {
			got, err := provider.Filter(&responses[0].resp)
			if err != nil || got != !bodyOnly {
				t.Errorf("bodyOnly=%t %s header: got %t, %v; want %t", bodyOnly, name, got, err, !bodyOnly)
			}
		}
	}
}
