// Package textinput 为Bubble Tea应用程序提供文本输入组件
package textinput

import (
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	rw "github.com/mattn/go-runewidth"
	"github.com/purpose168/bubbles-cn/cursor"
	"github.com/purpose168/bubbles-cn/key"
	"github.com/purpose168/bubbles-cn/runeutil"
	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
	"github.com/rivo/uniseg"
)

// 剪贴板操作的内部消息类型
type (
	pasteMsg    string          // 粘贴成功的消息
	pasteErrMsg struct{ error } // 粘贴失败的消息
)

// EchoMode 设置文本输入字段的输入行为
type EchoMode int

const (
	// EchoNormal 按原样显示文本。这是默认行为。
	EchoNormal EchoMode = iota

	// EchoPassword 显示EchoCharacter掩码而不是实际字符。
	// 这通常用于密码字段。
	EchoPassword

	// EchoNone 在输入字符时不显示任何内容。
	// 这通常在命令行的密码字段中看到。
	EchoNone
)

// ValidateFunc 是一个函数，如果输入无效则返回错误
type ValidateFunc func(string) error

// KeyMap 是文本输入框内不同操作的键绑定
type KeyMap struct {
	CharacterForward        key.Binding // 向前移动一个字符
	CharacterBackward       key.Binding // 向后移动一个字符
	WordForward             key.Binding // 向前移动一个单词
	WordBackward            key.Binding // 向后移动一个单词
	DeleteWordBackward      key.Binding // 删除光标前的一个单词
	DeleteWordForward       key.Binding // 删除光标后的一个单词
	DeleteAfterCursor       key.Binding // 删除光标后的所有内容
	DeleteBeforeCursor      key.Binding // 删除光标前的所有内容
	DeleteCharacterBackward key.Binding // 删除光标前的一个字符
	DeleteCharacterForward  key.Binding // 删除光标后的一个字符
	LineStart               key.Binding // 移动到行首
	LineEnd                 key.Binding // 移动到行尾
	Paste                   key.Binding // 粘贴
	AcceptSuggestion        key.Binding // 接受建议
	NextSuggestion          key.Binding // 下一个建议
	PrevSuggestion          key.Binding // 上一个建议
}

// DefaultKeyMap 是默认的键绑定集合，用于导航和操作文本输入框
var DefaultKeyMap = KeyMap{
	CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f")),                  // 右箭头或Ctrl+F
	CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b")),                   // 左箭头或Ctrl+B
	WordForward:             key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f")), // Alt+右箭头或Ctrl+右箭头或Alt+F
	WordBackward:            key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b")),   // Alt+左箭头或Ctrl+左箭头或Alt+B
	DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w")),          // Alt+退格键或Ctrl+W
	DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d")),              // Alt+删除键或Alt+D
	DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k")),                           // Ctrl+K
	DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u")),                           // Ctrl+U
	DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h")),              // 退格键或Ctrl+H
	DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d")),                 // 删除键或Ctrl+D
	LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a")),                   // Home键或Ctrl+A
	LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e")),                    // End键或Ctrl+E
	Paste:                   key.NewBinding(key.WithKeys("ctrl+v")),                           // Ctrl+V
	AcceptSuggestion:        key.NewBinding(key.WithKeys("tab")),                              // Tab键
	NextSuggestion:          key.NewBinding(key.WithKeys("down", "ctrl+n")),                   // 下箭头或Ctrl+N
	PrevSuggestion:          key.NewBinding(key.WithKeys("up", "ctrl+p")),                     // 上箭头或Ctrl+P
}

// Model 是文本输入元素的Bubble Tea模型
type Model struct {
	Err error // 验证错误

	// 常规设置
	Prompt        string       // 提示符
	Placeholder   string       // 占位符文本
	EchoMode      EchoMode     // 回显模式
	EchoCharacter rune         // 回显字符（用于密码模式）
	Cursor        cursor.Model // 光标模型

	// 已弃用：请使用[cursor.BlinkSpeed]代替
	BlinkSpeed time.Duration

	// 样式。这些将作为内联样式应用
	//
	// 有关Lip Gloss样式的介绍，请参阅：
	// https://github.com/charmbracelet/lipgloss
	PromptStyle      lipgloss.Style // 提示符样式
	TextStyle        lipgloss.Style // 文本样式
	PlaceholderStyle lipgloss.Style // 占位符样式
	CompletionStyle  lipgloss.Style // 自动补全样式

	// 已弃用：请使用Cursor.Style代替
	CursorStyle lipgloss.Style

	// CharLimit 是此输入元素接受的最大字符数
	// 如果为0或更小，则没有限制
	CharLimit int

	// Width 是一次可以显示的最大字符数
	// 它本质上将文本字段视为水平滚动的视口
	// 如果为0或更小，则忽略此设置
	Width int

	// KeyMap 是小部件识别的键绑定
	KeyMap KeyMap

	// 底层文本值
	value []rune

	// focus 表示用户输入焦点是否应在此输入组件上
	// 当为false时，忽略键盘输入并隐藏光标
	focus bool

	// 光标位置
	pos int

	// 当设置了宽度且内容溢出时，用于模拟视口
	offset      int // 左偏移量
	offsetRight int // 右偏移量

	// Validate 是一个函数，用于检查输入中的文本是否有效
	// 如果无效，`Err`字段将设置为函数返回的错误
	// 如果未定义该函数，则所有输入都被视为有效
	Validate ValidateFunc

	// 输入的符文清理器
	rsan runeutil.Sanitizer

	// 是否显示自动补全建议
	ShowSuggestions bool

	// suggestions 是可用于完成输入的建议列表
	suggestions            [][]rune // 所有建议
	matchedSuggestions     [][]rune // 匹配的建议
	currentSuggestionIndex int      // 当前选中的建议索引
}

// New 创建一个具有默认设置的新模型
func New() Model {
	return Model{
		Prompt:           "> ",                                                  // 默认提示符
		EchoCharacter:    '*',                                                   // 默认回显字符（用于密码模式）
		CharLimit:        0,                                                     // 无字符限制
		PlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // 占位符样式
		ShowSuggestions:  false,                                                 // 默认不显示自动补全建议
		CompletionStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // 自动补全样式
		Cursor:           cursor.New(),                                          // 新的光标模型
		KeyMap:           DefaultKeyMap,                                         // 默认键绑定

		suggestions: [][]rune{}, // 空的建议列表
		value:       nil,        // 空的文本值
		focus:       false,      // 默认没有焦点
		pos:         0,          // 默认光标位置在开头
	}
}

// NewModel creates a new model with default settings.
//
// Deprecated: Use [New] instead.
var NewModel = New

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	// Clean up any special characters in the input provided by the
	// caller. This avoids bugs due to e.g. tab characters and whatnot.
	runes := m.san().Sanitize([]rune(s))
	err := m.validate(runes)
	m.setValueInternal(runes, err)
}

func (m *Model) setValueInternal(runes []rune, err error) {
	m.Err = err

	empty := len(m.value) == 0

	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value = runes[:m.CharLimit]
	} else {
		m.value = runes
	}
	if (m.pos == 0 && empty) || m.pos > len(m.value) {
		m.SetCursor(len(m.value))
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return string(m.value)
}

// Position returns the cursor position.
func (m Model) Position() int {
	return m.pos
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(pos int) {
	m.pos = clamp(pos, 0, len(m.value))
	m.handleOverflow()
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.value))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be shown.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return m.Cursor.Focus()
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = nil
	m.SetCursor(0)
}

// SetSuggestions sets the suggestions for the input.
func (m *Model) SetSuggestions(suggestions []string) {
	m.suggestions = make([][]rune, len(suggestions))
	for i, s := range suggestions {
		m.suggestions[i] = []rune(s)
	}

	m.updateSuggestions()
}

// rsan initializes or retrieves the rune sanitizer.
func (m *Model) san() runeutil.Sanitizer {
	if m.rsan == nil {
		// Textinput has all its input on a single line so collapse
		// newlines/tabs to single spaces.
		m.rsan = runeutil.NewSanitizer(
			runeutil.ReplaceTabs(" "), runeutil.ReplaceNewlines(" "))
	}
	return m.rsan
}

func (m *Model) insertRunesFromUserInput(v []rune) {
	// Clean up any special characters in the input provided by the
	// clipboard. This avoids bugs due to e.g. tab characters and
	// whatnot.
	paste := m.san().Sanitize(v)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - len(m.value)

		// If the char limit's been reached, cancel.
		if availSpace <= 0 {
			return
		}

		// If there's not enough space to paste the whole thing cut the pasted
		// runes down so they'll fit.
		if availSpace < len(paste) {
			paste = paste[:availSpace]
		}
	}

	// Stuff before and after the cursor
	head := m.value[:m.pos]
	tailSrc := m.value[m.pos:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	// Insert pasted runes
	for _, r := range paste {
		head = append(head, r)
		m.pos++
		if m.CharLimit > 0 {
			availSpace--
			if availSpace <= 0 {
				break
			}
		}
	}

	// Put it all back together
	value := append(head, tail...)
	inputErr := m.validate(value)
	m.setValueInternal(value, inputErr)
}

// If a max width is defined, perform some logic to treat the visible area
// as a horizontally scrolling viewport.
func (m *Model) handleOverflow() {
	if m.Width <= 0 || uniseg.StringWidth(string(m.value)) <= m.Width {
		m.offset = 0
		m.offsetRight = len(m.value)
		return
	}

	// Correct right offset if we've deleted characters
	m.offsetRight = min(m.offsetRight, len(m.value))

	if m.pos < m.offset {
		m.offset = m.pos

		w := 0
		i := 0
		runes := m.value[m.offset:]

		for i < len(runes) && w <= m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width+1 {
				i++
			}
		}

		m.offsetRight = m.offset + i
	} else if m.pos >= m.offsetRight {
		m.offsetRight = m.pos

		w := 0
		runes := m.value[:m.offsetRight]
		i := len(runes) - 1

		for i > 0 && w < m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width {
				i--
			}
		}

		m.offset = m.offsetRight - (len(runes) - 1 - i)
	}
}

// deleteBeforeCursor deletes all text before the cursor.
func (m *Model) deleteBeforeCursor() {
	m.value = m.value[m.pos:]
	m.Err = m.validate(m.value)
	m.offset = 0
	m.SetCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. If input is masked
// delete everything after the cursor so as not to reveal word breaks in the
// masked input.
func (m *Model) deleteAfterCursor() {
	m.value = m.value[:m.pos]
	m.Err = m.validate(m.value)
	m.SetCursor(len(m.value))
}

// deleteWordBackward deletes the word left to the cursor.
func (m *Model) deleteWordBackward() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.deleteBeforeCursor()
		return
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldPos := m.pos

	m.SetCursor(m.pos - 1)
	for unicode.IsSpace(m.value[m.pos]) {
		if m.pos <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		m.SetCursor(m.pos - 1)
	}

	for m.pos > 0 {
		if !unicode.IsSpace(m.value[m.pos]) {
			m.SetCursor(m.pos - 1)
		} else {
			if m.pos > 0 {
				// keep the previous space
				m.SetCursor(m.pos + 1)
			}
			break
		}
	}

	if oldPos > len(m.value) {
		m.value = m.value[:m.pos]
	} else {
		m.value = append(m.value[:m.pos], m.value[oldPos:]...)
	}
	m.Err = m.validate(m.value)
}

// deleteWordForward deletes the word right to the cursor. If input is masked
// delete everything after the cursor so as not to reveal word breaks in the
// masked input.
func (m *Model) deleteWordForward() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.deleteAfterCursor()
		return
	}

	oldPos := m.pos
	m.SetCursor(m.pos + 1)
	for unicode.IsSpace(m.value[m.pos]) {
		// ignore series of whitespace after cursor
		m.SetCursor(m.pos + 1)

		if m.pos >= len(m.value) {
			break
		}
	}

	for m.pos < len(m.value) {
		if !unicode.IsSpace(m.value[m.pos]) {
			m.SetCursor(m.pos + 1)
		} else {
			break
		}
	}

	if m.pos > len(m.value) {
		m.value = m.value[:oldPos]
	} else {
		m.value = append(m.value[:oldPos], m.value[m.pos:]...)
	}
	m.Err = m.validate(m.value)

	m.SetCursor(oldPos)
}

// wordBackward moves the cursor one word to the left. If input is masked, move
// input to the start so as not to reveal word breaks in the masked input.
func (m *Model) wordBackward() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.CursorStart()
		return
	}

	i := m.pos - 1
	for i >= 0 {
		if unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}
}

// wordForward moves the cursor one word to the right. If the input is masked,
// move input to the end so as not to reveal word breaks in the masked input.
func (m *Model) wordForward() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.CursorEnd()
		return
	}

	i := m.pos
	for i < len(m.value) {
		if unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}

	for i < len(m.value) {
		if !unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}
}

func (m Model) echoTransform(v string) string {
	switch m.EchoMode {
	case EchoPassword:
		return strings.Repeat(string(m.EchoCharacter), uniseg.StringWidth(v))
	case EchoNone:
		return ""
	case EchoNormal:
		return v
	default:
		return v
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	// Need to check for completion before, because key is configurable and might be double assigned
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok && key.Matches(keyMsg, m.KeyMap.AcceptSuggestion) {
		if m.canAcceptSuggestion() {
			m.value = append(m.value, m.matchedSuggestions[m.currentSuggestionIndex][len(m.value):]...)
			m.CursorEnd()
		}
	}

	// Let's remember where the position of the cursor currently is so that if
	// the cursor position changes, we can reset the blink.
	oldPos := m.pos

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.DeleteWordBackward):
			m.deleteWordBackward()
		case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
			m.Err = nil
			if len(m.value) > 0 {
				m.value = append(m.value[:max(0, m.pos-1)], m.value[m.pos:]...)
				m.Err = m.validate(m.value)
				if m.pos > 0 {
					m.SetCursor(m.pos - 1)
				}
			}
		case key.Matches(msg, m.KeyMap.WordBackward):
			m.wordBackward()
		case key.Matches(msg, m.KeyMap.CharacterBackward):
			if m.pos > 0 {
				m.SetCursor(m.pos - 1)
			}
		case key.Matches(msg, m.KeyMap.WordForward):
			m.wordForward()
		case key.Matches(msg, m.KeyMap.CharacterForward):
			if m.pos < len(m.value) {
				m.SetCursor(m.pos + 1)
			}
		case key.Matches(msg, m.KeyMap.LineStart):
			m.CursorStart()
		case key.Matches(msg, m.KeyMap.DeleteCharacterForward):
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
				m.Err = m.validate(m.value)
			}
		case key.Matches(msg, m.KeyMap.LineEnd):
			m.CursorEnd()
		case key.Matches(msg, m.KeyMap.DeleteAfterCursor):
			m.deleteAfterCursor()
		case key.Matches(msg, m.KeyMap.DeleteBeforeCursor):
			m.deleteBeforeCursor()
		case key.Matches(msg, m.KeyMap.Paste):
			return m, Paste
		case key.Matches(msg, m.KeyMap.DeleteWordForward):
			m.deleteWordForward()
		case key.Matches(msg, m.KeyMap.NextSuggestion):
			m.nextSuggestion()
		case key.Matches(msg, m.KeyMap.PrevSuggestion):
			m.previousSuggestion()
		default:
			// Input one or more regular characters.
			m.insertRunesFromUserInput(msg.Runes)
		}

		// Check again if can be completed
		// because value might be something that does not match the completion prefix
		m.updateSuggestions()

	case pasteMsg:
		m.insertRunesFromUserInput([]rune(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	if oldPos != m.pos && m.Cursor.Mode() == cursor.CursorBlink {
		m.Cursor.Blink = false
		cmds = append(cmds, m.Cursor.BlinkCmd())
	}

	m.handleOverflow()
	return m, tea.Batch(cmds...)
}

// View renders the textinput in its current state.
func (m Model) View() string {
	// Placeholder text
	if len(m.value) == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	styleText := m.TextStyle.Inline(true).Render

	value := m.value[m.offset:m.offsetRight]
	pos := max(0, m.pos-m.offset)
	v := styleText(m.echoTransform(string(value[:pos])))

	if pos < len(value) { //nolint:nestif
		char := m.echoTransform(string(value[pos]))
		m.Cursor.SetChar(char)
		v += m.Cursor.View()                                   // cursor and text under it
		v += styleText(m.echoTransform(string(value[pos+1:]))) // text after cursor
		v += m.completionView(0)                               // suggested completion
	} else {
		if m.focus && m.canAcceptSuggestion() {
			suggestion := m.matchedSuggestions[m.currentSuggestionIndex]
			if len(value) < len(suggestion) {
				m.Cursor.TextStyle = m.CompletionStyle
				m.Cursor.SetChar(m.echoTransform(string(suggestion[pos])))
				v += m.Cursor.View()
				v += m.completionView(1)
			} else {
				m.Cursor.SetChar(" ")
				v += m.Cursor.View()
			}
		} else {
			m.Cursor.SetChar(" ")
			v += m.Cursor.View()
		}
	}

	// If a max width and background color were set fill the empty spaces with
	// the background color.
	valWidth := uniseg.StringWidth(string(value))
	if m.Width > 0 && valWidth <= m.Width {
		padding := max(0, m.Width-valWidth)
		if valWidth+padding <= m.Width && pos < len(value) {
			padding++
		}
		v += styleText(strings.Repeat(" ", padding))
	}

	return m.PromptStyle.Render(m.Prompt) + v
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		v     string
		style = m.PlaceholderStyle.Inline(true).Render
		p     = m.PromptStyle.Render(m.Prompt)
	)

	m.Cursor.TextStyle = m.PlaceholderStyle
	first, rest, _, _ := uniseg.FirstGraphemeClusterInString(m.Placeholder, 0)
	m.Cursor.SetChar(first)
	v += m.Cursor.View()

	// If the entire placeholder is already set and no padding is needed, finish
	if m.Width < 1 && uniseg.StringWidth(rest) <= 1 {
		return m.PromptStyle.Render(m.Prompt) + v
	}

	// If Width is set then size placeholder accordingly
	if m.Width > 0 {
		width := m.Width - lipgloss.Width(p) - lipgloss.Width(v)
		placeholderRest := ansi.Truncate(rest, width, "…")
		availWidth := max(0, width-lipgloss.Width(placeholderRest))
		v += style(placeholderRest) + strings.Repeat(" ", availWidth)
	} else {
		// if there is no width, the placeholder can be any length
		v += style(rest)
	}

	return p + v
}

// Blink is a command used to initialize cursor blinking.
func Blink() tea.Msg {
	return cursor.Blink()
}

// Paste is a command for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

// Deprecated.

// Deprecated: use [cursor.Mode].
//
//nolint:revive
type CursorMode int

//nolint:revive
const (
	// Deprecated: use [cursor.CursorBlink].
	CursorBlink = CursorMode(cursor.CursorBlink)
	// Deprecated: use [cursor.CursorStatic].
	CursorStatic = CursorMode(cursor.CursorStatic)
	// Deprecated: use [cursor.CursorHide].
	CursorHide = CursorMode(cursor.CursorHide)
)

func (c CursorMode) String() string {
	return cursor.Mode(c).String()
}

// Deprecated: use [cursor.Mode].
//
//nolint:revive
func (m Model) CursorMode() CursorMode {
	return CursorMode(m.Cursor.Mode())
}

// Deprecated: use cursor.SetMode().
//
//nolint:revive
func (m *Model) SetCursorMode(mode CursorMode) tea.Cmd {
	return m.Cursor.SetMode(cursor.Mode(mode))
}

func (m Model) completionView(offset int) string {
	var (
		value = m.value
		style = m.PlaceholderStyle.Inline(true).Render
	)

	if m.canAcceptSuggestion() {
		suggestion := m.matchedSuggestions[m.currentSuggestionIndex]
		if len(value) < len(suggestion) {
			return style(string(suggestion[len(value)+offset:]))
		}
	}
	return ""
}

func (m *Model) getSuggestions(sugs [][]rune) []string {
	suggestions := make([]string, len(sugs))
	for i, s := range sugs {
		suggestions[i] = string(s)
	}
	return suggestions
}

// AvailableSuggestions returns the list of available suggestions.
func (m *Model) AvailableSuggestions() []string {
	return m.getSuggestions(m.suggestions)
}

// MatchedSuggestions returns the list of matched suggestions.
func (m *Model) MatchedSuggestions() []string {
	return m.getSuggestions(m.matchedSuggestions)
}

// CurrentSuggestionIndex returns the currently selected suggestion index.
func (m *Model) CurrentSuggestionIndex() int {
	return m.currentSuggestionIndex
}

// CurrentSuggestion returns the currently selected suggestion.
func (m *Model) CurrentSuggestion() string {
	if m.currentSuggestionIndex >= len(m.matchedSuggestions) {
		return ""
	}

	return string(m.matchedSuggestions[m.currentSuggestionIndex])
}

// canAcceptSuggestion returns whether there is an acceptable suggestion to
// autocomplete the current value.
func (m *Model) canAcceptSuggestion() bool {
	return len(m.matchedSuggestions) > 0
}

// updateSuggestions refreshes the list of matching suggestions.
func (m *Model) updateSuggestions() {
	if !m.ShowSuggestions {
		return
	}

	if len(m.value) <= 0 || len(m.suggestions) <= 0 {
		m.matchedSuggestions = [][]rune{}
		return
	}

	matches := [][]rune{}
	for _, s := range m.suggestions {
		suggestion := string(s)

		if strings.HasPrefix(strings.ToLower(suggestion), strings.ToLower(string(m.value))) {
			matches = append(matches, []rune(suggestion))
		}
	}
	if !reflect.DeepEqual(matches, m.matchedSuggestions) {
		m.currentSuggestionIndex = 0
	}

	m.matchedSuggestions = matches
}

// nextSuggestion selects the next suggestion.
func (m *Model) nextSuggestion() {
	m.currentSuggestionIndex = (m.currentSuggestionIndex + 1)
	if m.currentSuggestionIndex >= len(m.matchedSuggestions) {
		m.currentSuggestionIndex = 0
	}
}

// previousSuggestion selects the previous suggestion.
func (m *Model) previousSuggestion() {
	m.currentSuggestionIndex = (m.currentSuggestionIndex - 1)
	if m.currentSuggestionIndex < 0 {
		m.currentSuggestionIndex = len(m.matchedSuggestions) - 1
	}
}

func (m Model) validate(v []rune) error {
	if m.Validate != nil {
		return m.Validate(string(v))
	}
	return nil
}
