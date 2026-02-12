package list

import (
	lipgloss "github.com/purpose168/lipgloss-cn"
)

const (
	// bullet 用于列表项的圆点符号
	bullet = "•"
	// ellipsis 用于文本截断的省略号
	ellipsis = "…"
)

// Styles 包含此列表组件的样式定义。默认情况下，这些值由 DefaultStyles 生成。
type Styles struct {
	// TitleBar 标题栏样式
	TitleBar lipgloss.Style
	// Title 标题样式
	Title lipgloss.Style
	// Spinner 加载动画样式
	Spinner lipgloss.Style
	// FilterPrompt 过滤提示符样式
	FilterPrompt lipgloss.Style
	// FilterCursor 过滤光标样式
	FilterCursor lipgloss.Style

	// DefaultFilterCharacterMatch 过滤器中匹配字符的默认样式。可由委托覆盖。
	DefaultFilterCharacterMatch lipgloss.Style

	// StatusBar 状态栏样式
	StatusBar lipgloss.Style
	// StatusEmpty 空状态样式
	StatusEmpty lipgloss.Style
	// StatusBarActiveFilter 激活过滤器时的状态栏样式
	StatusBarActiveFilter lipgloss.Style
	// StatusBarFilterCount 过滤器计数样式
	StatusBarFilterCount lipgloss.Style

	// NoItems 无项目时的样式
	NoItems lipgloss.Style

	// PaginationStyle 分页样式
	PaginationStyle lipgloss.Style
	// HelpStyle 帮助样式
	HelpStyle lipgloss.Style

	// Styled characters 样式化字符
	// ActivePaginationDot 激活的分页点样式
	ActivePaginationDot lipgloss.Style
	// InactivePaginationDot 未激活的分页点样式
	InactivePaginationDot lipgloss.Style
	// ArabicPagination 阿拉伯数字分页样式
	ArabicPagination lipgloss.Style
	// DividerDot 分隔点样式
	DividerDot lipgloss.Style
}

// DefaultStyles 返回此列表组件的默认样式定义集。
func DefaultStyles() (s Styles) {
	// verySubduedColor 非常柔和的颜色，用于次要元素
	verySubduedColor := lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
	// subduedColor 柔和的颜色，用于次要文本
	subduedColor := lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

	// 设置标题栏样式，添加底部和左侧内边距
	s.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2) //nolint:mnd

	// 设置标题样式，使用蓝色背景和浅色前景色，添加左右内边距
	s.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	// 设置加载动画样式，使用灰色前景色
	s.Spinner = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#8E8E8E", Dark: "#747373"})

	// 设置过滤提示符样式，使用绿色前景色
	s.FilterPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#ECFD65"})

	// 设置过滤光标样式，使用紫色前景色
	s.FilterCursor = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})

	// 设置过滤器匹配字符的默认样式，添加下划线
	s.DefaultFilterCharacterMatch = lipgloss.NewStyle().Underline(true)

	// 设置状态栏样式，使用灰色前景色，添加底部和左侧内边距
	s.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 1, 2) //nolint:mnd

	// 设置空状态样式，使用柔和的灰色前景色
	s.StatusEmpty = lipgloss.NewStyle().Foreground(subduedColor)

	// 设置激活过滤器时的状态栏样式
	s.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	// 设置过滤器计数样式，使用非常柔和的颜色
	s.StatusBarFilterCount = lipgloss.NewStyle().Foreground(verySubduedColor)

	// 设置无项目时的样式，使用灰色前景色
	s.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"})

	// 设置阿拉伯数字分页样式，使用柔和的灰色前景色
	s.ArabicPagination = lipgloss.NewStyle().Foreground(subduedColor)

	// 设置分页样式，添加左侧内边距
	s.PaginationStyle = lipgloss.NewStyle().PaddingLeft(2) //nolint:mnd

	// 设置帮助样式，添加顶部和左侧内边距
	s.HelpStyle = lipgloss.NewStyle().Padding(1, 0, 0, 2) //nolint:mnd

	// 设置激活的分页点样式，使用灰色前景色，设置为圆点符号
	s.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#847A85", Dark: "#979797"}).
		SetString(bullet)

	// 设置未激活的分页点样式，使用非常柔和的颜色，设置为圆点符号
	s.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(verySubduedColor).
		SetString(bullet)

	// 设置分隔点样式，使用非常柔和的颜色，设置为带空格的圆点符号
	s.DividerDot = lipgloss.NewStyle().
		Foreground(verySubduedColor).
		SetString(" " + bullet + " ")

	return s
}
