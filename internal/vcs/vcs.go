package vcs

import "runtime/debug"

func Version() string {
	i, ok := debug.ReadBuildInfo()
	if ok {
		return i.Main.Version
	}

	return ""
}
