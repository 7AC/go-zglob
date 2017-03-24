package zglob

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func check(got []string, expected []string) bool {
	sort.Strings(got)
	sort.Strings(expected)
	return reflect.DeepEqual(expected, got)
}

type testZGlob struct {
	pattern  string
	expected []string
	err      error
}

var testGlobs = []testZGlob{
	{`fo*`, []string{`foo`}, nil},
	{`foo`, []string{`foo`}, nil},
	{`foo/*`, []string{`foo/bar`, `foo/baz`}, nil},
	{`foo/**`, []string{`foo/bar`, `foo/baz`}, nil},
	{`f*o/**`, []string{`foo/bar`, `foo/baz`}, nil},
	{`*oo/**`, []string{`foo/bar`, `foo/baz`, `hoo/bar`}, nil},
	{`*oo/b*`, []string{`foo/bar`, `foo/baz`, `hoo/bar`}, nil},
	{`*oo/*z`, []string{`foo/baz`}, nil},
	{`foo/**/*`, []string{`foo/bar`, `foo/bar/baz`, `foo/bar/baz.txt`, `foo/bar/baz/noo.txt`, `foo/baz`}, nil},
	{`*oo/**/*`, []string{`foo/bar`, `foo/bar/baz`, `foo/bar/baz.txt`, `foo/bar/baz/noo.txt`, `foo/baz`, `hoo/bar`}, nil},
	{`*oo/*.txt`, []string{}, nil},
	{`*oo/*/*.txt`, []string{`foo/bar/baz.txt`}, nil},
	{`*oo/**/*.txt`, []string{`foo/bar/baz.txt`, `foo/bar/baz/noo.txt`}, nil},
	{`doo`, nil, os.ErrNotExist},
	{`./f*`, []string{`foo`}, nil},
}

func setup(t *testing.T) string {
	tmpdir, err := ioutil.TempDir("", "zglob")
	if err != nil {
		t.Fatal(err)
	}

	os.MkdirAll(filepath.Join(tmpdir, "foo/baz"), 0755)
	os.MkdirAll(filepath.Join(tmpdir, "foo/bar"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, "foo/bar/baz.txt"), []byte{}, 0644)
	os.MkdirAll(filepath.Join(tmpdir, "foo/bar/baz"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, "foo/bar/baz/noo.txt"), []byte{}, 0644)
	os.MkdirAll(filepath.Join(tmpdir, "hoo/bar"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, "foo/bar/baz.txt"), []byte{}, 0644)

	return tmpdir
}

func TestGlob(t *testing.T) {
	tmpdir := setup(t)
	defer os.RemoveAll(tmpdir)

	curdir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(curdir)

	tmpdir = "."
	for _, test := range testGlobs {
		expected := make([]string, len(test.expected))
		for i, e := range test.expected {
			expected[i] = e
		}
		got, err := Glob(test.pattern)
		if err != nil {
			if test.err != err {
				t.Error(err)
			}
			continue
		}
		if !check(expected, got) {
			t.Errorf(`zglob failed: pattern %q(%q): expected %v but got %v`, test.pattern, tmpdir, expected, got)
		}
	}
}

func TestGlobAbs(t *testing.T) {
	tmpdir := setup(t)
	defer os.RemoveAll(tmpdir)

	curdir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(curdir)

	for _, test := range testGlobs {
		pattern := filepath.ToSlash(filepath.Join(tmpdir, test.pattern))
		expected := make([]string, len(test.expected))
		for i, e := range test.expected {
			expected[i] = filepath.ToSlash(filepath.Join(tmpdir, e))
		}
		got, err := Glob(pattern)
		if err != nil {
			if test.err != err {
				t.Error(err)
			}
			continue
		}
		if !check(expected, got) {
			t.Errorf(`zglob failed: pattern %q(%q): expected %v but got %v`, pattern, tmpdir, expected, got)
		}
	}
}

func TestMatch(t *testing.T) {
	for _, test := range testGlobs {
		for _, f := range test.expected {
			got, err := Match(test.pattern, f)
			if err != nil {
				t.Error(err)
				continue
			}
			if !got {
				t.Errorf("%q should match with %q", f, test.pattern)
			}
		}
	}
}

type MatchTest struct {
	pattern, s string
	match      bool
	err        error
}

var matchTests = []MatchTest{
	{"abc", "abc", true, nil},
	{"*", "abc", true, nil},
	{"*c", "abc", true, nil},
	{"a*", "a", true, nil},
	{"a*", "abc", true, nil},
	{"a*", "ab/c", false, nil},
	{"a*/b", "abc/b", true, nil},
	{"a*/b", "a/c/b", false, nil},
	{"a*b*c*d*e*/f", "axbxcxdxe/f", true, nil},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/f", true, nil},
	{"a*b*c*d*e*/f", "axbxcxdxe/xxx/f", false, nil},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/fff", false, nil},
	{"a*b?c*x", "abxbbxdbxebxczzx", true, nil},
	{"a*b?c*x", "abxbbxdbxebxczzy", false, nil},
	{"ab[c]", "abc", true, nil},
	{"ab[b-d]", "abc", true, nil},
	{"ab[e-g]", "abc", false, nil},
	{"ab[^c]", "abc", false, nil},
	{"ab[^b-d]", "abc", false, nil},
	{"ab[^e-g]", "abc", true, nil},
	{"a\\*b", "a*b", true, nil},
	{"a\\*b", "ab", false, nil},
	{"a?b", "a☺b", true, nil},
	{"a[^a]b", "a☺b", true, nil},
	{"a???b", "a☺b", false, nil},
	{"a[^a][^a][^a]b", "a☺b", false, nil},
	{"[a-ζ]*", "α", true, nil},
	{"*[a-ζ]", "A", false, nil},
	{"a?b", "a/b", false, nil},
	{"a*b", "a/b", false, nil},
	{"[\\]a]", "]", true, nil},
	{"[\\-]", "-", true, nil},
	{"[x\\-]", "x", true, nil},
	{"[x\\-]", "-", true, nil},
	{"[x\\-]", "z", false, nil},
	{"[\\-x]", "x", true, nil},
	{"[\\-x]", "-", true, nil},
	{"[\\-x]", "a", false, nil},
	{"[]a]", "]", false, filepath.ErrBadPattern},
	{"[-]", "-", false, filepath.ErrBadPattern},
	{"[x-]", "x", false, filepath.ErrBadPattern},
	{"[x-]", "-", false, filepath.ErrBadPattern},
	{"[x-]", "z", false, filepath.ErrBadPattern},
	{"[-x]", "x", false, filepath.ErrBadPattern},
	{"[-x]", "-", false, filepath.ErrBadPattern},
	{"[-x]", "a", false, filepath.ErrBadPattern},
	{"\\", "a", false, filepath.ErrBadPattern},
	{"[a-b-c]", "a", false, filepath.ErrBadPattern},
	{"[", "a", false, filepath.ErrBadPattern},
	{"[^", "a", false, filepath.ErrBadPattern},
	{"[^bc", "a", false, filepath.ErrBadPattern},
	{"a[", "a", false, nil},
	{"a[", "ab", false, filepath.ErrBadPattern},
	{"*x", "xxx", true, nil},
}

func errp(e error) string {
	if e == nil {
		return "<nil>"
	}
	return e.Error()
}

func TestMatchStdlib(t *testing.T) {
	for _, tt := range matchTests {
		pattern := tt.pattern
		s := tt.s
		ok, err := Match(pattern, s)
		if ok != tt.match || err != tt.err {
			t.Fatalf("Match(%#q, %#q) = %v, %q want %v, %q", pattern, s, ok, errp(err), tt.match, errp(tt.err))
		}
	}
}
