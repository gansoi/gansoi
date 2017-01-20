package cluster

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

const (
	testPath = "/dev/shm/info.json"
)

func newInfo() *Info {
	return NewInfo(testPath)
}

func readInfo() []byte {
	b, _ := ioutil.ReadFile(testPath)

	return b
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

func TestInfoPeers(t *testing.T) {
	i := newInfo()
	defer rmInfo()

	cases := [][]string{
		nil,
		{},
		{"a"},
		{"a", "b"},
	}

	for _, c := range cases {
		err := i.SetPeers(c)
		if err != nil {
			t.Fatalf("SetPeers() failed: %s", err.Error())
		}

		// Let's have a roundtrip to disk.
		i.Save()
		i.PeerList = nil
		i.Load()

		peers, err := i.Peers()
		if err != nil {
			t.Fatalf("Peers() failed: %s", err.Error())
		}

		if !reflect.DeepEqual(c, peers) {
			t.Fatalf("Peers() did not return what we gave SetPeers()")
		}
	}
}

func TestInfoSelf(t *testing.T) {
	i := newInfo()
	defer rmInfo()

	cases := []string{
		"a",
		"",
		"\n",
	}

	for _, c := range cases {
		err := i.SetSelf(c)
		if err != nil {
			t.Fatalf("SetSelf() failed: %s", err.Error())
		}

		// Let's have a roundtrip to disk.
		i.Save()
		i.SelfName = "NOT THIS"
		i.Load()

		self := i.Self()
		if err != nil {
			t.Fatalf("Self() failed: %s", err.Error())
		}

		if c != self {
			t.Fatalf("Self() failed to return what we gave SetSelf()")
		}
	}
}

func TestInfoIP(t *testing.T) {
	i := newInfo()
	defer rmInfo()

	i.SetSelf("go-test-target-v4.gansoi-dev.com")
	ip := i.IP()
	if ip == nil {
		t.Fatalf("Failed to resolve IPv4 , got %s", ip)
	}

	if ip.String() != "198.51.100.1" {
		t.Fatalf("Failed to resolve self to IP, got %s", ip.String())
	}

	i.SetSelf("go-test-target-v6.gansoi-dev.com")
	ip = i.IP()
	if ip == nil {
		t.Fatalf("Failed to resolve IPv6 self, got %s", ip)
	}

	if ip.String() != "2001:db8::2" {
		t.Fatalf("Failed to resolve self to IP, got %s", ip.String())
	}

	i.SetSelf("go-test-nonexisting.gansoi-dev.com:4934")
	ip = i.IP()
	if ip != nil {
		t.Fatalf("Returned something for failing host in Self(), got %s", ip)
	}
}
