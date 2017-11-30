package cluster

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
)

var (
	testPath = path.Join(os.TempDir(), fmt.Sprintf(".info-%d.json", rand.Int63()))
)

func newInfo() *Info {
	return NewInfo(testPath)
}

func rmInfo() {
	os.Remove(testPath)
}

func TestInfoSave(t *testing.T) {
	info := newInfo()
	defer rmInfo()

	err := info.Save()
	if err != nil {
		t.Fatalf("Save() failed")
	}

	st, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Could not stat saved file")
	}

	if st == nil {
		t.Fatalf("Save() failed to write a file")
	}

	permissions := st.Mode().Perm()
	if permissions&077 != 0 {
		t.Fatalf("File mode is too permissive, got %#o", st.Mode())
	}
}

func TestInfoSaveFail(t *testing.T) {
	i := NewInfo("/nonexisting/path/something.json")
	err := i.Save()
	if err == nil {
		t.Fatalf("Save() failed to detect error")
	}
}

func TestInfoLoadFail(t *testing.T) {
	defer rmInfo()

	i := &Info{
		path: testPath,
	}

	err := i.Load()
	if err == nil {
		t.Fatalf("Load() failed to detect file read error")
	}

	// Write some broken JSON.
	ioutil.WriteFile(testPath, []byte("NOT JSON"), 0600)
	err = i.Load()
	if err == nil {
		t.Fatalf("Load() failed to detect broken JSON")
	}
}

func TestSelf(t *testing.T) {
	const self = "name-for-testing"
	info := newInfo()
	defer rmInfo()

	info.SetSelf(self)
	if info.Self() != self {
		t.Errorf("Wrong name returned from Self(), got %s", info.Self())
	}
}

func TestIP(t *testing.T) {
	info := newInfo()
	defer rmInfo()

	cases := []struct {
		in  string
		out string
	}{
		{"go-test-localhost-v6.gansoi-dev.com", "::1"},
		{"go-test-localhost-v4.gansoi-dev.com", "127.0.0.1"},
		{"go-test-nonexisting.gansoi-dev.com", ""},
	}

	for _, c := range cases {
		info.SetSelf(c.in)
		ip := info.IP()

		if ip == nil && c.out == "" {
			continue
		}

		if ip == nil && c.out != "" {
			t.Errorf("IP() returned wrong IP for '%s', expected '%s', got nil", c.in, c.out)
			continue
		}

		if ip.String() != c.out {
			t.Errorf("IP('%s') returned '%s', expected '%s'", c.in, ip.String(), c.out)
		}
	}
}

func TestDefaultPort(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"", ":4934"},
		{"hello", "hello:4934"},
		{"hello:12333", "hello:12333"},
	}

	for _, c := range cases {
		result := DefaultPort(c.in)
		if result != c.out {
			t.Errorf("DefaultPort('%s') returned '%s', was expecting '%s'", c.in, result, c.out)
		}
	}
}
