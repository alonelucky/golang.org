// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test tests are linked into the main binary and are run as part of
// the Docker build step.

package main

import (
	"fmt"
	stdlog "log"
	"os"
	"reflect"
	"strings"
	"time"
)

type compileTest struct {
	name               string // test name
	prog, want, errors string
	withVet            bool
	wantEvents         []Event
	wantVetErrors      string
}

func (s *server) test() {
	if err := s.healthCheck(); err != nil {
		stdlog.Fatal(err)
	}

	// Enable module downloads for testing:
	defer func(old string) { os.Setenv("ALLOW_PLAY_MODULE_DOWNLOADS", old) }(os.Getenv("ALLOW_PLAY_MODULE_DOWNLOADS"))
	os.Setenv("ALLOW_PLAY_MODULE_DOWNLOADS", "true")

	for i, t := range tests {
		fmt.Printf("testing case %d (%q)...\n", i, t.name)
		resp, err := compileAndRun(&request{Body: t.prog, WithVet: t.withVet})
		if err != nil {
			stdlog.Fatal(err)
		}
		if t.wantEvents != nil {
			if !reflect.DeepEqual(resp.Events, t.wantEvents) {
				stdlog.Fatalf("resp.Events = %q, want %q", resp.Events, t.wantEvents)
			}
			continue
		}
		if t.errors != "" {
			if resp.Errors != t.errors {
				stdlog.Fatalf("resp.Errors = %q, want %q", resp.Errors, t.errors)
			}
			continue
		}
		if resp.Errors != "" {
			stdlog.Fatal(resp.Errors)
		}
		if resp.VetErrors != t.wantVetErrors {
			stdlog.Fatalf("resp.VetErrs = %q, want %q", resp.VetErrors, t.wantVetErrors)
		}
		if t.withVet && (resp.VetErrors != "") == resp.VetOK {
			stdlog.Fatalf("resp.VetErrs & VetOK inconsistent; VetErrs = %q; VetOK = %v", resp.VetErrors, resp.VetOK)
		}
		if len(resp.Events) == 0 {
			stdlog.Fatalf("unexpected output: %q, want %q", "", t.want)
		}
		var b strings.Builder
		for _, e := range resp.Events {
			b.WriteString(e.Message)
		}
		if !strings.Contains(b.String(), t.want) {
			stdlog.Fatalf("unexpected output: %q, want %q", b.String(), t.want)
		}
	}
	fmt.Println("OK")
}

var tests = []compileTest{
	{
		name: "timezones_available",
		prog: `
package main

import "time"

func main() {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err.Error())
	}
	println(loc.String())
}
`, want: "America/New_York"},

	{
		name: "faketime_works",
		prog: `
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now())
}
`, want: "2009-11-10 23:00:00 +0000 UTC"},

	{
		name: "faketime_tickers",
		prog: `
package main

import (
	"fmt"
	"time"
)

func main() {
	t1 := time.Tick(time.Second * 3)
	t2 := time.Tick(time.Second * 7)
	t3 := time.Tick(time.Second * 11)
	end := time.After(time.Second * 19)
	want := "112131211"
	var got []byte
	for {
		var c byte
		select {
		case <-t1:
			c = '1'
		case <-t2:
			c = '2'
		case <-t3:
			c = '3'
		case <-end:
			if g := string(got); g != want {
				fmt.Printf("got %q, want %q\n", g, want)
			} else {
				fmt.Println("timers fired as expected")
			}
			return
		}
		got = append(got, c)
	}
}
`, want: "timers fired as expected"},

	{
		name: "old_tour_pkgs_in_gopath",
		prog: `
package main

import (
	"code.google.com/p/go-tour/pic"
	"code.google.com/p/go-tour/reader"
	"code.google.com/p/go-tour/tree"
	"code.google.com/p/go-tour/wc"
)

var (
	_ = pic.Show
	_ = reader.Validate
	_ = tree.New
	_ = wc.Test
)

func main() {
	println("ok")
}
`, want: "ok"},
	{
		name: "must_be_package_main",
		prog: `
package test

func main() {
	println("test")
}
`, want: "", errors: "package name must be main"},
	{
		name: "filesystem_contents",
		prog: `
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	filepath.Walk("/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
}
`, want: `/
/dev
/dev/null
/dev/random
/dev/urandom
/dev/zero
/etc
/etc/group
/etc/hosts
/etc/passwd
/etc/resolv.conf
/tmp
/usr
/usr/local
/usr/local/go
/usr/local/go/lib
/usr/local/go/lib/time
/usr/local/go/lib/time/zoneinfo.zip`},

	{
		name: "test_passes",
		prog: `
package main

import "testing"

func TestSanity(t *testing.T) {
	if 1+1 != 2 {
		t.Error("uhh...")
	}
}
`, want: `=== RUN   TestSanity
--- PASS: TestSanity (0.00s)
PASS`},

	{
		name: "test_without_import",
		prog: `
package main

func TestSanity(t *testing.T) {
	t.Error("uhh...")
}

func ExampleNotExecuted() {
	// Output: it should not run
}
`, want: "", errors: "./prog.go:4:20: undefined: testing\n"},

	{
		name: "test_with_import_ignored",
		prog: `
package main

import (
	"fmt"
	"testing"
)

func TestSanity(t *testing.T) {
	t.Error("uhh...")
}

func main() {
	fmt.Println("test")
}
`, want: "test"},

	{
		name: "example_runs",
		prog: `
package main//comment

import "fmt"

func ExampleOutput() {
	fmt.Println("The output")
	// Output: The output
}
`, want: `=== RUN   ExampleOutput
--- PASS: ExampleOutput (0.00s)
PASS`},

	{
		name: "example_unordered",
		prog: `
package main//comment

import "fmt"

func ExampleUnorderedOutput() {
	fmt.Println("2")
	fmt.Println("1")
	fmt.Println("3")
	// Unordered output: 3
	// 2
	// 1
}
`, want: `=== RUN   ExampleUnorderedOutput
--- PASS: ExampleUnorderedOutput (0.00s)
PASS`},

	{
		name: "example_fail",
		prog: `
package main

import "fmt"

func ExampleEmptyOutput() {
	// Output:
}

func ExampleEmptyOutputFail() {
	fmt.Println("1")
	// Output:
}
`, want: `=== RUN   ExampleEmptyOutput
--- PASS: ExampleEmptyOutput (0.00s)
=== RUN   ExampleEmptyOutputFail
--- FAIL: ExampleEmptyOutputFail (0.00s)
got:
1
want:

FAIL`},

	// Run program without executing this example function.
	{
		name: "example_no_output_skips_run",
		prog: `
package main

func ExampleNoOutput() {
	panic(1)
}
`, want: `testing: warning: no tests to run
PASS`},

	{
		name: "example_output",
		prog: `
package main

import "fmt"

func ExampleShouldNotRun() {
	fmt.Println("The output")
	// Output: The output
}

func main() {
	fmt.Println("Main")
}
`, want: "Main"},

	{
		name: "stdout_stderr_merge",
		prog: `
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "A")
	fmt.Fprintln(os.Stderr, "B")
	fmt.Fprintln(os.Stdout, "A")
	fmt.Fprintln(os.Stdout, "A")
}
`, want: "A\nB\nA\nA\n"},

	// Integration test for runtime.write fake timestamps.
	{
		name: "faketime_write_interaction",
		prog: `
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Fprintln(os.Stdout, "A")
	fmt.Fprintln(os.Stderr, "B")
	fmt.Fprintln(os.Stdout, "A")
	fmt.Fprintln(os.Stdout, "A")
	time.Sleep(time.Second)
	fmt.Fprintln(os.Stderr, "B")
	time.Sleep(time.Second)
	fmt.Fprintln(os.Stdout, "A")
}
`, wantEvents: []Event{
			{"A\n", "stdout", 0},
			{"B\n", "stderr", time.Nanosecond},
			{"A\nA\n", "stdout", time.Nanosecond},
			{"B\n", "stderr", time.Second - 2*time.Nanosecond},
			{"A\n", "stdout", time.Second},
		}},

	{
		name: "third_party_imports",
		prog: `
package main
import ("fmt"; "github.com/bradfitz/iter")
func main() { for i := range iter.N(5) { fmt.Println(i) } }
`, want: "0\n1\n2\n3\n4\n"},

	{
		name:          "compile_with_vet",
		withVet:       true,
		wantVetErrors: "./prog.go:5:2: Printf format %v reads arg #1, but call has 0 args\n",
		prog: `
package main
import "fmt"
func main() {
	fmt.Printf("hi %v")
}
`,
	},

	{
		name:    "compile_without_vet",
		withVet: false,
		prog: `
package main
import "fmt"
func main() {
	fmt.Printf("hi %v")
}
`,
	},

	{
		name:          "compile_modules_with_vet",
		withVet:       true,
		wantVetErrors: "./prog.go:6:2: Printf format %v reads arg #1, but call has 0 args\n",
		prog: `
package main
import ("fmt"; "github.com/bradfitz/iter")
func main() {
	for i := range iter.N(5) { fmt.Println(i) }
	fmt.Printf("hi %v")
}
`,
	},

	{
		name: "multi_file_basic",
		prog: `
package main
const foo = "bar"

-- two.go --
package main
func main() {
  println(foo)
}
`,
		wantEvents: []Event{
			{"bar\n", "stderr", 0},
		},
	},

	{
		name:    "multi_file_use_package",
		withVet: true,
		prog: `
package main

import "play.test/foo"

func main() {
    foo.Hello()
}

-- go.mod --
module play.test

-- foo/foo.go --
package foo

import "fmt"

func Hello() { fmt.Println("hello world") }
`,
	},
}
