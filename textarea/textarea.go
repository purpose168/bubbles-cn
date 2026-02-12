// Package textarea 为 Bubble Tea 应用程序提供多行文本输入组件。
package textarea

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	rw "github.com/mattn/go-runewidth"
	"github.com/purpose168/bubbles-cn/cursor"
	"github.com/purpose168/bubbles-cn/key"
	"github.com/purpose168/bubbles-cn/runeutil"
	"github.com/purpose168/bubbles-cn/textarea/memoization"
	"github.com/purpose168/bubbles-cn/viewport"
	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
	"github.com/rivo/uniseg"
)

const (
	minHeight        = 1   // 最小高度
	defaultHeight    = 6   // 默认高度
	defaultWidth     = 40  // 默认宽度
	defaultCharLimit = 0   // 无限制
	defaultMaxHeight = 99  // 默认最大高度
	defaultMaxWidth  = 500 // 默认最大宽度

	// XXX: 在 v2 版本中，使最大行数动态化，并使默认最大行数可配置。
	maxLines = 10000 // 最大行数
)

// 剪贴板操作的内部消息。
type (
	pasteMsg    string          // 粘贴消息
	pasteErrMsg struct{ error } // 粘贴错误消息
)

// KeyMap 定义了 textarea 中不同操作的键绑定。
type KeyMap struct {
	CharacterBackward       key.Binding // 字符向后
	CharacterForward        key.Binding // 字符向前
	DeleteAfterCursor       key.Binding // 删除光标后
	DeleteBeforeCursor      key.Binding // 删除光标前
	DeleteCharacterBackward key.Binding // 向后删除字符
	DeleteCharacterForward  key.Binding // 向前删除字符
	DeleteWordBackward      key.Binding // 向后删除单词
	DeleteWordForward       key.Binding // 向前删除单词
	InsertNewline           key.Binding // 插入换行
	LineEnd                 key.Binding // 行尾
	LineNext                key.Binding // 下一行
	LinePrevious            key.Binding // 上一行
	LineStart               key.Binding // 行首
	Paste                   key.Binding // 粘贴
	WordBackward            key.Binding // 单词向后
	WordForward             key.Binding // 单词向前
	InputBegin              key.Binding // 输入开始
	InputEnd                key.Binding // 输入结束

	UppercaseWordForward  key.Binding // 向前大写单词
	LowercaseWordForward  key.Binding // 向前小写单词
	CapitalizeWordForward key.Binding // 向前首字母大写单词

	TransposeCharacterBackward key.Binding // 向前交换字符
}

// DefaultKeyMap 是用于在 textarea 中导航和操作的默认键绑定集合。
var DefaultKeyMap = KeyMap{
	CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f"), key.WithHelp("right", "character forward")),
	CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b"), key.WithHelp("left", "character backward")),
	WordForward:             key.NewBinding(key.WithKeys("alt+right", "alt+f"), key.WithHelp("alt+right", "word forward")),
	WordBackward:            key.NewBinding(key.WithKeys("alt+left", "alt+b"), key.WithHelp("alt+left", "word backward")),
	LineNext:                key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down", "next line")),
	LinePrevious:            key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up", "previous line")),
	DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w"), key.WithHelp("alt+backspace", "delete word backward")),
	DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d"), key.WithHelp("alt+delete", "delete word forward")),
	DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "delete after cursor")),
	DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "delete before cursor")),
	InsertNewline:           key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "insert newline")),
	DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h"), key.WithHelp("backspace", "delete character backward")),
	DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d"), key.WithHelp("delete", "delete character forward")),
	LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a"), key.WithHelp("home", "line start")),
	LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e"), key.WithHelp("end", "line end")),
	Paste:                   key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
	InputBegin:              key.NewBinding(key.WithKeys("alt+<", "ctrl+home"), key.WithHelp("alt+<", "input begin")),
	InputEnd:                key.NewBinding(key.WithKeys("alt+>", "ctrl+end"), key.WithHelp("alt+>", "input end")),

	CapitalizeWordForward: key.NewBinding(key.WithKeys("alt+c"), key.WithHelp("alt+c", "capitalize word forward")),
	LowercaseWordForward:  key.NewBinding(key.WithKeys("alt+l"), key.WithHelp("alt+l", "lowercase word forward")),
	UppercaseWordForward:  key.NewBinding(key.WithKeys("alt+u"), key.WithHelp("alt+u", "uppercase word forward")),

	TransposeCharacterBackward: key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "transpose character backward")),
}

// LineInfo 是一个辅助结构，用于跟踪软换行相关的行信息。
type LineInfo struct {
	// Width 是行中的列数。
	Width int
	// CharWidth 是行中的字符数，用于处理双宽度字符。
	CharWidth int
	// Height 是行中的行数。
	Height int
	// StartColumn 是行第一列的索引。
	StartColumn int
	// ColumnOffset 是光标距离行首的列偏移量。
	ColumnOffset int
	// RowOffset 是光标距离行首的行偏移量。
	RowOffset int
	// CharOffset 是光标距离行首的字符偏移量。这通常与 ColumnOffset 相等，
	// 但如果光标前有双宽度字符，则会不同。
	CharOffset int
}

// Style 是应用于文本区域的样式。
//
// Style 可以应用于聚焦和非聚焦状态，以根据聚焦状态更改样式。
//
// 有关使用 Lip Gloss 进行样式设置的介绍，请参阅：
// https://github.com/charmbracelet/lipgloss
type Style struct {
	Base             lipgloss.Style // 基础样式
	CursorLine       lipgloss.Style // 光标行样式
	CursorLineNumber lipgloss.Style // 光标行号样式
	EndOfBuffer      lipgloss.Style // 缓冲区结束样式
	LineNumber       lipgloss.Style // 行号样式
	Placeholder      lipgloss.Style // 占位符样式
	Prompt           lipgloss.Style // 提示符样式
	Text             lipgloss.Style // 文本样式
}

func (s Style) computedCursorLine() lipgloss.Style {
	return s.CursorLine.Inherit(s.Base).Inline(true)
}

func (s Style) computedCursorLineNumber() lipgloss.Style {
	return s.CursorLineNumber.
		Inherit(s.CursorLine).
		Inherit(s.Base).
		Inline(true)
}

func (s Style) computedEndOfBuffer() lipgloss.Style {
	return s.EndOfBuffer.Inherit(s.Base).Inline(true)
}

func (s Style) computedLineNumber() lipgloss.Style {
	return s.LineNumber.Inherit(s.Base).Inline(true)
}

func (s Style) computedPlaceholder() lipgloss.Style {
	return s.Placeholder.Inherit(s.Base).Inline(true)
}

func (s Style) computedPrompt() lipgloss.Style {
	return s.Prompt.Inherit(s.Base).Inline(true)
}

func (s Style) computedText() lipgloss.Style {
	return s.Text.Inherit(s.Base).Inline(true)
}

// line 是文本换行函数的输入。这存储在一个结构体中，以便进行哈希和记忆化。
type line struct {
	runes []rune // 字符数组
	width int    // 宽度
}

// Hash 返回行的哈希值。
func (w line) Hash() string {
	v := fmt.Sprintf("%s:%d", string(w.runes), w.width)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
}

// Model 是此文本区域元素的 Bubble Tea 模型。
type Model struct {
	Err error // 错误

	// 通用设置。
	cache *memoization.MemoCache[line, [][]rune] // 缓存

	// Prompt 在每行的开头打印。
	//
	// 在模型初始化后更改 Prompt 的值时，确保之后调用 SetWidth()。
	//
	// 另请参阅 SetPromptFunc()。
	Prompt string

	// Placeholder 是当用户尚未输入任何内容时显示的文本。
	Placeholder string

	// ShowLineNumbers 如果启用，会导致在提示符后打印行号。
	ShowLineNumbers bool

	// EndOfBufferCharacter 在输入的末尾显示。
	EndOfBufferCharacter rune

	// KeyMap 编码了小部件识别的键绑定。
	KeyMap KeyMap

	// 样式。FocusedStyle 和 BlurredStyle 用于在聚焦和模糊状态下设置 textarea 的样式。
	FocusedStyle Style
	BlurredStyle Style
	// style 是当前使用的样式。
	// 它用于在设置模型样式时抽象聚焦状态的差异，因为我们可以简单地将一组样式
	// 分配给此变量，以便在切换聚焦状态时使用。
	style *Style

	// Cursor 是文本区域的光标。
	Cursor cursor.Model

	// CharLimit 是此输入元素将接受的最大字符数。如果为 0 或更小，则没有限制。
	CharLimit int

	// MaxHeight 是文本区域的最大高度（以行为单位）。如果为 0 或更小，则没有限制。
	MaxHeight int

	// MaxWidth 是文本区域的最大宽度（以列为单位）。如果为 0 或更小，则没有限制。
	MaxWidth int

	// 如果设置了 promptFunc，它将替换 Prompt 作为每行开头提示符字符串的生成器。
	promptFunc func(line int) string

	// promptWidth 是提示符的宽度。
	promptWidth int

	// width 是可以一次显示的最大字符数。如果为 0 或更小，则忽略此设置。
	width int

	// height 是可以一次显示的最大行数。如果行数超过允许的高度，
	// 它实际上将文本字段视为垂直滚动的视口。
	height int

	// 底层文本值。
	value [][]rune

	// focus 指示用户输入焦点是否应在此输入组件上。当为 false 时，忽略键盘输入并隐藏光标。
	focus bool

	// 光标列。
	col int

	// 光标行。
	row int

	// 最后一个字符偏移量，用于在垂直移动光标时保持状态，以便我们可以保持相同的导航位置。
	lastCharOffset int

	// viewport 是多行文本输入的垂直滚动视口。
	viewport *viewport.Model

	// 输入的字符清理器。
	rsan runeutil.Sanitizer
}

// New 创建一个具有默认设置的新模型。
func New() Model {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}
	cur := cursor.New()

	focusedStyle, blurredStyle := DefaultStyles()

	m := Model{
		CharLimit:            defaultCharLimit,
		MaxHeight:            defaultMaxHeight,
		MaxWidth:             defaultMaxWidth,
		Prompt:               lipgloss.ThickBorder().Left + " ",
		style:                &blurredStyle,
		FocusedStyle:         focusedStyle,
		BlurredStyle:         blurredStyle,
		cache:                memoization.NewMemoCache[line, [][]rune](maxLines),
		EndOfBufferCharacter: ' ',
		ShowLineNumbers:      true,
		Cursor:               cur,
		KeyMap:               DefaultKeyMap,

		value: make([][]rune, minHeight, maxLines),
		focus: false,
		col:   0,
		row:   0,

		viewport: &vp,
	}

	m.SetHeight(defaultHeight)
	m.SetWidth(defaultWidth)

	return m
}

// DefaultStyles 返回 textarea 的聚焦和模糊状态的默认样式。
func DefaultStyles() (Style, Style) {
	focused := Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "240"}),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle(),
	}
	blurred := Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
	}

	return focused, blurred
}

// SetValue 设置文本输入的值。
func (m *Model) SetValue(s string) {
	m.Reset()
	m.InsertString(s)
}

// InsertString 在光标位置插入一个字符串。
func (m *Model) InsertString(s string) {
	m.insertRunesFromUserInput([]rune(s))
}

// InsertRune 在光标位置插入一个字符。
func (m *Model) InsertRune(r rune) {
	m.insertRunesFromUserInput([]rune{r})
}

// insertRunesFromUserInput 在当前光标位置插入字符。
func (m *Model) insertRunesFromUserInput(runes []rune) {
	// 清理剪贴板提供的输入中的任何特殊字符。这避免了由于制表符等
	// 字符导致的错误。
	runes = m.san().Sanitize(runes)

	if m.CharLimit > 0 {
		availSpace := m.CharLimit - m.Length()
		// 如果已达到字符限制，则取消。
		if availSpace <= 0 {
			return
		}
		// 如果没有足够的空间粘贴整个内容，则截断粘贴的字符以使其适合。
		if availSpace < len(runes) {
			runes = runes[:availSpace]
		}
	}

	// 将输入分割成行。
	var lines [][]rune
	lstart := 0
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\n' {
			// 将一行排队成为下方文本区域中的新行。注意限制切片的最大容量，
			// 以确保不同行的数据在后续编辑修改此行时不会被覆盖。
			lines = append(lines, runes[lstart:i:i])
			lstart = i + 1
		}
	}
	if lstart <= len(runes) {
		// 最后一行没有以换行符结尾。现在获取它。
		lines = append(lines, runes[lstart:])
	}

	// 遵守最大行数限制。
	if maxLines > 0 && len(m.value)+len(lines)-1 > maxLines {
		allowedHeight := max(0, maxLines-len(m.value)+1)
		lines = lines[:allowedHeight]
	}

	if len(lines) == 0 {
		// 没有剩余内容可插入。
		return
	}

	// 保存当前光标位置处原始行的剩余部分。
	tail := make([]rune, len(m.value[m.row][m.col:]))
	copy(tail, m.value[m.row][m.col:])

	// 在当前光标位置粘贴第一行。
	m.value[m.row] = append(m.value[m.row][:m.col], lines[0]...)
	m.col += len(lines[0])

	if numExtraLines := len(lines) - 1; numExtraLines > 0 {
		// 添加新行。如果已有空间，我们尝试重用切片。
		var newGrid [][]rune
		if cap(m.value) >= len(m.value)+numExtraLines {
			// 可以重用额外的空间。
			newGrid = m.value[:len(m.value)+numExtraLines]
		} else {
			// 没有剩余空间；需要一个新的切片。
			newGrid = make([][]rune, len(m.value)+numExtraLines)
			copy(newGrid, m.value[:m.row+1])
		}
		// 将原始网格中光标之后的所有行添加到新网格的末尾。
		copy(newGrid[m.row+1+numExtraLines:], m.value[m.row+1:])
		m.value = newGrid
		// 在中间插入所有新行。
		for _, l := range lines[1:] {
			m.row++
			m.value[m.row] = l
			m.col = len(l)
		}
	}

	// 最后在插入的最后一行的末尾添加尾部。
	m.value[m.row] = append(m.value[m.row], tail...)

	m.SetCursor(m.col)
}

// Value 返回文本输入的值。
func (m Model) Value() string {
	if m.value == nil {
		return ""
	}

	var v strings.Builder
	for _, l := range m.value {
		v.WriteString(string(l))
		v.WriteByte('\n')
	}

	return strings.TrimSuffix(v.String(), "\n")
}

// Length 返回文本输入中当前的字符数。
func (m *Model) Length() int {
	var l int
	for _, row := range m.value {
		l += uniseg.StringWidth(string(row))
	}
	// 我们添加 len(m.value) 以包含换行符。
	return l + len(m.value) - 1
}

// LineCount 返回文本输入中当前的行数。
func (m *Model) LineCount() int {
	return len(m.value)
}

// Line 返回行位置。
func (m Model) Line() int {
	return m.row
}

// CursorDown 将光标向下移动一行。
// 返回是否应该重置光标闪烁。
func (m *Model) CursorDown() {
	li := m.LineInfo()
	charOffset := max(m.lastCharOffset, li.CharOffset)
	m.lastCharOffset = charOffset

	if li.RowOffset+1 >= li.Height && m.row < len(m.value)-1 {
		m.row++
		m.col = 0
	} else {
		// 将光标移动到下一行的开头，以便我们可以获取行信息。
		// 我们需要添加 2 列来考虑尾随空格换行。
		const trailingSpace = 2
		m.col = min(li.StartColumn+li.Width+trailingSpace, len(m.value[m.row])-1)
	}

	nli := m.LineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	offset := 0
	for offset < charOffset {
		if m.row >= len(m.value) || m.col >= len(m.value[m.row]) || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}
}

// CursorUp 将光标向上移动一行。
func (m *Model) CursorUp() {
	li := m.LineInfo()
	charOffset := max(m.lastCharOffset, li.CharOffset)
	m.lastCharOffset = charOffset

	if li.RowOffset <= 0 && m.row > 0 {
		m.row--
		m.col = len(m.value[m.row])
	} else {
		// 将光标移动到上一行的末尾。
		// 这可以通过将光标移动到行的开头，然后减去 2 来实现，
		// 以考虑我们在软换行上保留的尾随空格。
		const trailingSpace = 2
		m.col = li.StartColumn - trailingSpace
	}

	nli := m.LineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	offset := 0
	for offset < charOffset {
		if m.col >= len(m.value[m.row]) || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(m.value[m.row][m.col])
		m.col++
	}
}

// SetCursor 将光标移动到给定位置。如果位置超出范围，
// 光标将相应地移动到开头或结尾。
func (m *Model) SetCursor(col int) {
	m.col = clamp(col, 0, len(m.value[m.row]))
	// 每当我们水平移动光标时，我们需要重置最后的偏移量，
	// 以便在导航时调整水平位置。
	m.lastCharOffset = 0
}

// CursorStart 将光标移动到输入字段的开头。
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd 将光标移动到输入字段的末尾。
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.value[m.row]))
}

// Focused 返回模型上的聚焦状态。
func (m Model) Focused() bool {
	return m.focus
}

// Focus 在模型上设置聚焦状态。当模型处于聚焦状态时，它可以
// 接收键盘输入，光标将显示。
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.style = &m.FocusedStyle
	return m.Cursor.Focus()
}

// Blur 移除模型上的聚焦状态。当模型处于模糊状态时，它
// 不能接收键盘输入，光标将隐藏。
func (m *Model) Blur() {
	m.focus = false
	m.style = &m.BlurredStyle
	m.Cursor.Blur()
}

// Reset 将输入设置为其默认状态，没有输入。
func (m *Model) Reset() {
	m.value = make([][]rune, minHeight, maxLines)
	m.col = 0
	m.row = 0
	m.viewport.GotoTop()
	m.SetCursor(0)
}

// san 初始化或检索字符清理器。
func (m *Model) san() runeutil.Sanitizer {
	if m.rsan == nil {
		// Textinput 将其所有输入都放在单行上，因此将换行符/制表符折叠为单个空格。
		m.rsan = runeutil.NewSanitizer()
	}
	return m.rsan
}

// deleteBeforeCursor 删除光标之前的所有文本。返回是否应该重置光标闪烁。
func (m *Model) deleteBeforeCursor() {
	m.value[m.row] = m.value[m.row][m.col:]
	m.SetCursor(0)
}

// deleteAfterCursor 删除光标之后的所有文本。返回是否应该重置光标闪烁。
// 如果输入被屏蔽，则删除光标之后的所有内容，以免在屏蔽输入中显示单词中断。
func (m *Model) deleteAfterCursor() {
	m.value[m.row] = m.value[m.row][:m.col]
	m.SetCursor(len(m.value[m.row]))
}

// transposeLeft 交换光标处的字符和紧随其后的字符。如果光标在行的开头，则无操作。
// 如果光标尚未在行的末尾，则将光标向右移动。
func (m *Model) transposeLeft() {
	if m.col == 0 || len(m.value[m.row]) < 2 {
		return
	}
	if m.col >= len(m.value[m.row]) {
		m.SetCursor(m.col - 1)
	}
	m.value[m.row][m.col-1], m.value[m.row][m.col] = m.value[m.row][m.col], m.value[m.row][m.col-1]
	if m.col < len(m.value[m.row]) {
		m.SetCursor(m.col + 1)
	}
}

// deleteWordLeft 删除光标左侧的单词。返回是否应该重置光标闪烁。
func (m *Model) deleteWordLeft() {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return
	}

	// Linter 注意：在这里获取初始光标位置至关重要，因为在下面通过 SetCursor()
	// 修改它之前。因此，将此调用移动到相应的 if 子句中不适用。
	oldCol := m.col

	m.SetCursor(m.col - 1)
	for unicode.IsSpace(m.value[m.row][m.col]) {
		if m.col <= 0 {
			break
		}
		// 忽略光标前的空白字符序列
		m.SetCursor(m.col - 1)
	}

	for m.col > 0 {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.SetCursor(m.col - 1)
		} else {
			if m.col > 0 {
				// 保留前一个空格
				m.SetCursor(m.col + 1)
			}
			break
		}
	}

	if oldCol > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:m.col]
	} else {
		m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][oldCol:]...)
	}
}

// deleteWordRight 删除光标右侧的单词。
func (m *Model) deleteWordRight() {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return
	}

	oldCol := m.col

	for m.col < len(m.value[m.row]) && unicode.IsSpace(m.value[m.row][m.col]) {
		// 忽略光标后的空白字符序列
		m.SetCursor(m.col + 1)
	}

	for m.col < len(m.value[m.row]) {
		if !unicode.IsSpace(m.value[m.row][m.col]) {
			m.SetCursor(m.col + 1)
		} else {
			break
		}
	}

	if m.col > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:oldCol]
	} else {
		m.value[m.row] = append(m.value[m.row][:oldCol], m.value[m.row][m.col:]...)
	}

	m.SetCursor(oldCol)
}

// characterRight 将光标向右移动一个字符。
func (m *Model) characterRight() {
	if m.col < len(m.value[m.row]) {
		m.SetCursor(m.col + 1)
	} else {
		if m.row < len(m.value)-1 {
			m.row++
			m.CursorStart()
		}
	}
}

// characterLeft 将光标向左移动一个字符。
// 如果设置了 insideLine，光标将移动到上一行的最后一个字符，而不是其后的一个字符。
func (m *Model) characterLeft(insideLine bool) {
	if m.col == 0 && m.row != 0 {
		m.row--
		m.CursorEnd()
		if !insideLine {
			return
		}
	}
	if m.col > 0 {
		m.SetCursor(m.col - 1)
	}
}

// wordLeft 将光标向左移动一个单词。返回是否应该重置光标闪烁。
// 如果输入被屏蔽，则将输入移动到开头，以免在屏蔽输入中显示单词中断。
func (m *Model) wordLeft() {
	for {
		m.characterLeft(true /* insideLine */)
		if m.col < len(m.value[m.row]) && !unicode.IsSpace(m.value[m.row][m.col]) {
			break
		}
	}

	for m.col > 0 {
		if unicode.IsSpace(m.value[m.row][m.col-1]) {
			break
		}
		m.SetCursor(m.col - 1)
	}
}

// wordRight 将光标向右移动一个单词。返回是否应该重置光标闪烁。
// 如果输入被屏蔽，则将输入移动到末尾，以免在屏蔽输入中显示单词中断。
func (m *Model) wordRight() {
	m.doWordRight(func(int, int) { /* nothing */ })
}

func (m *Model) doWordRight(fn func(charIdx int, pos int)) {
	// 向前跳过空格。
	for m.col >= len(m.value[m.row]) || unicode.IsSpace(m.value[m.row][m.col]) {
		if m.row == len(m.value)-1 && m.col == len(m.value[m.row]) {
			// 文本末尾。
			break
		}
		m.characterRight()
	}

	charIdx := 0
	for m.col < len(m.value[m.row]) {
		if unicode.IsSpace(m.value[m.row][m.col]) {
			break
		}
		fn(charIdx, m.col)
		m.SetCursor(m.col + 1)
		charIdx++
	}
}

// uppercaseRight 将右侧的单词更改为大写。
func (m *Model) uppercaseRight() {
	m.doWordRight(func(_ int, i int) {
		m.value[m.row][i] = unicode.ToUpper(m.value[m.row][i])
	})
}

// lowercaseRight 将右侧的单词更改为小写。
func (m *Model) lowercaseRight() {
	m.doWordRight(func(_ int, i int) {
		m.value[m.row][i] = unicode.ToLower(m.value[m.row][i])
	})
}

// capitalizeRight 将右侧的单词更改为标题大小写。
func (m *Model) capitalizeRight() {
	m.doWordRight(func(charIdx int, i int) {
		if charIdx == 0 {
			m.value[m.row][i] = unicode.ToTitle(m.value[m.row][i])
		}
	})
}

// LineInfo 返回从（软换行）行开头到（软换行）行的字符数和（软换行）行宽度。
func (m Model) LineInfo() LineInfo {
	grid := m.memoizedWrap(m.value[m.row], m.width)

	// 找出我们当前在哪一行。这可以通过 m.col 和计算我们需要跳过的字符数来确定。
	var counter int
	for i, line := range grid {
		// 我们找到了我们所在的行
		if counter+len(line) == m.col && i+1 < len(grid) {
			// 如果我们在上一行的末尾，则绕到下一行，以便我们可以位于行的最开头
			return LineInfo{
				CharOffset:   0,
				ColumnOffset: 0,
				Height:       len(grid),
				RowOffset:    i + 1,
				StartColumn:  m.col,
				Width:        len(grid[i+1]),
				CharWidth:    uniseg.StringWidth(string(line)),
			}
		}

		if counter+len(line) >= m.col {
			return LineInfo{
				CharOffset:   uniseg.StringWidth(string(line[:max(0, m.col-counter)])),
				ColumnOffset: m.col - counter,
				Height:       len(grid),
				RowOffset:    i,
				StartColumn:  counter,
				Width:        len(line),
				CharWidth:    uniseg.StringWidth(string(line)),
			}
		}

		counter += len(line)
	}
	return LineInfo{}
}

// repositionView 根据定义的滚动行为重新定位视口的视图。
func (m *Model) repositionView() {
	minimum := m.viewport.YOffset
	maximum := minimum + m.viewport.Height - 1

	if row := m.cursorLineNumber(); row < minimum {
		m.viewport.ScrollUp(minimum - row)
	} else if row > maximum {
		m.viewport.ScrollDown(row - maximum)
	}
}

// Width 返回文本区域的宽度。
func (m Model) Width() int {
	return m.width
}

// moveToBegin 将光标移动到输入的开头。
func (m *Model) moveToBegin() {
	m.row = 0
	m.SetCursor(0)
}

// moveToEnd 将光标移动到输入的末尾。
func (m *Model) moveToEnd() {
	m.row = len(m.value) - 1
	m.SetCursor(len(m.value[m.row]))
}

// SetWidth 设置文本区域的宽度以完全适应给定的宽度。
// 这意味着文本区域将考虑提示符的宽度以及是否显示行号。
//
// 确保在设置 Prompt 和 ShowLineNumbers 之后调用 SetWidth，
// 文本区域的宽度必须正好是给定的宽度，不能更多。
func (m *Model) SetWidth(w int) {
	// 仅当没有提示符函数时才更新提示符宽度，因为 SetPromptFunc
	// 在调用时会更新提示符宽度。
	if m.promptFunc == nil {
		m.promptWidth = uniseg.StringWidth(m.Prompt)
	}

	// 将基础样式边框和填充添加到保留的外部宽度。
	reservedOuter := m.style.Base.GetHorizontalFrameSize()

	// 将提示符宽度添加到保留的内部宽度。
	reservedInner := m.promptWidth

	// 将行号宽度添加到保留的内部宽度。
	if m.ShowLineNumbers {
		const lnWidth = 4 // 行号最多 3 位数加上 1 个边距。
		reservedInner += lnWidth
	}

	// 输入宽度必须至少比保留的内部和外部宽度多 1。这给我们最小的输入宽度为 1。
	minWidth := reservedInner + reservedOuter + 1
	inputWidth := max(w, minWidth)

	// 输入宽度不得超过最大宽度。
	if m.MaxWidth > 0 {
		inputWidth = min(inputWidth, m.MaxWidth)
	}

	// 由于视口和输入区域的宽度取决于边框、提示符和行号的宽度，
	// 我们需要通过从中减去保留的宽度来计算它。

	m.viewport.Width = inputWidth - reservedOuter
	m.width = inputWidth - reservedOuter - reservedInner
}

// SetPromptFunc 取代 Prompt 字段并设置动态提示符。
// 如果函数返回的提示符比指定的 promptWidth 短，它将在左侧填充。
// 如果它返回的提示符更长，可能会出现显示伪影；
// 调用者负责计算足够的 promptWidth。
func (m *Model) SetPromptFunc(promptWidth int, fn func(lineIdx int) string) {
	m.promptFunc = fn
	m.promptWidth = promptWidth
}

// Height 返回文本区域的当前高度。
func (m Model) Height() int {
	return m.height
}

// SetHeight 设置文本区域的高度。
func (m *Model) SetHeight(h int) {
	if m.MaxHeight > 0 {
		m.height = clamp(h, minHeight, m.MaxHeight)
		m.viewport.Height = clamp(h, minHeight, m.MaxHeight)
	} else {
		m.height = max(h, minHeight)
		m.viewport.Height = max(h, minHeight)
	}
}

// Update 是 Bubble Tea 更新循环。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		m.Cursor.Blur()
		return m, nil
	}

	// 用于确定光标是否应该闪烁。
	oldRow, oldCol := m.cursorLineNumber(), m.col

	var cmds []tea.Cmd

	if m.value[m.row] == nil {
		m.value[m.row] = make([]rune, 0)
	}

	if m.MaxHeight > 0 && m.MaxHeight != m.cache.Capacity() {
		m.cache = memoization.NewMemoCache[line, [][]rune](m.MaxHeight)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.DeleteAfterCursor):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteAfterCursor()
		case key.Matches(msg, m.KeyMap.DeleteBeforeCursor):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteBeforeCursor()
		case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			if len(m.value[m.row]) > 0 {
				m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
				if m.col > 0 {
					m.SetCursor(m.col - 1)
				}
			}
		case key.Matches(msg, m.KeyMap.DeleteCharacterForward):
			if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
				m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
			}
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
		case key.Matches(msg, m.KeyMap.DeleteWordBackward):
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			m.deleteWordLeft()
		case key.Matches(msg, m.KeyMap.DeleteWordForward):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col >= len(m.value[m.row]) {
				m.mergeLineBelow(m.row)
				break
			}
			m.deleteWordRight()
		case key.Matches(msg, m.KeyMap.InsertNewline):
			if m.MaxHeight > 0 && len(m.value) >= m.MaxHeight {
				return m, nil
			}
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			m.splitLine(m.row, m.col)
		case key.Matches(msg, m.KeyMap.LineEnd):
			m.CursorEnd()
		case key.Matches(msg, m.KeyMap.LineStart):
			m.CursorStart()
		case key.Matches(msg, m.KeyMap.CharacterForward):
			m.characterRight()
		case key.Matches(msg, m.KeyMap.LineNext):
			m.CursorDown()
		case key.Matches(msg, m.KeyMap.WordForward):
			m.wordRight()
		case key.Matches(msg, m.KeyMap.Paste):
			return m, Paste
		case key.Matches(msg, m.KeyMap.CharacterBackward):
			m.characterLeft(false /* insideLine */)
		case key.Matches(msg, m.KeyMap.LinePrevious):
			m.CursorUp()
		case key.Matches(msg, m.KeyMap.WordBackward):
			m.wordLeft()
		case key.Matches(msg, m.KeyMap.InputBegin):
			m.moveToBegin()
		case key.Matches(msg, m.KeyMap.InputEnd):
			m.moveToEnd()
		case key.Matches(msg, m.KeyMap.LowercaseWordForward):
			m.lowercaseRight()
		case key.Matches(msg, m.KeyMap.UppercaseWordForward):
			m.uppercaseRight()
		case key.Matches(msg, m.KeyMap.CapitalizeWordForward):
			m.capitalizeRight()
		case key.Matches(msg, m.KeyMap.TransposeCharacterBackward):
			m.transposeLeft()

		default:
			m.insertRunesFromUserInput(msg.Runes)
		}

	case pasteMsg:
		m.insertRunesFromUserInput([]rune(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	vp, cmd := m.viewport.Update(msg)
	m.viewport = &vp
	cmds = append(cmds, cmd)

	newRow, newCol := m.cursorLineNumber(), m.col
	m.Cursor, cmd = m.Cursor.Update(msg)
	if (newRow != oldRow || newCol != oldCol) && m.Cursor.Mode() == cursor.CursorBlink {
		m.Cursor.Blink = false
		cmd = m.Cursor.BlinkCmd()
	}
	cmds = append(cmds, cmd)

	m.repositionView()

	return m, tea.Batch(cmds...)
}

// View 渲染文本区域的当前状态。
func (m Model) View() string {
	if m.Value() == "" && m.row == 0 && m.col == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}
	m.Cursor.TextStyle = m.style.computedCursorLine()

	var (
		s                strings.Builder
		style            lipgloss.Style
		newLines         int
		widestLineNumber int
		lineInfo         = m.LineInfo()
	)

	displayLine := 0
	for l, line := range m.value {
		wrappedLines := m.memoizedWrap(line, m.width)

		if m.row == l {
			style = m.style.computedCursorLine()
		} else {
			style = m.style.computedText()
		}

		for wl, wrappedLine := range wrappedLines {
			prompt := m.getPromptString(displayLine)
			prompt = m.style.computedPrompt().Render(prompt)
			s.WriteString(style.Render(prompt))
			displayLine++

			var ln string
			if m.ShowLineNumbers { //nolint:nestif
				if wl == 0 {
					if m.row == l {
						ln = style.Render(m.style.computedCursorLineNumber().Render(m.formatLineNumber(l + 1)))
						s.WriteString(ln)
					} else {
						ln = style.Render(m.style.computedLineNumber().Render(m.formatLineNumber(l + 1)))
						s.WriteString(ln)
					}
				} else {
					if m.row == l {
						ln = style.Render(m.style.computedCursorLineNumber().Render(m.formatLineNumber(" ")))
						s.WriteString(ln)
					} else {
						ln = style.Render(m.style.computedLineNumber().Render(m.formatLineNumber(" ")))
						s.WriteString(ln)
					}
				}
			}

			// 记录最宽的行号以便稍后填充。
			lnw := lipgloss.Width(ln)
			if lnw > widestLineNumber {
				widestLineNumber = lnw
			}

			strwidth := uniseg.StringWidth(string(wrappedLine))
			padding := m.width - strwidth
			// 如果尾随空格导致行比宽度更宽，我们不应该将其绘制到屏幕上，
			// 因为这会导致行末尾有一个额外的空格，这在显示光标行时看起来不正常。
			if strwidth > m.width {
				// 导致行比宽度更宽的字符保证是一个空格，因为任何其他字符
				// 都会被换行。
				wrappedLine = []rune(strings.TrimSuffix(string(wrappedLine), " "))
				padding -= m.width - strwidth
			}
			if m.row == l && lineInfo.RowOffset == wl {
				s.WriteString(style.Render(string(wrappedLine[:lineInfo.ColumnOffset])))
				if m.col >= len(line) && lineInfo.CharOffset >= m.width {
					m.Cursor.SetChar(" ")
					s.WriteString(m.Cursor.View())
				} else {
					m.Cursor.SetChar(string(wrappedLine[lineInfo.ColumnOffset]))
					s.WriteString(style.Render(m.Cursor.View()))
					s.WriteString(style.Render(string(wrappedLine[lineInfo.ColumnOffset+1:])))
				}
			} else {
				s.WriteString(style.Render(string(wrappedLine)))
			}
			s.WriteString(style.Render(strings.Repeat(" ", max(0, padding))))
			s.WriteRune('\n')
			newLines++
		}
	}

	// 始终至少显示 `m.Height` 行。为此，我们可以在视图中简单地填充一些额外的换行符。
	for i := 0; i < m.height; i++ {
		prompt := m.getPromptString(displayLine)
		prompt = m.style.computedPrompt().Render(prompt)
		s.WriteString(prompt)
		displayLine++

		// 写入缓冲区结束内容
		leftGutter := string(m.EndOfBufferCharacter)
		rightGapWidth := m.Width() - lipgloss.Width(leftGutter) + widestLineNumber
		rightGap := strings.Repeat(" ", max(0, rightGapWidth))
		s.WriteString(m.style.computedEndOfBuffer().Render(leftGutter + rightGap))
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.style.Base.Render(m.viewport.View())
}

// formatLineNumber 根据最大行数动态格式化行号以供显示。
func (m Model) formatLineNumber(x any) string {
	// XXX：最终我们应该使用最大缓冲区高度，但这尚未实现。
	digits := len(strconv.Itoa(m.MaxHeight))
	return fmt.Sprintf(" %*v ", digits, x)
}

func (m Model) getPromptString(displayLine int) (prompt string) {
	prompt = m.Prompt
	if m.promptFunc == nil {
		return prompt
	}
	prompt = m.promptFunc(displayLine)
	pl := uniseg.StringWidth(prompt)
	if pl < m.promptWidth {
		prompt = fmt.Sprintf("%*s%s", m.promptWidth-pl, "", prompt)
	}
	return prompt
}

// placeholderView 返回提示符和占位符视图（如果有）。
func (m Model) placeholderView() string {
	var (
		s     strings.Builder
		p     = m.Placeholder
		style = m.style.computedPlaceholder()
	)

	// 自动换行
	pwordwrap := ansi.Wordwrap(p, m.width, "")
	// 换行（处理无法自动换行的行）
	pwrap := ansi.Hardwrap(pwordwrap, m.width, true)
	// 按换行符分割字符串
	plines := strings.Split(strings.TrimSpace(pwrap), "\n")

	for i := 0; i < m.height; i++ {
		lineStyle := m.style.computedPlaceholder()
		lineNumberStyle := m.style.computedLineNumber()
		if len(plines) > i {
			lineStyle = m.style.computedCursorLine()
			lineNumberStyle = m.style.computedCursorLineNumber()
		}

		// 渲染提示符
		prompt := m.getPromptString(i)
		prompt = m.style.computedPrompt().Render(prompt)
		s.WriteString(lineStyle.Render(prompt))

		// 当启用显示行号时：
		// - 仅渲染光标行的行号
		// - 缩进其他占位符行
		// 这与启用行号的 vim 一致
		if m.ShowLineNumbers {
			var ln string

			switch {
			case i == 0:
				ln = strconv.Itoa(i + 1)
				fallthrough
			case len(plines) > i:
				s.WriteString(lineStyle.Render(lineNumberStyle.Render(m.formatLineNumber(ln))))
			default:
			}
		}

		switch {
		// 第一行
		case i == 0:
			// 第一行的第一个字符作为带有字符的光标
			m.Cursor.TextStyle = m.style.computedPlaceholder()

			ch, rest, _, _ := uniseg.FirstGraphemeClusterInString(plines[0], 0)
			m.Cursor.SetChar(ch)
			s.WriteString(lineStyle.Render(m.Cursor.View()))

			// 第一行的其余部分
			s.WriteString(lineStyle.Render(style.Render(rest)))
		// 剩余行
		case len(plines) > i:
			// 当前行占位符文本
			if len(plines) > i {
				s.WriteString(lineStyle.Render(style.Render(plines[i] + strings.Repeat(" ", max(0, m.width-uniseg.StringWidth(plines[i]))))))
			}
		default:
			// 行缓冲区结束字符
			eob := m.style.computedEndOfBuffer().Render(string(m.EndOfBufferCharacter))
			s.WriteString(eob)
		}

		// 以换行符终止
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.style.Base.Render(m.viewport.View())
}

// Blink 返回光标的闪烁命令。
func Blink() tea.Msg {
	return cursor.Blink()
}

func (m Model) memoizedWrap(runes []rune, width int) [][]rune {
	input := line{runes: runes, width: width}
	if v, ok := m.cache.Get(input); ok {
		return v
	}
	v := wrap(runes, width)
	m.cache.Set(input, v)
	return v
}

// cursorLineNumber 返回光标所在的行号。这考虑了软换行。
func (m Model) cursorLineNumber() int {
	line := 0
	for i := 0; i < m.row; i++ {
		// 计算当前行将被分割成的行数。
		line += len(m.memoizedWrap(m.value[i], m.width))
	}
	line += m.LineInfo().RowOffset
	return line
}

// mergeLineBelow 将光标所在的当前行与下面的行合并。
func (m *Model) mergeLineBelow(row int) {
	if row >= len(m.value)-1 {
		return
	}

	// 要执行合并，我们需要将两行组合起来，然后
	m.value[row] = append(m.value[row], m.value[row+1]...)

	// 将所有行向上移动一行
	for i := row + 1; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// 并且，删除最后一行
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

// mergeLineAbove 将光标所在的当前行与上面的行合并。
func (m *Model) mergeLineAbove(row int) {
	if row <= 0 {
		return
	}

	m.col = len(m.value[row-1])
	m.row = m.row - 1

	// 要执行合并，我们需要将两行组合起来，然后
	m.value[row-1] = append(m.value[row-1], m.value[row]...)

	// 将所有行向上移动一行
	for i := row; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// 并且，删除最后一行
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

func (m *Model) splitLine(row, col int) {
	// 要执行分割，取当前行并保留光标之前的内容，取光标之后的内容
	// 并使其成为下方行的内容，然后将剩余行向下移动一行
	head, tailSrc := m.value[row][:col], m.value[row][col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	m.value = append(m.value[:row+1], m.value[row:]...)

	m.value[row] = head
	m.value[row+1] = tail

	m.col = 0
	m.row++
}

// Paste 是从剪贴板粘贴到文本输入的命令。
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

func wrap(runes []rune, width int) [][]rune {
	var (
		lines  = [][]rune{{}}
		word   = []rune{}
		row    int
		spaces int
	)

	// 对字符进行自动换行
	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 { //nolint:nestif
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			}
		} else {
			// 如果最后一个字符是双宽度字符，那么我们可能无法将其添加到此行，
			// 因为它可能会导致我们超过宽度。
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			if uniseg.StringWidth(string(word))+lastCharLen > width {
				// 如果当前行有任何内容，让我们移动到下一行，
				// 因为当前单词填满了整行。
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		// 我们在行末尾添加一个额外的空格，以考虑前一个软换行行末尾的尾随空格，
		// 这样导航时的行为是一致的，并且我们不需要不断添加边缘来处理换行输入的最后一行。
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(string(' '), n))
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
