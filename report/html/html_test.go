package html

import (
	"testing"
)

func reportFailure(t *testing.T, got, expected, fmt string, args ...interface{}) {
	t.Fail()
	t.Logf(fmt, args...)
	t.Logf("expected: [%v]", expected)
	t.Logf("     got: [%v]", got)
}
func testPathToRoot(t *testing.T, path, expected string) {
	got := pathToRoot(path)
	if got != expected {
		reportFailure(t, got, expected, `pathToRoot("%s")`, path)
	}
}
func TestPathToRoot(t *testing.T) {
	testPathToRoot(t, "/a/b/file.txt", "../../")
	testPathToRoot(t, "a/b/file.txt", "../../")
	testPathToRoot(t, "file.txt", "")
}

var relativePathTests = []struct {
	to, from string
	want     string
}{
	{"a.txt", ".", "a.txt"},
	{"a.txt", "a.txt", ""},
	{"//a.txt", "/a.txt", ""},
	{"a.txt", "b.txt", "a.txt"},
	{"a/a.txt", "b/b.txt", "../a/a.txt"},
	{"a/a.txt", ".", "a/a.txt"},
	{"files/a/elephant/foo.html", "files/a/b/c/file.html", "../../elephant/foo.html"},
	{"files//home/foo/baz.html", "files/Obj.constructor.html", "home/foo/baz.html"},
}

func TestGetRelativePathTo(t *testing.T) {
	for i, tt := range relativePathTests {
		got := getRelativePathTo(tt.to, tt.from)
		if got != tt.want {
			t.Errorf("%d. getRelativePathTo(%q, %q)\n Got %q\nwant %q", i, tt.to, tt.from, got, tt.want)
		}

	}
}

func testStripCommonPath(t *testing.T, path1, path2, e1, e2 string) {
	got1, got2 := stripCommonPath(path1, path2)
	if got1 != e1 || got2 != e2 {
		t.Fail()
		t.Logf(`stripCommonPath("%s", "%s")`, path1, path2)
		t.Logf("expected: [%s %s]", e1, e2)
		t.Logf("     got: [%s %s]", got1, got2)

	}
}
func TestStripCommonPath(t *testing.T) {
	testStripCommonPath(t, "a.txt", "b.txt", "a.txt", "b.txt")
	testStripCommonPath(t, "a/a.txt", "b/b.txt", "a/a.txt", "b/b.txt")
	testStripCommonPath(t, "a/a.txt", "a/b.txt", "a.txt", "b.txt")
}
