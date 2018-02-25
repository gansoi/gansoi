// Package build keeps information about the current Gansoi build.
package build

import (
	"fmt"
	"strconv"
	"time"
)

var (
	// Version will hold a string describing the Gansoi version. If the
	// version is unknown it will be "undefined".
	Version = "undefined"

	// SHA is the full Git commit checksum.
	SHA = "unknown"

	// ShortSHA is a 7 character Git commit SHA. Please note that this is not
	// guaranteed to be unique like git rev-parse guarantees.
	ShortSHA = "unknown"

	// Time is the local build time.
	Time = "unknown"

	// timestamp is used as intermediate from the build step.
	timestamp string

	UserAgent string
)

func init() {
	if len(timestamp) > 0 {
		t, _ := strconv.ParseInt(timestamp, 10, 64)

		Time = time.Unix(t, 0).UTC().Format(time.RFC3339)
	}

	if len(SHA) > 7 {
		ShortSHA = SHA[0:7]
	}

	UserAgent = fmt.Sprintf("Gansoi/%s (%s)", Version, ShortSHA)
}
