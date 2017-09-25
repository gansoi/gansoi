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
	return fs.setResult(result, filesystems)
}

func (fs *Filesystem) invokeRemoteCommand(transport transports.Transport) ([]byte, error) {
	out, _, err := transport.Exec("df -k")
	if err != nil {
		return nil, errors.Wrap(err, "df -k")
	}
	return ioutil.ReadAll(out)
}

func (fs *Filesystem) setResult(result plugins.AgentResult, filesystems []filesystemInfo) error {
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
		result.AddValue("RootDevice", rootFs.Device)
		result.AddValue("RootTotal", rootFs.Total)
		result.AddValue("RootUsed", rootFs.Used)
		result.AddValue("RootAvailable", rootFs.Availabe)
		result.AddValue("RootUsedPercent", rootFs.UsedPercent)
		result.AddValue("RootMountpoint", rootFs.Mountpoint)
	}

	if fsInWorstConditions != nil {
		result.AddValue("WorstDevice", fsInWorstConditions.Device)
		result.AddValue("WorstTotal", fsInWorstConditions.Total)
		result.AddValue("WorstUsed", fsInWorstConditions.Used)
		result.AddValue("WorstAvailable", fsInWorstConditions.Availabe)
		result.AddValue("WorstUsedPercent", fsInWorstConditions.UsedPercent)
		result.AddValue("WorstMountpoint", fsInWorstConditions.Mountpoint)
	} else {
		return errors.New("Could not find a mounted device")
	}

	return nil
}

func (fs *Filesystem) isDeviceExcludedFromCheck(deviceName string) bool {
	if fs.excludedDevices == nil {
		fs.excludedDevices = make([]string, 0)
		for _, v := range strings.Split(fs.CommaSeparatedExcludedDevices, ",") {
			device := strings.TrimSpace(v)
			if device != "" {
				fs.excludedDevices = append(fs.excludedDevices, device)
			}
		}
	}
	for _, excluded := range fs.excludedDevices {
		if deviceName == excluded {
			return true
		}
	}
	return false
}
