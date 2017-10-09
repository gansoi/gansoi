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
