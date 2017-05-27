package linuxmemory

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports"
)

type (
	Memory struct{}
)

var (
	// ErrSyntax will be returned, if we don't understand the format of /proc/meminfo.
	ErrSyntax = errors.New("Unknown format of /proc/meminfo")
)

func init() {
	plugins.RegisterAgent("linuxmemory", Memory{})
}

func (m *Memory) RemoteCheck(transport transports.Transport, result plugins.AgentResult) error {
	contents, err := transport.ReadFile("/proc/meminfo")
	if err != nil {
		return err
	}

	if len(contents) == 0 {
		return ErrSyntax
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(contents))
	for scanner.Scan() {
		text := scanner.Text()

		n := strings.Index(text, ":")
		if n == -1 {
			return ErrSyntax
		}

		key := text[:n]
		data := strings.Split(strings.Trim(text[(n+1):], " "), " ")
		if len(data) == 1 {
			value, err := strconv.ParseInt(data[0], 10, 64)
			if err != nil {
				return ErrSyntax
			}
			result.AddValue(key, value)
		} else if len(data) == 2 {
			if data[1] == "kB" {
				value, err := strconv.ParseInt(data[0], 10, 64)
				if err != nil {
					return ErrSyntax
				}

				result.AddValue(key, value)
			}
		}
	}

	return nil
}
