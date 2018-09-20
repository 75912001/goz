package zutility

import (
	"runtime"
)

//IsWindows win
func IsWindows() bool {
	return `windows` == runtime.GOOS
}

//IsLinux linux
func IsLinux() bool {
	return `linux` == runtime.GOOS
}

//IsDarwin darwin
func IsDarwin() bool {
	return `darwin` == runtime.GOOS
}

//ShowOS 显示os的名称
func ShowOS() string {
	return runtime.GOOS
}
