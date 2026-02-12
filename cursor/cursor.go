// Package cursor 为 Bubble Tea 应用程序提供光标功能。
package cursor

import (
	"context"
	"time"

	tea "github.com/purpose168/bubbletea-cn"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

const defaultBlinkSpeed = time.Millisecond * 530

// initialBlinkMsg 初始化光标闪烁。
type initialBlinkMsg struct{}

// BlinkMsg 信号表示光标应该闪烁。它包含元数据，
// 允许我们判断闪烁消息是否是我们期望的。
type BlinkMsg struct {
	id  int
	tag int
}

// blinkCanceled 在闪烁操作被取消时发送。
type blinkCanceled struct{}

// blinkCtx 管理光标闪烁。
type blinkCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Mode 描述光标的行为。
type Mode int

// 可用的光标模式。
const (
	CursorBlink  Mode = iota // 光标闪烁
	CursorStatic             // 光标静态
	CursorHide               // 光标隐藏
)

// String 返回人类可读格式的光标模式。此方法是
// 临时的，仅用于信息目的。
func (c Mode) String() string {
	return [...]string{
		"blink",
		"static",
		"hidden",
	}[c]
}

// Model 是此光标元素的 Bubble Tea 模型。
type Model struct {
	BlinkSpeed time.Duration
	// Style 用于设置光标块的样式。
	Style lipgloss.Style
	// TextStyle 是光标隐藏时（闪烁时）使用的样式。
	// 即显示正常文本。
	TextStyle lipgloss.Style

	// char 是光标下的字符
	char string
	// id 是此 Model 与其他光标的关联 ID
	id int
	// focus 表示包含的输入是否被聚焦
	focus bool
	// Blink 光标闪烁状态。
	Blink bool
	// blinkCtx 用于管理光标闪烁
	blinkCtx *blinkCtx
	// blinkTag 是我们期望接收的闪烁消息的 ID。
	blinkTag int
	// mode 决定光标的行为
	mode Mode
}

// New 创建一个具有默认设置的新模型。
func New() Model {
	return Model{
		BlinkSpeed: defaultBlinkSpeed, // 设置默认闪烁速度

		Blink: true,        // 初始闪烁状态为 true
		mode:  CursorBlink, // 初始模式为闪烁

		blinkCtx: &blinkCtx{
			ctx: context.Background(), // 创建背景上下文
		},
	}
}

// Update 更新光标状态。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case initialBlinkMsg:
		// 我们接受所有由 Blink 命令生成的 initialBlinkMsg。

		if m.mode != CursorBlink || !m.focus {
			return m, nil
		}

		cmd := m.BlinkCmd()
		return m, cmd

	case tea.FocusMsg:
		return m, m.Focus()

	case tea.BlurMsg:
		m.Blur()
		return m, nil

	case BlinkMsg:
		// 我们对是否接受 blinkMsg 很挑剔，以便光标
		// 只在应该闪烁的时候闪烁。

		// 此模型是否可闪烁？
		if m.mode != CursorBlink || !m.focus {
			return m, nil
		}

		// 这是我们期望的闪烁消息吗？
		if msg.id != m.id || msg.tag != m.blinkTag {
			return m, nil
		}

		var cmd tea.Cmd
		if m.mode == CursorBlink {
			m.Blink = !m.Blink // 切换闪烁状态
			cmd = m.BlinkCmd() // 继续闪烁
		}
		return m, cmd

	case blinkCanceled: // no-op
		return m, nil
	}
	return m, nil
}

// Mode 返回模型的光标模式。有关可用的光标模式，请参见
// type Mode。
func (m Model) Mode() Mode {
	return m.mode
}

// SetMode 设置模型的光标模式。此方法返回一个命令。
//
// 有关可用的光标模式，请参见 type Mode。
func (m *Model) SetMode(mode Mode) tea.Cmd {
	// 如果模式值超出范围，则调整模式值
	if mode < CursorBlink || mode > CursorHide {
		return nil
	}
	m.mode = mode
	m.Blink = m.mode == CursorHide || !m.focus
	if mode == CursorBlink {
		return Blink
	}
	return nil
}

// BlinkCmd 是用于管理光标闪烁的命令。
func (m *Model) BlinkCmd() tea.Cmd {
	if m.mode != CursorBlink {
		return nil
	}

	if m.blinkCtx != nil && m.blinkCtx.cancel != nil {
		m.blinkCtx.cancel() // 取消之前的闪烁
	}

	ctx, cancel := context.WithTimeout(m.blinkCtx.ctx, m.BlinkSpeed)
	m.blinkCtx.cancel = cancel

	m.blinkTag++ // 增加闪烁标签

	blinkMsg := BlinkMsg{id: m.id, tag: m.blinkTag}

	return func() tea.Msg {
		defer cancel()
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			return blinkMsg
		}
		return blinkCanceled{}
	}
}

// Blink 是用于初始化光标闪烁的命令。
func Blink() tea.Msg {
	return initialBlinkMsg{}
}

// Focus 聚焦光标，使其在需要时闪烁。
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.Blink = m.mode == CursorHide // 显示光标，除非我们明确隐藏它

	if m.mode == CursorBlink && m.focus {
		return m.BlinkCmd()
	}
	return nil
}

// Blur 使光标失焦。
func (m *Model) Blur() {
	m.focus = false
	m.Blink = true
}

// SetChar 设置光标下的字符。
func (m *Model) SetChar(char string) {
	m.char = char
}

// View 显示光标。
func (m Model) View() string {
	if m.Blink {
		return m.TextStyle.Inline(true).Render(m.char) // 闪烁时显示正常文本
	}
	return m.Style.Inline(true).Reverse(true).Render(m.char) // 不闪烁时显示反转样式的光标
}
