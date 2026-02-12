//go:build !windows
// +build !windows

package filepicker

import "strings"

// IsHidden 报告文件是否为隐藏文件。
// 在 Unix 系统中，以点开头的文件被视为隐藏文件。
func IsHidden(file string) (bool, error) {
	return strings.HasPrefix(file, "."), nil
}
