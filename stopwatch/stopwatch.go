// Package stopwatch 提供一个简单的秒表组件。
package stopwatch

import (
	"sync/atomic"
	"time"

	tea "github.com/purpose168/bubbletea-cn"
)

var lastID int64

// nextID 生成下一个唯一的 ID
func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// TickMsg 是在每次计时器触发时发送的消息。
type TickMsg struct {
	// ID 是发送消息的秒表的标识符。这使得在多个秒表同时运行时，
	// 可以确定某个触发属于哪个秒表。
	//
	// 但是请注意，秒表将拒绝来自其他秒表的触发，
	// 因此将所有 TickMsg 流经所有秒表是安全的，它们仍然会正常工作。
	ID  int // 秒表 ID
	tag int // 标签，用于防止消息过多
}

// StartStopMsg 在秒表应该启动或停止时发送。
type StartStopMsg struct {
	ID      int  // 秒表 ID
	running bool // 是否正在运行
}

// ResetMsg 在秒表应该重置时发送。
type ResetMsg struct {
	ID int // 秒表 ID
}

// Model 秒表组件的模型。
type Model struct {
	d       time.Duration // 已经过的时间
	id      int           // 唯一标识符
	tag     int           // 标签，用于防止消息过多
	running bool          // 是否正在运行

	// 在每次触发之前等待多长时间。默认为 1 秒。
	Interval time.Duration // 触发间隔
}

// NewWithInterval 使用给定的超时和触发间隔创建一个新的秒表。
func NewWithInterval(interval time.Duration) Model {
	return Model{
		Interval: interval,
		id:       nextID(),
	}
}

// New 创建一个间隔为 1 秒的新秒表。
func New() Model {
	return NewWithInterval(time.Second)
}

// ID 返回模型的唯一 ID。
func (m Model) ID() int {
	return m.id
}

// Init 启动秒表。
func (m Model) Init() tea.Cmd {
	return m.Start()
}

// Start 启动秒表。
func (m Model) Start() tea.Cmd {
	return tea.Sequence(func() tea.Msg {
		return StartStopMsg{ID: m.id, running: true}
	}, tick(m.id, m.tag, m.Interval))
}

// Stop 停止秒表。
func (m Model) Stop() tea.Cmd {
	return func() tea.Msg {
		return StartStopMsg{ID: m.id, running: false}
	}
}

// Toggle 如果秒表正在运行则停止它，如果已停止则启动它。
func (m Model) Toggle() tea.Cmd {
	if m.Running() {
		return m.Stop()
	}
	return m.Start()
}

// Reset 将秒表重置为 0。
func (m Model) Reset() tea.Cmd {
	return func() tea.Msg {
		return ResetMsg{ID: m.id}
	}
}

// Running 如果秒表正在运行则返回 true，如果已停止则返回 false。
func (m Model) Running() bool {
	return m.running
}

// Update 处理计时器触发。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StartStopMsg:
		if msg.ID != m.id {
			return m, nil
		}
		m.running = msg.running
	case ResetMsg:
		if msg.ID != m.id {
			return m, nil
		}
		m.d = 0
	case TickMsg:
		if !m.running || msg.ID != m.id {
			break
		}

		// 如果设置了标签，并且它不是我们期望的标签，则拒绝该消息。
		// 这可以防止秒表接收过多消息，从而导致触发过快。
		if msg.tag > 0 && msg.tag != m.tag {
			return m, nil
		}

		m.d += m.Interval
		m.tag++
		return m, tick(m.id, m.tag, m.Interval)
	}

	return m, nil
}

// Elapsed 返回已经过的时间。
func (m Model) Elapsed() time.Duration {
	return m.d
}

// View 计时器组件的视图。
func (m Model) View() string {
	return m.d.String()
}

// tick 触发计时器
func tick(id int, tag int, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return TickMsg{ID: id, tag: tag}
	})
}
