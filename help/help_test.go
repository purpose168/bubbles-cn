package help

import (
	"fmt"
	"testing"

	"github.com/purpose168/charm-experimental-packages-cn/exp/golden"

	"github.com/purpose168/bubbles-cn/key"
)

// TestFullHelp 测试完整帮助视图的渲染功能。
// 此测试验证在不同宽度下，FullHelpView 方法是否能正确渲染帮助信息。
func TestFullHelp(t *testing.T) {
	// 创建新的帮助模型
	m := New()
	// 设置完整帮助视图的分隔符
	m.FullSeparator = " | "

	// 创建按键绑定
	k := key.WithKeys("x")
	kb := [][]key.Binding{
		{
			// 第一组按键绑定：enter 键继续
			key.NewBinding(k, key.WithHelp("enter", "continue")),
		},
		{
			// 第二组按键绑定：esc 键返回，? 键显示帮助
			key.NewBinding(k, key.WithHelp("esc", "back")),
			key.NewBinding(k, key.WithHelp("?", "help")),
		},
		{
			// 第三组按键绑定：H 键主页，ctrl+c 键退出，ctrl+l 键日志
			key.NewBinding(k, key.WithHelp("H", "home")),
			key.NewBinding(k, key.WithHelp("ctrl+c", "quit")),
			key.NewBinding(k, key.WithHelp("ctrl+l", "log")),
		},
	}

	// 测试不同宽度下的帮助视图渲染
	for _, w := range []int{20, 30, 40} {
		t.Run(fmt.Sprintf("full help %d width", w), func(t *testing.T) {
			// 设置帮助视图宽度
			m.Width = w
			// 生成帮助视图
			s := m.FullHelpView(kb)
			// 使用 golden 测试库验证输出是否符合预期
			golden.RequireEqual(t, []byte(s))
		})
	}
}
