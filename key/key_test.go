package key

import (
	"testing"
)

// TestBinding_Enabled 测试 Binding 的 Enabled 方法。
// 此测试验证绑定在不同状态下的启用状态。
func TestBinding_Enabled(t *testing.T) {
	// 创建一个新的绑定，设置按键为 "k" 和 "up"，帮助信息为 "↑/k" 和 "move up"
	binding := NewBinding(
		WithKeys("k", "up"),
		WithHelp("↑/k", "move up"),
	)
	// 验证新创建的绑定默认是启用的
	if !binding.Enabled() {
		t.Errorf("expected key to be Enabled")
	}

	// 禁用绑定
	binding.SetEnabled(false)
	// 验证绑定现在是禁用的
	if binding.Enabled() {
		t.Errorf("expected key not to be Enabled")
	}

	// 重新启用绑定
	binding.SetEnabled(true)
	// 解绑绑定
	binding.Unbind()
	// 验证解绑后的绑定是禁用的
	if binding.Enabled() {
		t.Errorf("expected key not to be Enabled")
	}
}
