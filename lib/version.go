package lib

// TODO  Use a getter to determine the version and switch between
//
//	   real version and possible the commit hash, which can be
//		 sent through build metadata, just like the version.
var appVersion string
var goVersion string
var commitSha string

/*
GetVersion returns the application version set by the build process.
If the binary is a dev build, then it returns a truncated version
of the latest commit hash.
*/
func GetVersion() string {
	if appVersion == "" {
		if commitSha == "" {
			return "invalid version"
		}
		return commitSha[:8]
	}
	return appVersion
}

func GetGoVersion() string {
	return goVersion
}
