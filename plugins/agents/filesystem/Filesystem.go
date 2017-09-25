package filesystem

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports"
)

func init() {
	plugins.RegisterAgent("filesystem", Filesystem{})
}

// Filesystem will check the condition of mounted storage devices.
type Filesystem struct {
	CommaSeparatedExcludedDevices string `json:"excludedDevices" description:"Comma-separated list of devices, that will we excluded from the check"`
	excludedDevices               []string
	parser                        commandParser
}

// RemoteCheck implements plugins.RemoteAgent.
func (fs *Filesystem) RemoteCheck(transport transports.Transport, result plugins.AgentResult) error {
	if fs.parser == nil {
		fs.parser = &dfCommandParser{}
	}
	commandsOutput, invokeError := fs.invokeRemoteCommand(transport)
	if invokeError != nil {
		return invokeError
	}
	filesystems, parseError := fs.parser.parse(commandsOutput)
	if parseError != nil {
		return parseError
	}
	return fs.setResults(result, filesystems)
}

func (fs *Filesystem) invokeRemoteCommand(transport transports.Transport) ([]byte, error) {
	out, _, err := transport.Exec("df -k")
	if err != nil {
		return nil, errors.Wrap(err, "df -k")
	}
	return ioutil.ReadAll(out)
}

func (fs *Filesystem) setResults(result plugins.AgentResult, filesystems []filesystemInfo) error {
	var fsInWorstConditions, rootFs *filesystemInfo

	for i, currentFs := range filesystems {
		if fs.isDeviceExcludedFromCheck(currentFs.Device) {
			continue
		}
		if fsInWorstConditions == nil || currentFs.UsedPercent >= fsInWorstConditions.UsedPercent {
			fsInWorstConditions = &filesystems[i]
		}
		if currentFs.isRoot() {
			rootFs = &filesystems[i]
		}
	}

	if rootFs != nil {
		fs.setSingleDevicesResults("Root", rootFs, result)
	}

	if fsInWorstConditions != nil {
		fs.setSingleDevicesResults("Worst", fsInWorstConditions, result)
	} else {
		return errors.New("Could not find a mounted device")
	}

	return nil
}

func (fs *Filesystem) isDeviceExcludedFromCheck(deviceName string) bool {
	if fs.excludedDevices == nil {
		fs.populateExcludedDevices()
	}
	for _, excluded := range fs.excludedDevices {
		if deviceName == excluded {
			return true
		}
	}
	return false
}

func (fs *Filesystem) populateExcludedDevices() {
	fs.excludedDevices = make([]string, 0)
	for _, v := range strings.Split(fs.CommaSeparatedExcludedDevices, ",") {
		device := strings.TrimSpace(v)
		if device != "" {
			fs.excludedDevices = append(fs.excludedDevices, device)
		}
	}

}
func (fs *Filesystem) setSingleDevicesResults(name string, fi *filesystemInfo, result plugins.AgentResult) {
	result.AddValue(name+"Device", fi.Device)
	result.AddValue(name+"Total", fi.Total)
	result.AddValue(name+"Used", fi.Used)
	result.AddValue(name+"Available", fi.Availabe)
	result.AddValue(name+"UsedPercent", fi.UsedPercent)
	result.AddValue(name+"Mountpoint", fi.Mountpoint)
}
