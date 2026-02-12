// Package spinner 为 Bubble Tea 应用程序提供一个加载动画组件。
package spinner

import (
	"sync/atomic"
	"time"

	tea "github.com/purpose168/bubbletea-cn"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// 内部 ID 管理。在动画过程中使用，以确保帧消息仅由发送它们的加载动画组件接收。
var lastID int64

// nextID 生成下一个唯一的 ID
func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// Spinner 是一组用于加载动画的帧。
type Spinner struct {
	Frames []string      // 帧序列
	FPS    time.Duration // 帧率（每秒帧数）
}

// 一些可供选择的加载动画。您也可以创建自己的加载动画。
var (
	// Line 线条加载动画
	Line = Spinner{
		Frames: []string{"|", "/", "-", "\\"},
		FPS:    time.Second / 10, //nolint:mnd
	}
	// Dot 点加载动画
	Dot = Spinner{
		Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
		FPS:    time.Second / 10, //nolint:mnd
	}
	// MiniDot 小点加载动画
	MiniDot = Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 12, //nolint:mnd
	}
	// Jump 跳跃加载动画
	Jump = Spinner{
		Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"},
		FPS:    time.Second / 10, //nolint:mnd
	}
	// Pulse 脉冲加载动画
	Pulse = Spinner{
		Frames: []string{"█", "▓", "▒", "░"},
		FPS:    time.Second / 8, //nolint:mnd
	}
	// Points 点加载动画
	Points = Spinner{
		Frames: []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"},
		FPS:    time.Second / 7, //nolint:mnd
	}
	// Globe 地球加载动画
	Globe = Spinner{
		Frames: []string{"🌍", "🌎", "🌏"},
		FPS:    time.Second / 4, //nolint:mnd
	}
	// Moon 月亮加载动画
	Moon = Spinner{
		Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
		FPS:    time.Second / 8, //nolint:mnd
	}
	// Monkey 猴子加载动画
	Monkey = Spinner{
		Frames: []string{"🙈", "🙉", "🙊"},
		FPS:    time.Second / 3, //nolint:mnd
	}
	// Meter 仪表盘加载动画
	Meter = Spinner{
		Frames: []string{
			"▱▱▱",
			"▰▱▱",
			"▰▰▱",
			"▰▰▰",
			"▰▰▱",
			"▰▱▱",
			"▱▱▱",
		},
		FPS: time.Second / 7, //nolint:mnd
	}
	// Hamburger 汉堡加载动画
	Hamburger = Spinner{
		Frames: []string{"☱", "☲", "☴", "☲"},
		FPS:    time.Second / 3, //nolint:mnd
	}
	// Ellipsis 省略号加载动画
	Ellipsis = Spinner{
		Frames: []string{"", ".", "..", "..."},
		FPS:    time.Second / 3, //nolint:mnd
	}
)

// Model 包含加载动画的状态。使用 New 来创建新模型，
// 而不是将 Model 用作结构体字面量。
type Model struct {
	// Spinner 设置。参见类型 Spinner。
	Spinner Spinner

	// Style 设置加载动画的样式。大多数情况下，您只需要
	// 前景色和背景色，以及可能的一些内边距。
	//
	// 有关使用 Lip Gloss 进行样式的介绍，请参阅：
	// https://github.com/charmbracelet/lipgloss
	Style lipgloss.Style

	frame int // 当前帧索引
	id    int // 唯一标识符
	tag   int // 标签，用于防止消息过多
}

// ID 返回加载动画的唯一 ID。
func (m Model) ID() int {
	return m.id
}

// New 返回一个具有默认值的模型。
func New(opts ...Option) Model {
	m := Model{
		Spinner: Line,
		id:      nextID(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// NewModel 返回一个具有默认值的模型。
//
// 已弃用：请改用 [New]。
var NewModel = New

// TickMsg 表示计时器已触发，我们应该渲染一帧。
type TickMsg struct {
	Time time.Time // 触发时间
	tag  int       // 标签
	ID   int       // 加载动画 ID
}

// Update 是 Tea 更新函数。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		// 如果设置了 ID，并且该 ID 不属于此加载动画，则拒绝该消息。
		if msg.ID > 0 && msg.ID != m.id {
			return m, nil
		}

		// 如果设置了标签，并且它不是我们期望的标签，则拒绝该消息。
		// 这可以防止加载动画接收过多消息，从而导致旋转过快。
		if msg.tag > 0 && msg.tag != m.tag {
			return m, nil
		}

		m.frame++
		if m.frame >= len(m.Spinner.Frames) {
			m.frame = 0
		}

		m.tag++
		return m, m.tick(m.id, m.tag)
	default:
		return m, nil
	}
}

// View 渲染模型的视图。
func (m Model) View() string {
	if m.frame >= len(m.Spinner.Frames) {
		return "(error)"
	}

	return m.Style.Render(m.Spinner.Frames[m.frame])
}

// Tick 是用于推进加载动画一帧的命令。使用此命令来有效地启动加载动画。
func (m Model) Tick() tea.Msg {
	return TickMsg{
		// 触发发生的时间。
		Time: time.Now(),

		// 此消息所属的加载动画的 ID。这在路由消息时很有帮助，
		// 但请记住，默认情况下加载动画将忽略不包含 ID 的消息。
		ID: m.id,

		tag: m.tag,
	}
}

func (m Model) tick(id, tag int) tea.Cmd {
	return tea.Tick(m.Spinner.FPS, func(t time.Time) tea.Msg {
		return TickMsg{
			Time: t,
			ID:   id,
			tag:  tag,
		}
	})
}

// Tick 是用于推进加载动画一帧的命令。使用此命令来有效地启动加载动画。
//
// 已弃用：请改用 [Model.Tick]。
func Tick() tea.Msg {
	return TickMsg{Time: time.Now()}
}

// Option 用于在 New 中设置选项。例如：
//
//	spinner := New(WithSpinner(Dot))
type Option func(*Model)

// WithSpinner 是设置加载动画的选项。
func WithSpinner(spinner Spinner) Option {
	return func(m *Model) {
		m.Spinner = spinner
	}
}

// WithStyle 是设置加载动画样式的选项。
func WithStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.Style = style
	}
}
