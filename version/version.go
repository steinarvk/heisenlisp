package version

import "strings"

var VersionString string
var CommitHash string
var BuildTimestampISO8601 string
var BuildMachineInfo string
var GoVersion string

func init() {
	VersionString = strings.Replace(VersionString, "___", " ", -1)
	CommitHash = strings.Replace(CommitHash, "___", " ", -1)
	BuildTimestampISO8601 = strings.Replace(BuildTimestampISO8601, "___", " ", -1)
	BuildMachineInfo = strings.Replace(BuildMachineInfo, "___", " ", -1)
	GoVersion = strings.Replace(GoVersion, "___", " ", -1)
}
