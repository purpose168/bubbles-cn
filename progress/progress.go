// Package progress 为 Bubble Tea 应用程序提供简单的进度条。
package progress

import (
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/harmonica"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// 内部 ID 管理。用于动画期间确保帧消息只能由发送它们的进度组件接收。
var lastID int64

// nextID 生成下一个唯一的 ID
func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

const (
	fps              = 60      // 帧率
	defaultWidth     = 40      // 默认宽度
	defaultFrequency = 18.0    // 默认频率
	defaultDamping   = 1.0     // 默认阻尼
)

// Option 用于在 New 中设置选项。例如：
//
//	    progress := New(
//		       WithRamp("#ff0000", "#0000ff"),
//		       WithoutPercentage(),
//	    )
type Option func(*Model)

// WithDefaultGradient 设置使用默认颜色的渐变填充。
func WithDefaultGradient() Option {
	return WithGradient("#5A56E0", "#EE6FF8")
}

// WithGradient 设置在两种颜色之间混合的渐变填充。
func WithGradient(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, false)
	}
}

// WithDefaultScaledGradient 设置使用默认颜色的渐变，并缩放渐变以适应填充的渐变部分。
func WithDefaultScaledGradient() Option {
	return WithScaledGradient("#5A56E0", "#EE6FF8")
}

// WithScaledGradient 缩放渐变以适应进度条填充部分的宽度。
func WithScaledGradient(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, true)
	}
}

// WithSolidFill 设置进度条使用给定颜色的纯色填充。
func WithSolidFill(color string) Option {
	return func(m *Model) {
		m.FullColor = color
		m.useRamp = false
	}
}

// WithFillCharacters 设置用于构建进度条完整和空部分的字符。
func WithFillCharacters(full rune, empty rune) Option {
	return func(m *Model) {
		m.Full = full
		m.Empty = empty
	}
}

// WithoutPercentage 隐藏数字百分比。
func WithoutPercentage() Option {
	return func(m *Model) {
		m.ShowPercentage = false
	}
}

// WithWidth 设置进度条的初始宽度。请注意，您也可以通过 Width 属性设置宽度，
// 如果您正在等待 tea.WindowSizeMsg，这会很方便。
func WithWidth(w int) Option {
	return func(m *Model) {
		m.Width = w
	}
}

// WithSpringOptions 设置进度条内置基于弹簧动画的初始频率和阻尼选项。
// 频率对应速度，阻尼对应弹性。详细信息请参阅：
//
// https://github.com/charmbracelet/harmonica
func WithSpringOptions(frequency, damping float64) Option {
	return func(m *Model) {
		m.SetSpringOptions(frequency, damping)
		m.springCustomized = true
	}
}

// WithColorProfile 设置进度条使用的颜色配置文件。
func WithColorProfile(p termenv.Profile) Option {
	return func(m *Model) {
		m.colorProfile = p
	}
}

// FrameMsg 指示应该发生动画步骤。
type FrameMsg struct {
	id  int // 进度条 ID
	tag int // 标签，用于防止接收帧消息过快
}

// Model 存储我们在渲染进度条时将使用的值。
type Model struct {
	// 一个标识符，防止我们接收其他进度条的消息。
	id int

	// 一个标识符，防止我们过快地接收帧消息。
	tag int

	// 进度条的总宽度，包括百分比（如果设置）。
	Width int

	// 进度条的"已填充"部分。
	Full      rune   // 填充字符
	FullColor string // 填充颜色

	// 进度条的"空"部分。
	Empty      rune   // 空字符
	EmptyColor string // 空颜色

	// 渲染数字百分比的设置。
	ShowPercentage  bool            // 是否显示百分比
	PercentFormat   string          // 浮点数的格式字符串
	PercentageStyle lipgloss.Style  // 百分比样式

	// 动画过渡的成员。
	spring           harmonica.Spring // 弹簧对象
	springCustomized bool            // 弹簧是否已自定义
	percentShown     float64         // 当前显示的百分比
	targetPercent    float64         // 我们正在动画化的目标百分比
	velocity         float64         // 速度

	// 渐变设置
	useRamp    bool            // 是否使用渐变
	rampColorA colorful.Color  // 渐变起始颜色
	rampColorB colorful.Color  // 渐变结束颜色

	// 当为 true 时，我们缩放渐变以适应进度条填充部分的宽度。
	// 当为 false 时，渐变的宽度将设置为进度条的全宽。
	scaleRamp bool

	// 进度条的颜色配置文件。
	colorProfile termenv.Profile
}

// New 返回一个带有默认值的模型。
func New(opts ...Option) Model {
	m := Model{
		id:             nextID(),
		Width:          defaultWidth,
		Full:           '█',
		FullColor:      "#7571F9",
		Empty:          '░',
		EmptyColor:     "#606060",
		ShowPercentage: true,
		PercentFormat:  " %3.0f%%",
		colorProfile:   termenv.ColorProfile(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	if !m.springCustomized {
		m.SetSpringOptions(defaultFrequency, defaultDamping)
	}

	return m
}

// NewModel 返回一个带有默认值的模型。
//
// Deprecated: 请改用 [New]。
var NewModel = New

// Init 存在以满足 tea.Model 接口。
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 用于在过渡期间动画化进度条。使用 SetPercent 创建触发动画所需的命令。
//
// 如果您使用 ViewAs 渲染，则不需要此功能。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}

		// 如果我们已或多或少达到平衡，则停止更新。
		if !m.IsAnimating() {
			return m, nil
		}

		m.percentShown, m.velocity = m.spring.Update(m.percentShown, m.velocity, m.targetPercent)
		return m, m.nextFrame()

	default:
		return m, nil
	}
}

// SetSpringOptions 设置当前弹簧的频率和阻尼。
// 频率对应速度，阻尼对应弹性。详细信息请参阅：
//
// https://github.com/charmbracelet/harmonica
func (m *Model) SetSpringOptions(frequency, damping float64) {
	m.spring = harmonica.NewSpring(harmonica.FPS(fps), frequency, damping)
}

// Percent 返回模型上当前可见的百分比。这仅在您动画化进度条时相关。
//
// 如果您使用 ViewAs 渲染，则不需要此功能。
func (m Model) Percent() float64 {
	return m.targetPercent
}

// SetPercent 设置模型的百分比状态以及将进度条动画化到此新百分比所需的命令。
//
// 如果您使用 ViewAs 渲染，则不需要此功能。
func (m *Model) SetPercent(p float64) tea.Cmd {
	m.targetPercent = math.Max(0, math.Min(1, p))
	m.tag++
	return m.nextFrame()
}

// IncrPercent 按给定量增加百分比，返回将进度条动画化到新百分比所需的命令。
//
// 如果您使用 ViewAs 渲染，则不需要此功能。
func (m *Model) IncrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() + v)
}

// DecrPercent 按给定量减少百分比，返回将进度条动画化到新百分比所需的命令。
//
// 如果您使用 ViewAs 渲染，则不需要此功能。
func (m *Model) DecrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() - v)
}

// View 在其当前状态下渲染动画进度条。要基于您自己的计算渲染静态进度条，请改用 ViewAs。
func (m Model) View() string {
	return m.ViewAs(m.percentShown)
}

// ViewAs 使用给定的百分比渲染进度条。
func (m Model) ViewAs(percent float64) string {
	b := strings.Builder{}
	percentView := m.percentageView(percent)
	m.barView(&b, percent, ansi.StringWidth(percentView))
	b.WriteString(percentView)
	return b.String()
}

// nextFrame 生成下一帧动画的命令
func (m *Model) nextFrame() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(time.Time) tea.Msg {
		return FrameMsg{id: m.id, tag: m.tag}
	})
}

// barView 渲染进度条
func (m Model) barView(b *strings.Builder, percent float64, textWidth int) {
	var (
		tw = max(0, m.Width-textWidth)                // 总宽度
		fw = int(math.Round((float64(tw) * percent))) // 填充宽度
		p  float64                                    // 渐变位置
	)

	fw = max(0, min(tw, fw))

	if m.useRamp {
		// 渐变填充
		for i := 0; i < fw; i++ {
			if fw == 1 {
				// 这有待商榷：在宽度=1 的渐变中，单个渲染的字符应该是
				// 第一种颜色、最后一种颜色还是正好在中间 50%？我选择了 50%
				p = 0.5
			} else if m.scaleRamp {
				p = float64(i) / float64(fw-1)
			} else {
				p = float64(i) / float64(tw-1)
			}
			c := m.rampColorA.BlendLuv(m.rampColorB, p).Hex()
			b.WriteString(termenv.
				String(string(m.Full)).
				Foreground(m.color(c)).
				String(),
			)
		}
	} else {
		// 纯色填充
		s := termenv.String(string(m.Full)).Foreground(m.color(m.FullColor)).String()
		b.WriteString(strings.Repeat(s, fw))
	}

	// 空填充
	e := termenv.String(string(m.Empty)).Foreground(m.color(m.EmptyColor)).String()
	n := max(0, tw-fw)
	b.WriteString(strings.Repeat(e, n))
}

// percentageView 渲染百分比视图
func (m Model) percentageView(percent float64) string {
	if !m.ShowPercentage {
		return ""
	}
	percent = math.Max(0, math.Min(1, percent))
	percentage := fmt.Sprintf(m.PercentFormat, percent*100) //nolint:mnd
	percentage = m.PercentageStyle.Inline(true).Render(percentage)
	return percentage
}

// setRamp 设置渐变颜色
func (m *Model) setRamp(colorA, colorB string, scaled bool) {
	// 如果出现错误，这里的颜色将默认为黑色。为了可用性的缘故，
	// 并且因为这样的错误只是美观问题，我们忽略错误。
	a, _ := colorful.Hex(colorA)
	b, _ := colorful.Hex(colorB)

	m.useRamp = true
	m.scaleRamp = scaled
	m.rampColorA = a
	m.rampColorB = b
}

// color 返回颜色对象
func (m Model) color(c string) termenv.Color {
	return m.colorProfile.Color(c)
}

// IsAnimating 如果进度条达到平衡并且不再动画化，则返回 false。
func (m *Model) IsAnimating() bool {
	dist := math.Abs(m.percentShown - m.targetPercent)
	return !(dist < 0.001 && m.velocity < 0.01)
}
