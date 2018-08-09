package zutility

import (
	"runtime"
)

////////////////////////////////////////////////////////////////////////////////
func IsWindows() bool {
	return `windows` == runtime.GOOS
}

func IsLinux() bool {
	return `linux` == runtime.GOOS
}

func IsMac() bool {
	return `darwin` == runtime.GOOS
}

func IsIos() bool {
	return `darwin` == runtime.GOOS
}

func ShowOS() string {
	return runtime.GOOS
}
