// Package timer 提供了一个简单的超时组件。
package timer

import (
	"sync/atomic"
	"time"

	tea "github.com/purpose168/bubbletea-cn"
)

var lastID int64

// nextID 生成唯一的计时器 ID
func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// Authors note with regard to start and stop commands:
//
// 从技术上讲，在这种情况下发送启动和停止计时器的命令是多余的。
// 要停止计时器，我们只需要将模型的 'running' 属性设置为 false，
// 这会导致更新函数中的逻辑停止响应 TickMsg。要启动模型，
// 我们需要将 'running' 设置为 true 并触发一个 TickMsg。辅助函数如下所示：
//
//     func (m *model) Start() tea.Cmd
//     func (m *model) Stop()
//
// 然而，这种方法的风险在于操作顺序对于类似上面的辅助函数变得很重要。请考虑以下情况：
//
//     // 不会工作
//     return m, m.timer.Start()
//
//	   // 会工作
//     cmd := m.timer.start()
//     return m, cmd
//
// 因此，由于存在上述潜在的陷阱，我们引入了额外的 StartStopMsg，
// 以简化使用此包时的心智模型。请注意，向应用程序的其他部分发送命令来进行通信的做法，
// 如此包中所示，仍然不推荐。

// StartStopMsg 用于启动和停止计时器。
type StartStopMsg struct {
	ID      int
	running bool
}

// TickMsg 是每次计时器滴答时发送的消息。
type TickMsg struct {
	// ID 是发送消息的计时器的标识符。这使得在多个计时器运行时，
	// 可以确定滴答属于哪个计时器。
	//
	// 但是请注意，计时器会拒绝来自其他计时器的滴答，
	// 因此可以将所有 TickMsg 流经所有计时器，它们仍然会表现正常。
	ID int

	// Timeout 返回此次滴答是否为超时滴答。
	// 你也可以选择监听 TimeoutMsg。
	Timeout bool

	tag int
}

// TimeoutMsg 是计时器超时时发送一次的消息。
//
// 这是一个便利消息，与 TickMsg 一起发送，Timeout 值设置为 true。
type TimeoutMsg struct {
	ID int
}

// Model 计时器组件的模型。
type Model struct {
	// Timeout 计时器到期的持续时间。
	Timeout time.Duration

	// Interval 每次滴答前的等待时间。默认为 1 秒。
	Interval time.Duration

	id      int
	tag     int
	running bool
}

// NewWithInterval 创建一个具有指定超时和滴答间隔的新计时器。
func NewWithInterval(timeout, interval time.Duration) Model {
	return Model{
		Timeout:  timeout,
		Interval: interval,
		running:  true,
		id:       nextID(),
	}
}

// New 创建一个具有指定超时和默认 1 秒间隔的新计时器。
func New(timeout time.Duration) Model {
	return NewWithInterval(timeout, time.Second)
}

// ID 返回模型的标识符。当存在多个计时器时，可用于确定消息是否属于此计时器实例。
func (m Model) ID() int {
	return m.id
}

// Running 返回计时器是否正在运行。如果计时器已超时，此方法将始终返回 false。
func (m Model) Running() bool {
	if m.Timedout() || !m.running {
		return false
	}
	return true
}

// Timedout 返回计时器是否已超时。
func (m Model) Timedout() bool {
	return m.Timeout <= 0
}

// Init 启动计时器。
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// Update 处理计时器滴答。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StartStopMsg:
		if msg.ID != 0 && msg.ID != m.id {
			return m, nil
		}
		m.running = msg.running
		return m, m.tick()
	case TickMsg:
		if !m.Running() || (msg.ID != 0 && msg.ID != m.id) {
			break
		}

		// 如果设置了标签，且不是我们期望的标签，则拒绝该消息。
		// 这可以防止接收器收到太多消息，从而导致滴答过快。
		if msg.tag > 0 && msg.tag != m.tag {
			return m, nil
		}

		m.Timeout -= m.Interval
		return m, tea.Batch(m.tick(), m.timedout())
	}

	return m, nil
}

// View 计时器组件的视图。
func (m Model) View() string {
	return m.Timeout.String()
}

// Start 恢复计时器。如果计时器已超时，则无效。
func (m *Model) Start() tea.Cmd {
	return m.startStop(true)
}

// Stop 暂停计时器。如果计时器已超时，则无效。
func (m *Model) Stop() tea.Cmd {
	return m.startStop(false)
}

// Toggle 如果计时器正在运行则停止，如果已停止则启动。
func (m *Model) Toggle() tea.Cmd {
	return m.startStop(!m.Running())
}

// tick 生成滴答消息的命令
func (m Model) tick() tea.Cmd {
	return tea.Tick(m.Interval, func(_ time.Time) tea.Msg {
		return TickMsg{ID: m.id, tag: m.tag, Timeout: m.Timedout()}
	})
}

// timedout 生成超时消息的命令
func (m Model) timedout() tea.Cmd {
	if !m.Timedout() {
		return nil
	}
	return func() tea.Msg {
		return TimeoutMsg{ID: m.id}
	}
}

// startStop 生成启动/停止消息的命令
func (m Model) startStop(v bool) tea.Cmd {
	return func() tea.Msg {
		return StartStopMsg{ID: m.id, running: v}
	}
}
