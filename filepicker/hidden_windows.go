//go:build windows
// +build windows

package filepicker

import (
	"syscall"
)

// IsHidden 报告文件是否为隐藏文件。
// 在 Windows 系统中，通过检查文件属性来判断文件是否为隐藏文件。
func IsHidden(file string) (bool, error) {
	// 将文件路径转换为 UTF16 指针，这是 Windows API 所需的格式
	pointer, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return false, err //nolint:wrapcheck
	}
	// 获取文件属性
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err //nolint:wrapcheck
	}
	// 检查文件是否具有隐藏属性
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
