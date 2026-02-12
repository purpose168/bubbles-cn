package list

import (
	"fmt"
	"io"
	"strings"

	"github.com/purpose168/bubbles-cn/key"
	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// DefaultItemStyles 定义了默认列表项的样式。
// 有关这些样式何时生效，请参见 DefaultItemView。
type DefaultItemStyles struct {
	// 正常状态。
	NormalTitle lipgloss.Style
	NormalDesc  lipgloss.Style

	// 选中项状态。
	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style

	// 暗淡状态，用于过滤器输入最初激活时。
	DimmedTitle lipgloss.Style
	DimmedDesc  lipgloss.Style

	// 匹配当前过滤器的字符（如果有）。
	FilterMatch lipgloss.Style
}

// NewDefaultItemStyles 返回默认项目的样式定义。
// 有关这些样式何时生效，请参见 DefaultItemView。
func NewDefaultItemStyles() (s DefaultItemStyles) {
	s.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2) //nolint:mnd

	s.NormalDesc = s.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	s.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2) //nolint:mnd

	s.DimmedDesc = s.DimmedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})

	s.FilterMatch = lipgloss.NewStyle().Underline(true)

	return s
}

// DefaultItem 描述了一个设计用于与 DefaultDelegate 一起工作的项目。
type DefaultItem interface {
	Item
	Title() string
	Description() string
}

// DefaultDelegate 是一个设计用于列表中的标准委托。
// 它由 DefaultItemStyles 设置样式，可以根据需要自定义。
//
// 通过将 Description 设置为 false 可以隐藏描述行，
// 这会将列表渲染为单行项目。项目之间的间距可以通过 SetSpacing 方法设置。
//
// 设置 UpdateFunc 是可选的。如果设置了，它将在 ItemDelegate 被调用时被调用，
// 而 ItemDelegate 是在列表的 Update 函数被调用时被调用的。
//
// 设置 ShortHelpFunc 和 FullHelpFunc 是可选的。它们可以设置为在列表的默认简短和完整帮助菜单中包含项目。
type DefaultDelegate struct {
	ShowDescription bool
	Styles          DefaultItemStyles
	UpdateFunc      func(tea.Msg, *Model) tea.Cmd
	ShortHelpFunc   func() []key.Binding
	FullHelpFunc    func() [][]key.Binding
	height          int
	spacing         int
}

// NewDefaultDelegate 创建一个带有默认样式的新委托。
func NewDefaultDelegate() DefaultDelegate {
	const defaultHeight = 2
	const defaultSpacing = 1
	return DefaultDelegate{
		ShowDescription: true,
		Styles:          NewDefaultItemStyles(),
		height:          defaultHeight,
		spacing:         defaultSpacing,
	}
}

// SetHeight 设置委托的首选高度。
func (d *DefaultDelegate) SetHeight(i int) {
	d.height = i
}

// Height 返回委托的首选高度。
// 这仅在 ShowDescription 为 true 时有效，否则高度始终为 1。
func (d DefaultDelegate) Height() int {
	if d.ShowDescription {
		return d.height
	}
	return 1
}

// SetSpacing 设置委托的间距。
func (d *DefaultDelegate) SetSpacing(i int) {
	d.spacing = i
}

// Spacing 返回委托的间距。
func (d DefaultDelegate) Spacing() int {
	return d.spacing
}

// Update 检查委托的 UpdateFunc 是否设置，并调用它。
func (d DefaultDelegate) Update(msg tea.Msg, m *Model) tea.Cmd {
	if d.UpdateFunc == nil {
		return nil
	}
	return d.UpdateFunc(msg, m)
}

// Render 打印一个项目。
func (d DefaultDelegate) Render(w io.Writer, m Model, index int, item Item) {
	var (
		title, desc  string
		matchedRunes []int
		s            = &d.Styles
	)

	// 检查项目是否实现了 DefaultItem 接口
	if i, ok := item.(DefaultItem); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	if m.width <= 0 {
		// 短路，宽度无效时直接返回
		return
	}

	// 防止文本超过列表宽度
	textwidth := m.width - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()
	title = ansi.Truncate(title, textwidth, ellipsis)
	if d.ShowDescription {
		var lines []string
		for i, line := range strings.Split(desc, "\n") {
			if i >= d.height-1 {
				break
			}
			lines = append(lines, ansi.Truncate(line, textwidth, ellipsis))
		}
		desc = strings.Join(lines, "\n")
	}

	// 条件判断
	var (
		isSelected  = index == m.Index()                                               // 是否选中
		emptyFilter = m.FilterState() == Filtering && m.FilterValue() == ""            // 是否为空过滤器
		isFiltered  = m.FilterState() == Filtering || m.FilterState() == FilterApplied // 是否处于过滤状态
	)

	if isFiltered && index < len(m.filteredItems) {
		// 获取匹配字符的索引
		matchedRunes = m.MatchesForItem(index)
	}

	// 根据不同状态应用不同样式
	if emptyFilter {
		// 空过滤器状态
		title = s.DimmedTitle.Render(title)
		desc = s.DimmedDesc.Render(desc)
	} else if isSelected && m.FilterState() != Filtering {
		// 选中状态
		if isFiltered {
			// 高亮匹配项
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		// 正常状态
		if isFiltered {
			// 高亮匹配项
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}

	// 输出渲染结果
	if d.ShowDescription {
		fmt.Fprintf(w, "%s\n%s", title, desc) //nolint: errcheck
		return
	}
	fmt.Fprintf(w, "%s", title) //nolint: errcheck
}

// ShortHelp 返回委托的简短帮助。
func (d DefaultDelegate) ShortHelp() []key.Binding {
	if d.ShortHelpFunc != nil {
		return d.ShortHelpFunc()
	}
	return nil
}

// FullHelp 返回委托的完整帮助。
func (d DefaultDelegate) FullHelp() [][]key.Binding {
	if d.FullHelpFunc != nil {
		return d.FullHelpFunc()
	}
	return nil
}
