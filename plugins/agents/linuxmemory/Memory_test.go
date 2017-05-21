package linuxmemory

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports/mock"
)

type (
	Mock struct {
		mock.Mock
		Contents []byte
	}
)

var (
	good = []byte(`MemTotal:       16318960 kB
MemFree:          701092 kB
MemAvailable:   10580100 kB
Buffers:          415120 kB
Cached:         10319428 kB
SwapCached:            0 kB
Active:         10094760 kB
Inactive:        4522724 kB
Active(anon):    3927680 kB
Inactive(anon):  1031152 kB
Active(file):    6167080 kB
Inactive(file):  3491572 kB
Unevictable:        4808 kB
Mlocked:            4808 kB
SwapTotal:      33440252 kB
SwapFree:       33440252 kB
Dirty:               240 kB
Writeback:             0 kB
AnonPages:       3615476 kB
Mapped:           950980 kB
Shmem:           1075900 kB
Slab:             756648 kB
SReclaimable:     562636 kB
SUnreclaim:       194012 kB
KernelStack:       13920 kB
PageTables:        63512 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:    41599732 kB
Committed_AS:   11339424 kB
VmallocTotal:   34359738367 kB
VmallocUsed:           0 kB
VmallocChunk:          0 kB
HardwareCorrupted:     0 kB
AnonHugePages:   1355776 kB
ShmemHugePages:        0 kB
ShmemPmdMapped:        0 kB
CmaTotal:              0 kB
CmaFree:               0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:      517540 kB
DirectMap2M:    16146432 kB
DirectMap1G:           0 kB
`)

	empty = []byte("")
	bad1  = []byte("hello")
	bad2  = []byte("hello: 763443\nhello2:")
	bad3  = []byte("hello: hello kB\n")
)

func (m *Mock) ReadFile(path string) ([]byte, error) {
	return m.Contents, nil
}

func TestCheck(t *testing.T) {
	transport := &Mock{Contents: good}
	mem := &Memory{}
	result := plugins.NewAgentResult()

	err := mem.RemoteCheck(transport, result)

	if err != nil {
		t.Fatalf("RemoteCheck() returned an error: %s", err.Error())
	}

	if result["MemFree"] != int64(701092) {
		t.Fatalf("RemoteCheck() returned wrong value for \"MemFree\", got %d, expected %d", result["MemFree"], 701092)
	}
}

func TestCheckSyntaxError(t *testing.T) {
	cases := [][]byte{empty, bad1, bad2, bad3}
	mem := &Memory{}
	result := plugins.NewAgentResult()

	for i, c := range cases {
		transport := &Mock{Contents: c}
		err := mem.RemoteCheck(transport, result)
		if err == nil {
			t.Fatalf("%d RemoteCheck() did not return an error", i)
		}
	}
}

func TestCheckError(t *testing.T) {
	transport := &mock.Mock{}
	mem := &Memory{}
	result := plugins.NewAgentResult()

	err := mem.RemoteCheck(transport, result)

	if err != mock.ErrNotImplemented {
		t.Fatalf("RemoteCheck() did not return error")
	}
}

var _ plugins.RemoteAgent = (*Memory)(nil)
