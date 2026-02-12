// Package viewport 提供了一个在 Bubble Tea 应用程序中渲染视口的组件。
package viewport

import "github.com/purpose168/bubbles-cn/key"

const spacebar = " " // 空格常量

// KeyMap 定义了视口的按键绑定。注意，你并不一定需要使用按键绑定；
// 视口可以通过 Model.LineDown(1) 等方法以编程方式控制。
// 详情请参阅 GoDocs。
type KeyMap struct {
	PageDown     key.Binding // 向下翻页
	PageUp       key.Binding // 向上翻页
	HalfPageUp   key.Binding // 向上半页
	HalfPageDown key.Binding // 向下半页
	Down         key.Binding // 向下移动一行
	Up           key.Binding // 向上移动一行
	Left         key.Binding // 向左移动一列
	Right        key.Binding // 向右移动一列
}

// DefaultKeyMap 返回一组类似分页器的默认按键绑定。
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// 向下翻页：pgdown、空格、f
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", spacebar, "f"),
			key.WithHelp("f/pgdn", "向下翻页"),
		),
		// 向上翻页：pgup、b
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "向上翻页"),
		),
		// 向上半页：u、ctrl+u
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "半页向上"),
		),
		// 向下半页：d、ctrl+d
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "半页向下"),
		),
		// 向上移动一行：上箭头、k
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "向上"),
		),
		// 向下移动一行：下箭头、j
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "向下"),
		),
		// 向左移动一列：左箭头、h
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "向左移动"),
		),
		// 向右移动一列：右箭头、l
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "向右移动"),
		),
	}
}
