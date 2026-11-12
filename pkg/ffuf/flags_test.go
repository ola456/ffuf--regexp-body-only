package ffuf

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// TestRegisterFlags_Wellformed runs RegisterFlags against the real ConfigOptions
// struct so any malformed declaration — a slice field missing its `kind`, an
// unsupported field type, an unexported tagged field, or a duplicate flag name —
// fails here in CI instead of panicking at runtime. This is what buys back the
// compile-time safety that reflection-based registration gives up.
func TestRegisterFlags_Wellformed(t *testing.T) {
	fs := flag.NewFlagSet("ffuf", flag.ContinueOnError)
	opts := NewConfigOptions()

	var reg *FlagRegistry
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("RegisterFlags panicked (malformed tag or unsupported field type): %v", r)
			}
		}()
		reg = RegisterFlags(fs, opts)
	}()

	validSection := map[string]bool{
		SectionHTTP: true, SectionGeneral: true, SectionCompat: true, SectionMatcher: true,
		SectionFilter: true, SectionInput: true, SectionOutput: true,
	}

	// The full expected flag surface. Unlike a bare count, this catches a net-zero
	// add-and-remove, a rename, or a duplicate — update it DELIBERATELY when the CLI
	// surface changes. 81 visible flags + 7 hidden compat (4 aliases + 3 dummies).
	expected := map[string]bool{
		// HTTP
		"H": true, "X": true, "b": true, "cc": true, "ck": true, "d": true,
		"http2": true, "ignore-body": true, "r": true, "raw": true, "recursion": true,
		"recursion-depth": true, "recursion-strategy": true, "replay-proxy": true,
		"sni": true, "timeout": true, "u": true, "x": true,
		"preflight-mode": true, "preflight-error": true, "preflight-anyhost": true,
		"preflight": true, "preflight-var": true, "postflight": true, "postflight-var": true,
		"preflight-var-auto": true, "postflight-var-auto": true,
		// General
		"V": true, "ac": true, "acc": true, "ach": true, "ack": true, "acs": true,
		"c": true, "config": true, "json": true, "maxtime": true, "maxtime-job": true,
		"noninteractive": true, "p": true, "rate": true, "s": true, "sa": true,
		"scraperfile": true, "scrapers": true, "se": true, "search": true, "sf": true,
		"t": true, "v": true,
		// Matcher
		"mc": true, "ml": true, "mmode": true, "mr": true, "ms": true, "mt": true, "mw": true,
		// Filter
		"fc": true, "fl": true, "fmode": true, "fr": true, "fs": true, "ft": true, "fw": true, "mrb": true,
		// Input
		"D": true, "e": true, "enc": true, "ic": true, "input-cmd": true, "input-num": true,
		"input-shell": true, "mode": true, "request": true, "request-proto": true, "w": true,
		// Output
		"audit-log": true, "debug-log": true, "o": true, "od": true, "of": true, "or": true,
		// Compat aliases
		"cookie": true, "data": true, "data-ascii": true, "data-binary": true,
		// Compat dummies
		"compressed": true, "i": true, "k": true,
	}

	registered := map[string]bool{}
	fs.VisitAll(func(f *flag.Flag) {
		registered[f.Name] = true
		if f.Usage == "" {
			t.Errorf("flag -%s has an empty usage string", f.Name)
		}
		section, ok := reg.SectionOf(f.Name)
		if !ok {
			t.Errorf("flag -%s is registered but missing from the registry (no section)", f.Name)
			return
		}
		if !validSection[section] {
			t.Errorf("flag -%s has unknown section %q", f.Name, section)
		}
	})

	for name := range expected {
		if !registered[name] {
			t.Errorf("expected flag -%s is not registered (dropped or renamed?)", name)
		}
	}
	for name := range registered {
		if !expected[name] {
			t.Errorf("unexpected flag -%s registered — add it to the expected set deliberately", name)
		}
	}
}

func TestRegexpBodyOnlyFlagAndConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ffufrc")
	if err := os.WriteFile(path, []byte("[filter]\nregexp_body_only = true\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opts, err := ReadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.Filter.RegexpBodyOnly {
		t.Fatal("config file did not enable regexp body matching")
	}

	fs := flag.NewFlagSet("ffuf", flag.ContinueOnError)
	RegisterFlags(fs, opts)
	if err := fs.Parse([]string{"-mrb=false"}); err != nil {
		t.Fatal(err)
	}
	if opts.Filter.RegexpBodyOnly {
		t.Fatal("CLI flag did not override the config file")
	}
	if err := fs.Set("mrb", "true"); err != nil {
		t.Fatal(err)
	}
	opts.HTTP.URL = "https://example.org/FUZZ"
	opts.Input.Wordlists = []string{"/tmp/wl.txt"}
	conf, err := ConfigFromOptions(opts, context.Background(), func() {})
	if err != nil {
		t.Fatal(err)
	}
	if !conf.RegexpBodyOnly || !conf.Options.Filter.RegexpBodyOnly {
		t.Fatal("regexp body setting was lost in the parsed config")
	}
}
