package viewport

import (
	"math"
	"strings"

	"github.com/purpose168/bubbles-cn/key"
	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// New 创建一个具有给定宽度和高度的视口模型，并设置默认按键映射
func New(width, height int) (m Model) {
	m.Width = width
	m.Height = height
	m.setInitialValues()
	return m
}

// Model 是视口组件的 Bubble Tea 模型
type Model struct {
	Width  int
	Height int
	KeyMap KeyMap

	// MouseWheelEnabled 是否响应鼠标滚轮事件。
	// 必须在 Bubble Tea 中启用鼠标支持才能正常工作。详情请参阅 Bubble Tea 文档。
	MouseWheelEnabled bool

	// MouseWheelDelta 鼠标滚轮滚动的行数。默认为 3
	MouseWheelDelta int

	// YOffset 垂直滚动位置
	YOffset int

	// xOffset 水平滚动位置
	xOffset int

	// horizontalStep 默认水平滚动时左右移动的列数
	horizontalStep int

	// YPosition 视口相对于终端窗口的位置。仅用于高性能渲染
	YPosition int

	// Style 为视口应用 lipgloss 样式。实际上，它最常用于设置边框、边距和内边距
	Style lipgloss.Style

	// HighPerformanceRendering 绕过正常的 Bubble Tea 渲染器，提供更高性能的渲染。
	// 大多数情况下，普通的 Bubble Tea 渲染方法已经足够，但如果你传递的内容包含大量
	// ANSI 转义代码，启用此选项后你可能会在某些终端看到改善的渲染效果。
	//
	// 此选项仅应用于占据整个终端的程序，这通常是通过交替屏幕缓冲区实现的。
	//
	// 已废弃：高性能渲染现已在 Bubble Tea 中被废弃
	HighPerformanceRendering bool

	initialized      bool
	lines            []string
	longestLineWidth int
}

// setInitialValues 设置模型的初始默认值
func (m *Model) setInitialValues() {
	m.KeyMap = DefaultKeyMap()
	m.MouseWheelEnabled = true
	m.MouseWheelDelta = 3
	m.initialized = true
}

// Init 存在是为了满足 tea.Model 接口，以实现组合性
func (m Model) Init() tea.Cmd {
	return nil
}

// AtTop 返回视口是否处于最顶部位置
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom 返回视口是否处于或超过最底部位置
func (m Model) AtBottom() bool {
	return m.YOffset >= m.maxYOffset()
}

// PastBottom 返回视口是否已滚动超过最后一行。
// 这种情况可能在调整视口高度时发生
func (m Model) PastBottom() bool {
	return m.YOffset > m.maxYOffset()
}

// ScrollPercent 返回滚动量作为 0 到 1 之间的浮点数
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(len(m.lines))
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// HorizontalScrollPercent 返回水平滚动量作为 0 到 1 之间的浮点数
func (m Model) HorizontalScrollPercent() float64 {
	if m.xOffset >= m.longestLineWidth-m.Width {
		return 1.0
	}
	y := float64(m.xOffset)
	h := float64(m.Width)
	t := float64(m.longestLineWidth)
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// SetContent 设置分页器的文本内容
func (m *Model) SetContent(s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n") // 规范化行尾
	m.lines = strings.Split(s, "\n")
	m.longestLineWidth = findLongestLineWidth(m.lines)

	if m.YOffset > len(m.lines)-1 {
		m.GotoBottom()
	}
}

// maxYOffset 根据视口的内容和设置的高度返回 y 偏移量的最大可能值
func (m Model) maxYOffset() int {
	return max(0, len(m.lines)-m.Height+m.Style.GetVerticalFrameSize())
}

// visibleLines 返回当前应该在视口中可见的行
func (m Model) visibleLines() (lines []string) {
	h := m.Height - m.Style.GetVerticalFrameSize()
	w := m.Width - m.Style.GetHorizontalFrameSize()

	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := clamp(m.YOffset+h, top, len(m.lines))
		lines = m.lines[top:bottom]
	}

	if (m.xOffset == 0 && m.longestLineWidth <= w) || w == 0 {
		return lines
	}

	cutLines := make([]string, len(lines))
	for i := range lines {
		cutLines[i] = ansi.Cut(lines[i], m.xOffset, m.xOffset+w)
	}
	return cutLines
}

// scrollArea 返回高性能渲染的滚动边界
//
// 已废弃：高性能渲染已在 Bubble Tea 中被废弃
func (m Model) scrollArea() (top, bottom int) {
	top = max(0, m.YPosition)
	bottom = max(top, top+m.Height)
	if top > 0 && bottom > top {
		bottom--
	}
	return top, bottom
}

// SetYOffset 设置 Y 偏移量
func (m *Model) SetYOffset(n int) {
	m.YOffset = clamp(n, 0, m.maxYOffset())
}

// ViewDown 将视图向下移动视口行数的行数。基本上就是"向下翻页"
//
// 已废弃：请改用 [Model.PageDown]
func (m *Model) ViewDown() []string {
	return m.PageDown()
}

// PageDown 将视图向下移动视口行数的行数
func (m *Model) PageDown() []string {
	if m.AtBottom() {
		return nil
	}

	return m.ScrollDown(m.Height)
}

// ViewUp 将视图向上移动一个视口的高度。基本上就是"向上翻页"
//
// 已废弃：请改用 [Model.PageUp]
func (m *Model) ViewUp() []string {
	return m.PageUp()
}

// PageUp 将视图向上移动一个视口的高度
func (m *Model) PageUp() []string {
	if m.AtTop() {
		return nil
	}

	return m.ScrollUp(m.Height)
}

// HalfViewDown 将视图向下移动视口高度的一半
//
// 已废弃：请改用 [Model.HalfPageDown]
func (m *Model) HalfViewDown() (lines []string) {
	return m.HalfPageDown()
}

// HalfPageDown 将视图向下移动视口高度的一半
func (m *Model) HalfPageDown() (lines []string) {
	if m.AtBottom() {
		return nil
	}

	return m.ScrollDown(m.Height / 2) //nolint:mnd
}

// HalfViewUp 将视图向上移动视口高度的一半
//
// 已废弃：请改用 [Model.HalfPageUp]
func (m *Model) HalfViewUp() (lines []string) {
	return m.HalfPageUp()
}

// HalfPageUp 将视图向上移动视口高度的一半
func (m *Model) HalfPageUp() (lines []string) {
	if m.AtTop() {
		return nil
	}

	return m.ScrollUp(m.Height / 2) //nolint:mnd
}

// LineDown 将视图向下移动指定的行数
//
// 已废弃：请改用 [Model.ScrollDown]
func (m *Model) LineDown(n int) (lines []string) {
	return m.ScrollDown(n)
}

// ScrollDown 将视图向下移动指定的行数
func (m *Model) ScrollDown(n int) (lines []string) {
	if m.AtBottom() || n == 0 || len(m.lines) == 0 {
		return nil
	}

	// 确保我们要滚动的行数不大于到达底部之前实际剩余的行数
	m.SetYOffset(m.YOffset + n)

	// 收集用于性能滚动的行
	//
	// XXX：高性能渲染已在 Bubble Tea 中被废弃
	bottom := clamp(m.YOffset+m.Height, 0, len(m.lines))
	top := clamp(m.YOffset+m.Height-n, 0, bottom)
	return m.lines[top:bottom]
}

// LineUp 将视图向下移动指定的行数。返回要显示的新行
//
// 已废弃：请改用 [Model.ScrollUp]
func (m *Model) LineUp(n int) (lines []string) {
	return m.ScrollUp(n)
}

// ScrollUp 将视图向下移动指定的行数。返回要显示的新行
func (m *Model) ScrollUp(n int) (lines []string) {
	if m.AtTop() || n == 0 || len(m.lines) == 0 {
		return nil
	}

	// 确保我们要滚动的行数不大于距离顶部的行数
	m.SetYOffset(m.YOffset - n)

	// 收集用于性能滚动的行
	//
	// XXX：高性能渲染已在 Bubble Tea 中被废弃
	top := max(0, m.YOffset)
	bottom := clamp(m.YOffset+n, 0, m.maxYOffset())
	return m.lines[top:bottom]
}

// SetHorizontalStep 设置使用默认视口按键映射时左右滚动的默认列数
//
// 如果设置为 0 或更小，水平滚动将被禁用
//
// 在 v1 版本中，水平滚动默认是禁用的
func (m *Model) SetHorizontalStep(n int) {
	m.horizontalStep = max(n, 0)
}

// SetXOffset 设置 X 偏移量
func (m *Model) SetXOffset(n int) {
	m.xOffset = clamp(n, 0, m.longestLineWidth-m.Width)
}

// ScrollLeft 将视口向左移动指定的列数
func (m *Model) ScrollLeft(n int) {
	m.SetXOffset(m.xOffset - n)
}

// ScrollRight 将视口向右移动指定的列数
func (m *Model) ScrollRight(n int) {
	m.SetXOffset(m.xOffset + n)
}

// TotalLineCount 返回视口内行的总数（包括隐藏和可见的行）
func (m Model) TotalLineCount() int {
	return len(m.lines)
}

// VisibleLineCount 返回视口内可见行的数量
func (m Model) VisibleLineCount() int {
	return len(m.visibleLines())
}

// GotoTop 将视口设置到顶部位置
func (m *Model) GotoTop() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.SetYOffset(0)
	return m.visibleLines()
}

// GotoBottom 将视口设置到底部位置
func (m *Model) GotoBottom() (lines []string) {
	m.SetYOffset(m.maxYOffset())
	return m.visibleLines()
}

// Sync 告诉渲染器视口将位于何处，并请求渲染视口的当前状态。
// 它应该在第一次渲染和窗口调整大小后调用
//
// 仅用于高性能渲染
//
// 已废弃：高性能渲染已在 Bubble Tea 中被废弃
func Sync(m Model) tea.Cmd {
	if len(m.lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()
	return tea.SyncScrollArea(m.visibleLines(), top, bottom)
}

// ViewDown 是一个高性能命令，将视口向上移动指定的行数。
// 使用 Model.ViewDown 获取应该渲染的行。例如：
//
//	lines := model.ViewDown(1)
//	cmd := ViewDown(m, lines)
//
// 已废弃：高性能渲染已在 Bubble Tea 中被废弃
func ViewDown(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()

	// XXX：高性能渲染已在 Bubble Tea 中被废弃。在 v2 版本中，
	// 我们不需要在这里返回命令
	return tea.ScrollDown(lines, top, bottom)
}

// ViewUp 是一个高性能命令，将视口向下移动指定的高度行数。
// 使用 Model.ViewUp 获取应该渲染的行
//
// 已废弃：高性能渲染已在 Bubble Tea 中被废弃
func ViewUp(m Model, lines []string) tea.Cmd {
	if len(lines) == 0 {
		return nil
	}
	top, bottom := m.scrollArea()

	// XXX：高性能渲染已在 Bubble Tea 中被废弃。在 v2 版本中，
	// 我们不需要在这里返回命令
	return tea.ScrollUp(lines, top, bottom)
}

// Update 处理基于消息的标准视口更新
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m, cmd = m.updateAsModel(msg)
	return m, cmd
}

// updateAsModel 此方法被分离出来，以便更容易地将 Update 转换为满足 tea.Model
func (m Model) updateAsModel(msg tea.Msg) (Model, tea.Cmd) {
	if !m.initialized {
		m.setInitialValues()
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.PageDown):
			lines := m.PageDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		case key.Matches(msg, m.KeyMap.PageUp):
			lines := m.PageUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			lines := m.HalfPageDown()
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		case key.Matches(msg, m.KeyMap.HalfPageUp):
			lines := m.HalfPageUp()
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case key.Matches(msg, m.KeyMap.Down):
			lines := m.ScrollDown(1)
			if m.HighPerformanceRendering {
				cmd = ViewDown(m, lines)
			}

		case key.Matches(msg, m.KeyMap.Up):
			lines := m.ScrollUp(1)
			if m.HighPerformanceRendering {
				cmd = ViewUp(m, lines)
			}

		case key.Matches(msg, m.KeyMap.Left):
			m.ScrollLeft(m.horizontalStep)

		case key.Matches(msg, m.KeyMap.Right):
			m.ScrollRight(m.horizontalStep)
		}

	case tea.MouseMsg:
		if !m.MouseWheelEnabled || msg.Action != tea.MouseActionPress {
			break
		}
		switch msg.Button { //nolint:exhaustive
		case tea.MouseButtonWheelUp:
			if msg.Shift {
				// 注意：并非每个终端模拟器默认都发送鼠标动作的 Shift 事件（看看你，Konsole）
				m.ScrollLeft(m.horizontalStep)
			} else {
				lines := m.ScrollUp(m.MouseWheelDelta)
				if m.HighPerformanceRendering {
					cmd = ViewUp(m, lines)
				}
			}

		case tea.MouseButtonWheelDown:
			if msg.Shift {
				m.ScrollRight(m.horizontalStep)
			} else {
				lines := m.ScrollDown(m.MouseWheelDelta)
				if m.HighPerformanceRendering {
					cmd = ViewDown(m, lines)
				}
			}
		// 注意：并非每个终端模拟器默认都发送水平滚轮事件（看看你，Konsole）
		case tea.MouseButtonWheelLeft:
			m.ScrollLeft(m.horizontalStep)
		case tea.MouseButtonWheelRight:
			m.ScrollRight(m.horizontalStep)
		}
	}

	return m, cmd
}

// View 将视口渲染为字符串
func (m Model) View() string {
	if m.HighPerformanceRendering {
		// 由于我们将单独渲染实际内容，只需发送换行符。
		// 我们仍然需要发送一些等于此视图高度的内容，
		// 以便 Bubble Tea 标准渲染器正确定位此视图下方的任何内容
		return strings.Repeat("\n", max(0, m.Height-1))
	}

	w, h := m.Width, m.Height
	if sw := m.Style.GetWidth(); sw != 0 {
		w = min(w, sw)
	}
	if sh := m.Style.GetHeight(); sh != 0 {
		h = min(h, sh)
	}
	contentWidth := w - m.Style.GetHorizontalFrameSize()
	contentHeight := h - m.Style.GetVerticalFrameSize()
	contents := lipgloss.NewStyle().
		Width(contentWidth).      // 填充到宽度
		Height(contentHeight).    // 填充到高度
		MaxHeight(contentHeight). // 如果更高则截断高度
		MaxWidth(contentWidth).   // 如果更宽则截断宽度
		Render(strings.Join(m.visibleLines(), "\n"))
	return m.Style.
		UnsetWidth().UnsetHeight(). // 样式大小已在 contents 中应用
		Render(contents)
}

// clamp 将值限制在指定的最小值和最大值之间
func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

// findLongestLineWidth 查找所有行中最长行的宽度
func findLongestLineWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if ww := ansi.StringWidth(l); ww > w {
			w = ww
		}
	}
	return w
}
