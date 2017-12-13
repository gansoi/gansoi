package filesystem

type filesystemInfo struct {
	Device      string
	Total       int64
	Used        int64
	Available   int64
	UsedPercent float64
	Mountpoint  string
}

func (fi filesystemInfo) isRoot() bool {
	return fi.Mountpoint == "/"
}
