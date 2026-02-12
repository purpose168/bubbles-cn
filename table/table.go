// Package table 为 Bubble Tea 应用程序提供一个简单的表格组件。
package table

import (
	"strings"

	"github.com/mattn/go-runewidth"
	tea "github.com/purpose168/bubbletea-cn"
	lipgloss "github.com/purpose168/lipgloss-cn"

	"github.com/purpose168/bubbles-cn/help"
	"github.com/purpose168/bubbles-cn/key"
	"github.com/purpose168/bubbles-cn/viewport"
)

// Model 定义表格小部件的状态。
type Model struct {
	KeyMap KeyMap     // 键位映射
	Help   help.Model // 帮助模型

	cols   []Column // 列定义
	rows   []Row    // 行数据
	cursor int      // 光标位置
	focus  bool     // 是否聚焦
	styles Styles   // 样式

	viewport viewport.Model // 视口
	start    int            // 起始行
	end      int            // 结束行
}

// Row 表示表格中的一行。
type Row []string

// Column 定义表格结构。
type Column struct {
	Title string // 列标题
	Width int    // 列宽度
}

// KeyMap 定义键绑定。它满足 help.KeyMap 接口，
// 该接口用于渲染帮助菜单。
type KeyMap struct {
	LineUp       key.Binding // 向上移动一行
	LineDown     key.Binding // 向下移动一行
	PageUp       key.Binding // 向上翻页
	PageDown     key.Binding // 向下翻页
	HalfPageUp   key.Binding // 向上翻半页
	HalfPageDown key.Binding // 向下翻半页
	GotoTop      key.Binding // 跳转到顶部
	GotoBottom   key.Binding // 跳转到底部
}

// ShortHelp 实现 KeyMap 接口。
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown}
}

// FullHelp 实现 KeyMap 接口。
func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.LineUp, km.LineDown, km.GotoTop, km.GotoBottom},
		{km.PageUp, km.PageDown, km.HalfPageUp, km.HalfPageDown},
	}
}

// DefaultKeyMap 返回默认的键绑定集合。
func DefaultKeyMap() KeyMap {
	const spacebar = " "
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", spacebar),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
	}
}

// Styles 包含此列表组件的样式定义。默认情况下，
// 这些值由 DefaultStyles 生成。
type Styles struct {
	Header   lipgloss.Style // 表头样式
	Cell     lipgloss.Style // 单元格样式
	Selected lipgloss.Style // 选中样式
}

// DefaultStyles 返回此表格的默认样式定义集合。
func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

// SetStyles 设置表格样式。
func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.UpdateViewport()
}

// Option 用于在 New 中设置选项。例如：
//
//	table := New(WithColumns([]Column{{Title: "ID", Width: 10}}))
type Option func(*Model)

// New 为表格小部件创建一个新模型。
func New(opts ...Option) Model {
	m := Model{
		cursor:   0,
		viewport: viewport.New(0, 20), //nolint:mnd

		KeyMap: DefaultKeyMap(),
		Help:   help.New(),
		styles: DefaultStyles(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	m.UpdateViewport()

	return m
}

// WithColumns 设置表格列（表头）。
func WithColumns(cols []Column) Option {
	return func(m *Model) {
		m.cols = cols
	}
}

// WithRows 设置表格行（数据）。
func WithRows(rows []Row) Option {
	return func(m *Model) {
		m.rows = rows
	}
}

// WithHeight 设置表格的高度。
func WithHeight(h int) Option {
	return func(m *Model) {
		m.viewport.Height = h - lipgloss.Height(m.headersView())
	}
}

// WithWidth 设置表格的宽度。
func WithWidth(w int) Option {
	return func(m *Model) {
		m.viewport.Width = w
	}
}

// WithFocused 设置表格的聚焦状态。
func WithFocused(f bool) Option {
	return func(m *Model) {
		m.focus = f
	}
}

// WithStyles 设置表格样式。
func WithStyles(s Styles) Option {
	return func(m *Model) {
		m.styles = s
	}
}

// WithKeyMap 设置键映射。
func WithKeyMap(km KeyMap) Option {
	return func(m *Model) {
		m.KeyMap = km
	}
}

// Update 是 Bubble Tea 更新循环。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2) //nolint:mnd
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2) //nolint:mnd
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		}
	}

	return m, nil
}

// Focused 返回表格的聚焦状态。
func (m Model) Focused() bool {
	return m.focus
}

// Focus 聚焦表格，允许用户在行之间移动并
// 进行交互。
func (m *Model) Focus() {
	m.focus = true
	m.UpdateViewport()
}

// Blur 模糊表格，防止选择或移动。
func (m *Model) Blur() {
	m.focus = false
	m.UpdateViewport()
}

// View 渲染组件。
func (m Model) View() string {
	return m.headersView() + "\n" + m.viewport.View()
}

// HelpView 是从键映射渲染帮助菜单的辅助方法。
// 请注意，默认情况下不会渲染此视图，您必须在应用程序中
// 手动调用它（如果适用）。
func (m Model) HelpView() string {
	return m.Help.View(m.KeyMap)
}

// UpdateViewport 根据先前定义的列和行更新列表内容。
func (m *Model) UpdateViewport() {
	renderedRows := make([]string, 0, len(m.rows))

	// 仅渲染从 m.cursor-m.viewport.Height 到 m.cursor+m.viewport.Height 的行
	// 恒定运行时，独立于表格中的行数
	// 将 renderedRows 的数量限制为最多 2*m.viewport.Height
	if m.cursor >= 0 {
		m.start = clamp(m.cursor-m.viewport.Height, 0, m.cursor)
	} else {
		m.start = 0
	}
	m.end = clamp(m.cursor+m.viewport.Height, m.cursor, len(m.rows))
	for i := m.start; i < m.end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// SelectedRow 返回选中的行。
// 您可以将其转换为您自己的实现。
func (m Model) SelectedRow() Row {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil
	}

	return m.rows[m.cursor]
}

// Rows 返回当前行。
func (m Model) Rows() []Row {
	return m.rows
}

// Columns 返回当前列。
func (m Model) Columns() []Column {
	return m.cols
}

// SetRows 设置新的行状态。
func (m *Model) SetRows(r []Row) {
	m.rows = r

	if m.cursor > len(m.rows)-1 {
		m.cursor = len(m.rows) - 1
	}

	m.UpdateViewport()
}

// SetColumns 设置新的列状态。
func (m *Model) SetColumns(c []Column) {
	m.cols = c
	m.UpdateViewport()
}

// SetWidth 设置表格视口的宽度。
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

// SetHeight 设置表格视口的高度。
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
	m.UpdateViewport()
}

// Height 返回表格视口的高度。
func (m Model) Height() int {
	return m.viewport.Height
}

// Width 返回表格视口的宽度。
func (m Model) Width() int {
	return m.viewport.Width
}

// Cursor 返回选中行的索引。
func (m Model) Cursor() int {
	return m.cursor
}

// SetCursor 设置表格中的光标位置。
func (m *Model) SetCursor(n int) {
	m.cursor = clamp(n, 0, len(m.rows)-1)
	m.UpdateViewport()
}

// MoveUp 将选择向上移动任意行数。
// 它不能超过第一行。
func (m *Model) MoveUp(n int) {
	m.cursor = clamp(m.cursor-n, 0, len(m.rows)-1)
	switch {
	case m.start == 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset, 0, m.cursor))
	case m.start < m.viewport.Height:
		m.viewport.YOffset = (clamp(clamp(m.viewport.YOffset+n, 0, m.cursor), 0, m.viewport.Height))
	case m.viewport.YOffset >= 1:
		m.viewport.YOffset = clamp(m.viewport.YOffset+n, 1, m.viewport.Height)
	}
	m.UpdateViewport()
}

// MoveDown 将选择向下移动任意行数。
// 它不能低于最后一行。
func (m *Model) MoveDown(n int) {
	m.cursor = clamp(m.cursor+n, 0, len(m.rows)-1)
	m.UpdateViewport()

	switch {
	case m.end == len(m.rows) && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.viewport.Height))
	case m.cursor > (m.end-m.start)/2 && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset-n, 1, m.cursor))
	case m.viewport.YOffset > 1:
	case m.cursor > m.viewport.YOffset+m.viewport.Height-1:
		m.viewport.SetYOffset(clamp(m.viewport.YOffset+1, 0, 1))
	}
}

// GotoTop 将选择移动到第一行。
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom 将选择移动到最后一行。
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.rows))
}

// FromValues 从简单字符串创建表格行。默认情况下，它使用 `\n`
// 来获取所有行，并使用给定的分隔符来分隔每行的字段。
func (m *Model) FromValues(value, separator string) {
	rows := []Row{}
	for _, line := range strings.Split(value, "\n") {
		r := Row{}
		for _, field := range strings.Split(line, separator) {
			r = append(r, field)
		}
		rows = append(rows, r)
	}

	m.SetRows(rows)
}

func (m Model) headersView() string {
	s := make([]string, 0, len(m.cols))
	for _, col := range m.cols {
		if col.Width <= 0 {
			continue
		}
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col.Title, col.Width, "…"))
		s = append(s, m.styles.Header.Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, s...)
}

func (m *Model) renderRow(r int) string {
	s := make([]string, 0, len(m.cols))
	for i, value := range m.rows[r] {
		if m.cols[i].Width <= 0 {
			continue
		}
		style := lipgloss.NewStyle().Width(m.cols[i].Width).MaxWidth(m.cols[i].Width).Inline(true)
		renderedCell := m.styles.Cell.Render(style.Render(runewidth.Truncate(value, m.cols[i].Width, "…")))
		s = append(s, renderedCell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, s...)

	if r == m.cursor {
		return m.styles.Selected.Render(row)
	}

	return row
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
