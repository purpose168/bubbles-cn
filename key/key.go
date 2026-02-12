// Package key 提供了一些类型和函数，用于生成用户可定义的按键映射，
// 在 Bubble Tea 组件中非常有用。使用此包定义按键映射有几种不同的方法。
// 以下是一个示例：
//
//	type KeyMap struct {
//	    Up key.Binding
//	    Down key.Binding
//	}
//
//	var DefaultKeyMap = KeyMap{
//	    Up: key.NewBinding(
//	        key.WithKeys("k", "up"),        // 实际的按键绑定
//	        key.WithHelp("↑/k", "move up"), // 对应的帮助文本
//	    ),
//	    Down: key.NewBinding(
//	        key.WithKeys("j", "down"),
//	        key.WithHelp("↓/j", "move down"),
//	    ),
//	}
//
//	func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.KeyMsg:
//	        switch {
//	        case key.Matches(msg, DefaultKeyMap.Up):
//	            // 用户按下了上键
//	        case key.Matches(msg, DefaultKeyMap.Down):
//	            // 用户按下了下键
//	        }
//	    }
//
//	    // ...
//	}
//
// 上面示例中未使用的帮助信息可用于在视图中渲染按键的帮助文本。
package key

import "fmt"

// Binding 描述了一组按键绑定以及可选的相关帮助文本。
type Binding struct {
	keys     []string // 按键列表
	help     Help     // 帮助信息
	disabled bool     // 是否禁用
}

// BindingOpt 是按键绑定的初始化选项。它用作 NewBinding 的参数。
type BindingOpt func(*Binding)

// NewBinding 从一组 BindingOpt 选项返回一个新的按键绑定。
func NewBinding(opts ...BindingOpt) Binding {
	b := &Binding{}
	for _, opt := range opts {
		opt(b)
	}
	return *b
}

// WithKeys 使用给定的按键初始化按键绑定。
func WithKeys(keys ...string) BindingOpt {
	return func(b *Binding) {
		b.keys = keys
	}
}

// WithHelp 使用给定的帮助文本初始化按键绑定。
func WithHelp(key, desc string) BindingOpt {
	return func(b *Binding) {
		b.help = Help{Key: key, Desc: desc}
	}
}

// WithDisabled 初始化一个已禁用的按键绑定。
func WithDisabled() BindingOpt {
	return func(b *Binding) {
		b.disabled = true
	}
}

// SetKeys 设置按键绑定的按键。
func (b *Binding) SetKeys(keys ...string) {
	b.keys = keys
}

// Keys 返回按键绑定的按键。
func (b Binding) Keys() []string {
	return b.keys
}

// SetHelp 设置按键绑定的帮助文本。
func (b *Binding) SetHelp(key, desc string) {
	b.help = Help{Key: key, Desc: desc}
}

// Help 返回按键绑定的帮助信息。
func (b Binding) Help() Help {
	return b.help
}

// Enabled 返回按键绑定是否启用。禁用的按键绑定不会被激活，也不会在帮助中显示。
// 按键绑定默认是启用的。
func (b Binding) Enabled() bool {
	return !b.disabled && b.keys != nil
}

// SetEnabled 启用或禁用按键绑定。
func (b *Binding) SetEnabled(v bool) {
	b.disabled = !v
}

// Unbind 从绑定中移除按键和帮助，有效地使其无效。这比禁用它更进一步，
// 因为应用程序可以根据应用程序状态启用或禁用按键绑定。
func (b *Binding) Unbind() {
	b.keys = nil
	b.help = Help{}
}

// Help 是给定按键绑定的帮助信息。
type Help struct {
	Key  string // 按键
	Desc string // 描述
}

// Matches 检查给定的按键是否匹配给定的绑定。
func Matches[Key fmt.Stringer](k Key, b ...Binding) bool {
	keys := k.String()
	for _, binding := range b {
		for _, v := range binding.keys {
			if keys == v && binding.Enabled() {
				return true
			}
		}
	}
	return false
}
