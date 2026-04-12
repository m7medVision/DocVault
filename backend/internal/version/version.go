// Package version provides build information.
package version

import "runtime/debug"

// Info holds version information populated at build time.
var Info = struct {
	Version   string
	Commit    string
	BuildTime string
}{
	Version:   "dev",
	Commit:    "unknown",
	BuildTime: "unknown",
}

// Init populates version info from build settings.
func Init() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				Info.Commit = setting.Value
			case "vcs.time":
				Info.BuildTime = setting.Value
			case "vcs.modification":
				if setting.Value == "true" {
					Info.Commit += "-dirty"
				}
			}
		}
		Info.Version = buildInfo.Main.Version
		if Info.Version == "" {
			Info.Version = "dev"
		}
	}
}

// Get returns version info as a map.
func Get() map[string]string {
	return map[string]string{
		"version":    Info.Version,
		"commit":     Info.Commit,
		"build_time": Info.BuildTime,
	}
}
