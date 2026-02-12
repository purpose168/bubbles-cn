// Package paginator 提供一个 Bubble Tea 包，用于计算分页
// 和渲染分页信息。请注意，此包不渲染实际页面：
// 它纯粹用于处理与分页相关的按键操作，以及渲染分页状态。
package paginator

import (
	"fmt"

	"github.com/purpose168/bubbles-cn/key"
	tea "github.com/purpose168/bubbletea-cn"
)

// Type 指定我们渲染分页的方式。
type Type int

// 分页渲染选项。
const (
	// Arabic 阿拉伯数字分页方式
	Arabic Type = iota
	// Dots 圆点分页方式
	Dots
)

// KeyMap 是分页器中不同操作的按键绑定。
type KeyMap struct {
	// PrevPage 上一页按键绑定
	PrevPage key.Binding
	// NextPage 下一页按键绑定
	NextPage key.Binding
}

// DefaultKeyMap 是用于导航和操作分页器的默认按键绑定集。
var DefaultKeyMap = KeyMap{
	PrevPage: key.NewBinding(key.WithKeys("pgup", "left", "h")),
	NextPage: key.NewBinding(key.WithKeys("pgdown", "right", "l")),
}

// Model 是此用户界面的 Bubble Tea 模型。
type Model struct {
	// Type 配置分页的渲染方式（阿拉伯数字、圆点）。
	Type Type
	// Page 是当前页码。
	Page int
	// PerPage 是每页的项目数量。
	PerPage int
	// TotalPages 是总页数。
	TotalPages int
	// ActiveDot 用于在圆点显示类型下标记当前页面。
	ActiveDot string
	// InactiveDot 用于在圆点显示类型下标记非活动页面。
	InactiveDot string
	// ArabicFormat 是用于阿拉伯数字显示类型的 printf 风格格式字符串。
	ArabicFormat string

	// KeyMap 编码小部件识别的按键绑定。
	KeyMap KeyMap

	// Deprecated: 请改为自定义 [KeyMap]。
	UsePgUpPgDownKeys bool
	// Deprecated: 请改为自定义 [KeyMap]。
	UseLeftRightKeys bool
	// Deprecated: 请改为自定义 [KeyMap]。
	UseUpDownKeys bool
	// Deprecated: 请改为自定义 [KeyMap]。
	UseHLKeys bool
	// Deprecated: 请改为自定义 [KeyMap]。
	UseJKKeys bool
}

// SetTotalPages 是一个辅助函数，用于从给定的项目数量计算总页数。
// 其使用是可选的，因为此分页器可用于导航集合之外的其他用途。
// 请注意，它既返回总页数，又修改模型。
func (m *Model) SetTotalPages(items int) int {
	if items < 1 {
		return m.TotalPages
	}
	n := items / m.PerPage
	if items%m.PerPage > 0 {
		n++
	}
	m.TotalPages = n
	return n
}

// ItemsOnPage 是一个辅助函数，用于返回当前页面上的项目数量，
// 参数为传入的总项目数。
func (m Model) ItemsOnPage(totalItems int) int {
	if totalItems < 1 {
		return 0
	}
	start, end := m.GetSliceBounds(totalItems)
	return end - start
}

// GetSliceBounds 是一个用于分页切片的辅助函数。
// 传入您正在渲染的切片的长度，您将收到与分页对应的起始和结束边界。
// 例如：
//
//	bunchOfStuff := []stuff{...}
//	start, end := model.GetSliceBounds(len(bunchOfStuff))
//	sliceToRender := bunchOfStuff[start:end]
func (m *Model) GetSliceBounds(length int) (start int, end int) {
	start = m.Page * m.PerPage
	end = min(m.Page*m.PerPage+m.PerPage, length)
	return start, end
}

// PrevPage 是一个辅助函数，用于向后导航一页。
// 它不会翻到第一页之前（即第 0 页）。
func (m *Model) PrevPage() {
	if m.Page > 0 {
		m.Page--
	}
}

// NextPage 是一个辅助函数，用于向前导航一页。
// 它不会翻到最后一页之后（即 totalPages - 1）。
func (m *Model) NextPage() {
	if !m.OnLastPage() {
		m.Page++
	}
}

// OnLastPage 返回我们是否在最后一页。
func (m Model) OnLastPage() bool {
	return m.Page == m.TotalPages-1
}

// OnFirstPage 返回我们是否在第一页。
func (m Model) OnFirstPage() bool {
	return m.Page == 0
}

// Option 用于在 New 中设置选项。
type Option func(*Model)

// New 创建一个带有默认值的新模型。
func New(opts ...Option) Model {
	m := Model{
		Type:         Arabic,
		Page:         0,
		PerPage:      1,
		TotalPages:   1,
		KeyMap:       DefaultKeyMap,
		ActiveDot:    "•",
		InactiveDot:  "○",
		ArabicFormat: "%d/%d",
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// NewModel 创建一个带有默认值的新模型。
//
// Deprecated: 请改用 [New]。
var NewModel = New

// WithTotalPages 设置总页数。
func WithTotalPages(totalPages int) Option {
	return func(m *Model) {
		m.TotalPages = totalPages
	}
}

// WithPerPage 设置每页项目数。
func WithPerPage(perPage int) Option {
	return func(m *Model) {
		m.PerPage = perPage
	}
}

// Update 是 Tea 更新函数，将按键绑定到分页操作。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.NextPage):
			m.NextPage()
		case key.Matches(msg, m.KeyMap.PrevPage):
			m.PrevPage()
		}
	}

	return m, nil
}

// View 将分页渲染为字符串。
func (m Model) View() string {
	switch m.Type { //nolint:exhaustive
	case Dots:
		return m.dotsView()
	default:
		return m.arabicView()
	}
}

// dotsView 渲染圆点分页视图
func (m Model) dotsView() string {
	var s string
	for i := 0; i < m.TotalPages; i++ {
		if i == m.Page {
			s += m.ActiveDot
			continue
		}
		s += m.InactiveDot
	}
	return s
}

// arabicView 渲染阿拉伯数字分页视图
func (m Model) arabicView() string {
	return fmt.Sprintf(m.ArabicFormat, m.Page+1, m.TotalPages)
}
