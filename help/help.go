// Package help 为 Bubble Tea 应用程序提供简单的帮助视图。
package help

import (
	"strings"

	"github.com/purpose168/bubbles-cn/key"
	tea "github.com/purpose168/bubbletea-cn"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// KeyMap 是用于生成帮助信息的按键绑定映射。由于它是一个接口，
// 可以是任何类型，不过结构体或 map[string][]key.Binding 是常见的选择。
//
// 注意：如果按键被禁用（通过 key.Binding.SetEnabled），它将不会在帮助视图中渲染，
// 因此理论上生成的帮助信息应该能自我管理。
type KeyMap interface {
	// ShortHelp 返回一组绑定，用于在帮助的简短版本中显示。
	// 帮助组件将按照这里返回的帮助项顺序渲染帮助信息。
	ShortHelp() []key.Binding

	// FullHelp 返回一组扩展的帮助项，按列分组。
	// 帮助组件将按照这里返回的帮助项顺序渲染帮助信息。
	FullHelp() [][]key.Binding
}

// Styles 是帮助组件可用的样式定义集合。
type Styles struct {
	Ellipsis lipgloss.Style

	// 简短帮助的样式
	ShortKey       lipgloss.Style
	ShortDesc      lipgloss.Style
	ShortSeparator lipgloss.Style

	// 完整帮助的样式
	FullKey       lipgloss.Style
	FullDesc      lipgloss.Style
	FullSeparator lipgloss.Style
}

// Model 包含帮助视图的状态。
type Model struct {
	Width   int
	ShowAll bool // 如果为 true，渲染"完整"帮助菜单

	ShortSeparator string
	FullSeparator  string

	// 在简短帮助中，当帮助项因宽度而被截断时使用的符号。默认为省略号。
	Ellipsis string

	Styles Styles
}

// New 创建一个带有一些有用默认值的新帮助视图。
func New() Model {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})

	return Model{
		ShortSeparator: " • ",
		FullSeparator:  "    ",
		Ellipsis:       "…",
		Styles: Styles{
			ShortKey:       keyStyle,
			ShortDesc:      descStyle,
			ShortSeparator: sepStyle,
			Ellipsis:       sepStyle,
			FullKey:        keyStyle,
			FullDesc:       descStyle,
			FullSeparator:  sepStyle,
		},
	}
}

// NewModel 创建一个带有一些有用默认值的新帮助视图。
//
// 已弃用：使用 [New] 代替。
var NewModel = New

// Update 帮助满足 Bubble Tea Model 接口。它是一个空操作。
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View 渲染帮助视图的当前状态。
func (m Model) View(k KeyMap) string {
	if m.ShowAll {
		return m.FullHelpView(k.FullHelp())
	}
	return m.ShortHelpView(k.ShortHelp())
}

// ShortHelpView 从按键绑定切片渲染单行帮助视图。
// 如果行长度超过最大宽度，它会被优雅地截断，只显示尽可能多的帮助项。
func (m Model) ShortHelpView(bindings []key.Binding) string {
	if len(bindings) == 0 {
		return ""
	}

	var b strings.Builder
	var totalWidth int
	separator := m.Styles.ShortSeparator.Inline(true).Render(m.ShortSeparator)

	for i, kb := range bindings {
		if !kb.Enabled() {
			continue
		}

		// 分隔符
		var sep string
		if totalWidth > 0 && i < len(bindings) {
			sep = separator
		}

		// 帮助项
		str := sep +
			m.Styles.ShortKey.Inline(true).Render(kb.Help().Key) + " " +
			m.Styles.ShortDesc.Inline(true).Render(kb.Help().Desc)
		w := lipgloss.Width(str)

		// 尾部处理
		if tail, ok := m.shouldAddItem(totalWidth, w); !ok {
			if tail != "" {
				b.WriteString(tail)
			}
			break
		}

		totalWidth += w
		b.WriteString(str)
	}

	return b.String()
}

// FullHelpView 从按键绑定切片的切片渲染帮助列。每个顶层切片条目渲染为一列。
func (m Model) FullHelpView(groups [][]key.Binding) string {
	if len(groups) == 0 {
		return ""
	}

	// 代码注释：此时我们认为预分配此切片的额外代码复杂性不值得。
	//nolint:prealloc
	var (
		out []string

		totalWidth int
		separator  = m.Styles.FullSeparator.Inline(true).Render(m.FullSeparator)
	)

	// 遍历组以构建列
	for i, group := range groups {
		if group == nil || !shouldRenderColumn(group) {
			continue
		}
		var (
			sep          string
			keys         []string
			descriptions []string
		)

		// 分隔符
		if totalWidth > 0 && i < len(groups) {
			sep = separator
		}

		// 将按键和描述分离到不同的切片中
		for _, kb := range group {
			if !kb.Enabled() {
				continue
			}
			keys = append(keys, kb.Help().Key)
			descriptions = append(descriptions, kb.Help().Desc)
		}

		// 列
		col := lipgloss.JoinHorizontal(lipgloss.Top,
			sep,
			m.Styles.FullKey.Render(strings.Join(keys, "\n")),
			" ",
			m.Styles.FullDesc.Render(strings.Join(descriptions, "\n")),
		)
		w := lipgloss.Width(col)

		// 尾部处理
		if tail, ok := m.shouldAddItem(totalWidth, w); !ok {
			if tail != "" {
				out = append(out, tail)
			}
			break
		}

		totalWidth += w
		out = append(out, col)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}

// shouldAddItem 检查是否应该添加新项，考虑当前总宽度和新项宽度。
// 返回值：
// - tail: 如果空间不足，返回要添加的尾部字符串（通常是省略号）
// - ok: 是否可以添加新项
func (m Model) shouldAddItem(totalWidth, width int) (tail string, ok bool) {
	// 如果有空间显示省略号，则显示它。
	if m.Width > 0 && totalWidth+width > m.Width {
		tail = " " + m.Styles.Ellipsis.Inline(true).Render(m.Ellipsis)

		if totalWidth+lipgloss.Width(tail) < m.Width {
			return tail, false
		}
	}
	return "", true
}

// shouldRenderColumn 检查列是否应该渲染（即是否有启用的绑定）。
func shouldRenderColumn(b []key.Binding) (ok bool) {
	for _, v := range b {
		if v.Enabled() {
			return true
		}
	}
	return false
}
