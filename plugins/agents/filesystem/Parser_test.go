package filesystem

import (
	"testing"
)

func TestParseCommandsOutput(t *testing.T) {
	commandsOutput := []byte(`Filesystem     1K-blocks      Used Available Use% Mounted on
dev              4046984         0   4046984   0% /dev
/dev/sda1      215322880 153861796  50453548  76% /
/dev/sda2            4052580    203316   3849264   6% /tmp
`)
	parser := dfCommandParser{}
	filesystems, getFilesystemsError := parser.parse(commandsOutput)
	if getFilesystemsError != nil {
		t.Error(getFilesystemsError)
	}
	if len(filesystems) != 2 {
		t.Error("Invalid amount of found filesystems.")
	}
}

func TestParseCommandsOnParseLineError(t *testing.T) {
	commandsOutput := []byte(`/dev/sda1      2153huehue-not-a-number22880 153861796  50453548  76% /`)
	parser := dfCommandParser{}
	filesystems, getFilesystemsError := parser.parse(commandsOutput)
	if getFilesystemsError == nil {
		t.Error("Despite failed line parsing, the method did not return error")
	}
	if filesystems != nil {
		t.Error("filesystems should be nil")
	}
}

func TestParseCommandsOfLineThatShouldNotBeParsed(t *testing.T) {
	commandsOutput := []byte(`not-a-device      215322880 153861796  50453548  76% /`)
	parser := dfCommandParser{}
	filesystems, getFilesystemsError := parser.parse(commandsOutput)
	if getFilesystemsError != nil {
		t.Error("The line should not be parsed, but no error should be returned")
	}
	if len(filesystems) != 0 {
		t.Error("filesystems should be empty")
	}
}
func TestParseLineNotParsable(t *testing.T) {
	commandsOutput := []byte(`not-a-device      215322880 153861796  50453548  76% /`)
	parser := dfCommandParser{}
	filesystemInfo, parseLineError := parser.parseLine(commandsOutput)
	if parseLineError != nil {
		t.Error("parseLineError should be nil")
	}
	if filesystemInfo != nil {
		t.Error("filesystemInfo should be nil")
	}
}

func TestParseLineNotParsableNumbers(t *testing.T) {
	exampleOutputs := [][]byte{
		[]byte(`/dev/blabla      215322880 1538-hue-61796  50453548  76% /`),
		[]byte(`/dev/blabla      215-hue-322880 153861796  50453548  76% /`),
		[]byte(`/dev/blabla      215322880 153861796  50453-hue-548  76% /`),
	}
	parser := dfCommandParser{}
	for _, commandsOutput := range exampleOutputs {
		filesystemInfo, parseLineError := parser.parseLine(commandsOutput)
		if parseLineError == nil {
			t.Error("parseLineError should not be nil")
		}
		if filesystemInfo != nil {
			t.Error("filesystemInfo should be nil")
		}
	}
}
func TestParseLineWithZeros(t *testing.T) {
	commandsOutput := []byte(`/dev/blabla      0 0 0  100% /`)

	parser := dfCommandParser{}
	filesystemInfo, parseLineError := parser.parseLine(commandsOutput)

	if parseLineError != nil {
		t.Error("parseLineError should be nil")
	}
	if filesystemInfo == nil {
		t.Error("filesystemInfo should not be nil")
	}

	if filesystemInfo.Availabe != 0 || filesystemInfo.Used != 0 || filesystemInfo.Total != 0 {
		t.Error("Disk values incorrectly parsed")
	}
	if filesystemInfo.UsedPercent > 0 {
		t.Error("If total is 0, percentage should not be calculated")
	}

	if filesystemInfo.Mountpoint != "/" {
		t.Error("Mountpoint incorrectly parsed")
	}

}
func TestParseLine(t *testing.T) {
	commandsOutput := []byte(`/dev/blabla      100 25 75  25% /`)
	parser := dfCommandParser{}
	filesystemInfo, _ := parser.parseLine(commandsOutput)

	if filesystemInfo.Availabe != 75 || filesystemInfo.Used != 25 || filesystemInfo.Total != 100 {
		t.Error("Disk values incorrectly parsed")
	}
	if filesystemInfo.UsedPercent != 25.0 {
		t.Error("If total is 0, percentage should not be calculated")
	}

}
