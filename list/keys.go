package list

import "github.com/purpose168/bubbles-cn/key"

// KeyMap 定义了按键绑定。它满足 help.KeyMap 接口，用于渲染菜单。
type KeyMap struct {
	// 浏览列表时使用的按键绑定。
	CursorUp    key.Binding // 光标向上
	CursorDown  key.Binding // 光标向下
	NextPage    key.Binding // 下一页
	PrevPage    key.Binding // 上一页
	GoToStart   key.Binding // 前往开始
	GoToEnd     key.Binding // 前往结束
	Filter      key.Binding // 过滤器
	ClearFilter key.Binding // 清除过滤器

	// 设置过滤器时使用的按键绑定。
	CancelWhileFiltering key.Binding // 取消过滤
	AcceptWhileFiltering key.Binding // 接受过滤

	// 帮助切换按键绑定。
	ShowFullHelp  key.Binding // 显示完整帮助
	CloseFullHelp key.Binding // 关闭完整帮助

	// 退出按键绑定。在过滤时不会被捕获。
	Quit key.Binding // 退出

	// 强制退出按键绑定。在过滤时也会被捕获。
	ForceQuit key.Binding // 强制退出
}

// DefaultKeyMap 返回一组默认的按键绑定。
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// 浏览。
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),

		// 过滤。
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),

		// 切换帮助。
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		// 退出。
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
