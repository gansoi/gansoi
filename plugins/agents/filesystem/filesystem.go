package filesystem

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports"
)

func init() {
	plugins.RegisterAgent("filesystem", Filesystem{})
}

// Filesystem will check if a process with exact name is running on the host.
type Filesystem struct {
}

type FilesystemInfo struct {
	Device      string
	Total       int64
	Used        int64
	Availabe    int64
	UsedPercent float64
	Mountpoint  string
}

// RemoteCheck implements plugins.RemoteAgent.
func (m *Filesystem) RemoteCheck(transport transports.Transport, result plugins.AgentResult) error {
	out, _, err := transport.Exec("df -k")
	if err != nil {
		return errors.Wrap(err, "df -k")
	}

	resultAsBytes, err := ioutil.ReadAll(out)
	if err != nil {
		return err
	}
	lines := bytes.Split(resultAsBytes, []byte("\n"))

	filesystems := make([]FilesystemInfo, 0)
	var rootFs *FilesystemInfo

	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > 0 && lines[i][0] == byte('/') {
			fields := strings.Fields(string(lines[i]))
			var used, available int64
			total, parseErr := strconv.ParseInt(fields[1], 10, 64)
			if parseErr != nil {
				return parseErr
			}
			used, parseErr = strconv.ParseInt(fields[2], 10, 64)
			if parseErr != nil {
				return parseErr
			}
			available, parseErr = strconv.ParseInt(fields[3], 10, 64)
			if parseErr != nil {
				return parseErr
			}
			newFsInfo := FilesystemInfo{
				Device:     fields[0],
				Total:      total,
				Used:       used,
				Availabe:   available,
				Mountpoint: fields[5],
			}
			newFsInfo.UsedPercent = (float64(used) / float64(total)) * 100
			fmt.Println(newFsInfo)
			filesystems = append(filesystems, newFsInfo)
			if newFsInfo.Mountpoint == "/" {
				rootFs = &newFsInfo
			}

		}
	}

	if rootFs == nil {
		return errors.New("Could not find root mount device")
	}

	result.AddValue("RootDevice", rootFs.Device)
	result.AddValue("RootTotal", rootFs.Total)
	result.AddValue("RootUsed", rootFs.Used)
	result.AddValue("RootAvailable", rootFs.Availabe)
	result.AddValue("RootUsedPercent", rootFs.UsedPercent)

	highestPercent := rootFs

	for i := 0; i < len(filesystems); i++ {
		if filesystems[i].UsedPercent > highestPercent.UsedPercent {
			highestPercent = &filesystems[i]
		}
	}

	result.AddValue("HighestDevice", highestPercent.Device)
	result.AddValue("HighestTotal", highestPercent.Total)
	result.AddValue("HighestUsed", highestPercent.Used)
	result.AddValue("HighestAvailable", highestPercent.Availabe)
	result.AddValue("HighestUsedPercent", highestPercent.UsedPercent)
	result.AddValue("HighestMountpoint", highestPercent.Mountpoint)

	return nil
}
