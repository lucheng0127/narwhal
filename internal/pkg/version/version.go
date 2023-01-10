package version

import "fmt"

const (
	MAJOR  uint   = 0
	MINOR  uint   = 0
	PATCH  uint   = 0
	SUFFIX string = ""
)

func Version() string {
	return fmt.Sprintf("%d.%d.%d%s", MAJOR, MINOR, PATCH, SUFFIX)
}
