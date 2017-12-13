package filesystem

import (
	"bytes"
	"strconv"
	"strings"
)

type commandParser interface {
	parse([]byte) ([]filesystemInfo, error)
}

type dfCommandParser struct {
}

func (p *dfCommandParser) parse(theOutput []byte) ([]filesystemInfo, error) {
	lines := bytes.Split(theOutput, []byte("\n"))
	filesystems := make([]filesystemInfo, 0)

	for _, line := range lines {
		newFsInfo, newFsErr := p.parseLine(line)
		if newFsErr != nil {
			return nil, newFsErr
		}
		if newFsInfo == nil {
			continue
		}
		filesystems = append(filesystems, *newFsInfo)
	}
	return filesystems, nil
}

func (p *dfCommandParser) parseLine(singleLine []byte) (*filesystemInfo, error) {
	// Remove entries like tmpfs
	if len(singleLine) == 0 || singleLine[0] != byte('/') {
		return nil, nil
	}
	fields := strings.Fields(string(singleLine))
	newFsInfo := &filesystemInfo{
		Device:     fields[0],
		Mountpoint: fields[5],
	}
	parseErr := p.parseValues(fields, newFsInfo)
	if parseErr != nil {
		return nil, parseErr
	}
	if newFsInfo.Total != 0 {
		newFsInfo.UsedPercent = (float64(newFsInfo.Used) / float64(newFsInfo.Total)) * 100
	}
	return newFsInfo, nil
}

func (p *dfCommandParser) parseValues(fields []string, fi *filesystemInfo) error {
	var parseErr error
	fi.Total, parseErr = strconv.ParseInt(fields[1], 10, 64)
	if parseErr != nil {
		return parseErr
	}
	fi.Used, parseErr = strconv.ParseInt(fields[2], 10, 64)
	if parseErr != nil {
		return parseErr
	}
	fi.Available, parseErr = strconv.ParseInt(fields[3], 10, 64)
	if parseErr != nil {
		return parseErr
	}
	return nil
}
