package textarea

import (
	"strings"
	"testing"
	"unicode"

	"github.com/MakeNowJust/heredoc"
	"github.com/aymanbagabas/go-udiff"
	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// 测试垂直滚动功能
// 验证文本区域在内容超出可视范围时能否正确滚动显示
func TestVerticalScrolling(t *testing.T) {
	textarea := newTextArea()
	textarea.Prompt = ""
	textarea.ShowLineNumbers = false
	textarea.SetHeight(1)    // 设置文本区域高度为1行
	textarea.SetWidth(20)    // 设置文本区域宽度为20个字符
	textarea.CharLimit = 100 // 设置字符限制为100

	textarea, _ = textarea.Update(nil)

	// 输入一段超长文本，超出文本区域宽度
	input := "This is a really long line that should wrap around the text area."

	// 逐个字符输入文本
	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	// 验证视图是否显示输入的第一行
	if !strings.Contains(view, "This is a really") {
		t.Log(view)
		t.Error("文本区域未正确渲染输入内容")
	}

	// 验证能否通过滚动查看后续内容
	// 逐行向下滚动以查看完整输入内容
	expectedLines := []string{
		"long line that",
		"should wrap around",
		"the text area.",
	}
	for _, line := range expectedLines {
		textarea.viewport.ScrollDown(1) // 向下滚动一行
		view = textarea.View()
		if !strings.Contains(view, line) {
			t.Log(view)
			t.Error("文本区域未正确渲染滚动后的内容")
		}
	}
}

// 测试自动换行溢出处理
// 验证当用户在已填满的文本区域中插入单词导致级联换行时，能否正确处理最后一行的溢出
func TestWordWrapOverflowing(t *testing.T) {
	// 一个有趣的边界情况是：用户输入大量单词填满文本区域后，回到开头插入几个单词，
	// 这会导致级联换行并可能使最后一行溢出。
	//
	// 在这种情况下，如果整个换行完成后最后一行仍然溢出，我们应该阻止用户继续插入单词。
	textarea := newTextArea()

	textarea.SetHeight(3)    // 设置文本区域高度为3行
	textarea.SetWidth(20)    // 设置文本区域宽度为20个字符
	textarea.CharLimit = 500 // 设置字符限制为500

	textarea, _ = textarea.Update(nil)

	// 输入重复的"Testing"单词，填满文本区域
	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	// 逐个字符输入文本
	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View() // 触发视图更新
	}

	// 现在文本区域已被填满
	// 尝试在开头插入单词，看是否会导致最后一行溢出
	textarea.row = 0 // 将光标移到第一行
	textarea.col = 0 // 将光标移到行首

	input = "Testing" // 要插入的单词

	// 逐个字符插入单词
	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View() // 触发视图更新
	}

	// 检查最后一行的宽度是否超过限制
	lastLineWidth := textarea.LineInfo().Width
	if lastLineWidth > 20 {
		t.Log(lastLineWidth)
		t.Log(textarea.View())
		t.Fail() // 如果超过宽度则测试失败
	}
}

// 测试软换行对值的影响
// 验证软换行不会改变文本区域的实际值（仅影响显示）
func TestValueSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(16)    // 设置文本区域宽度为16个字符
	textarea.SetHeight(10)   // 设置文本区域高度为10行
	textarea.CharLimit = 500 // 设置字符限制为500

	textarea, _ = textarea.Update(nil)

	// 输入重复的"Testing"单词，触发软换行
	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	// 逐个字符输入文本
	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View() // 触发视图更新，触发软换行
	}

	// 获取文本区域的实际值
	value := textarea.Value()
	// 验证实际值是否与输入一致（软换行不影响实际值）
	if value != input {
		t.Log(value)
		t.Log(input)
		t.Fatal("文本区域的实际值不正确")
	}
}

// 测试SetValue方法
// 验证SetValue方法能否正确设置文本区域的值，并在设置后正确定位光标
func TestSetValue(t *testing.T) {
	textarea := newTextArea()
	// 设置多行文本，包含三个单词，每行一个
	textarea.SetValue(strings.Join([]string{"Foo", "Bar", "Baz"}, "\n"))

	// 验证光标位置：应该在第2行（索引从0开始），第3列（"Baz"的末尾）
	if textarea.row != 2 && textarea.col != 3 {
		t.Log(textarea.row, textarea.col)
		t.Fatal("插入2个新行后，光标应该位于第2行第3列")
	}

	// 获取文本区域的实际值
	value := textarea.Value()
	// 验证实际值是否与预期一致
	if value != "Foo\nBar\nBaz" {
		t.Fatal("文本区域的值应该是Foo\nBar\nBaz")
	}

	// 验证SetValue方法是否会重置文本区域
	textarea.SetValue("Test") // 设置新值
	value = textarea.Value()  // 获取新值
	if value != "Test" {
		t.Log(value)
		t.Fatal("调用SetValue()时文本区域未正确重置")
	}
}

// 测试InsertString方法
// 验证InsertString方法能否在指定位置正确插入字符串
func TestInsertString(t *testing.T) {
	textarea := newTextArea()

	// 输入初始文本
	input := "foo baz"

	// 逐个字符输入文本
	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	// 将光标移到文本中间（"foo "和"baz"之间）
	textarea.col = 4

	// 在光标位置插入字符串"bar "
	textarea.InsertString("bar ")

	// 获取文本区域的实际值
	value := textarea.Value()
	// 验证插入后的文本是否正确
	if value != "foo bar baz" {
		t.Log(value)
		t.Fatal("InsertString方法应该在foo和baz之间插入bar")
	}
}

// 测试表情符号处理
// 验证文本区域能否正确处理表情符号（双宽度字符）
func TestCanHandleEmoji(t *testing.T) {
	textarea := newTextArea()
	// 输入单个奶茶表情符号
	input := "🧋"

	// 逐个字符输入文本
	for _, k := range []rune(input) {
		textarea, _ = textarea.Update(keyPress(k))
	}

	// 获取文本区域的实际值
	value := textarea.Value()
	// 验证表情符号是否正确插入
	if value != input {
		t.Log(value)
		t.Fatal("应该正确插入表情符号")
	}

	// 输入三个奶茶表情符号
	input = "🧋🧋🧋"

	// 使用SetValue方法设置值
	textarea.SetValue(input)

	// 获取文本区域的实际值
	value = textarea.Value()
	// 验证表情符号是否正确插入
	if value != input {
		t.Log(value)
		t.Fatal("应该正确插入表情符号")
	}

	// 验证光标位置：应该在第3个字符（表情符号）的末尾
	if textarea.col != 3 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于第3个字符的位置")
	}

	// 验证字符偏移量：每个表情符号占2个字符位置，3个表情符号占6个字符位置
	if charOffset := textarea.LineInfo().CharOffset; charOffset != 6 {
		t.Log(charOffset)
		t.Fatal("光标应该位于第6个字符的位置")
	}
}

// 测试垂直导航时光标水平位置保持
// 验证在垂直导航（上下箭头）时，光标能否保持相同的视觉列位置（考虑双宽度字符）
func TestVerticalNavigationKeepsCursorHorizontalPosition(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(20) // 设置文本区域宽度为20个字符

	// 设置包含双宽度字符（中文）和单宽度字符（英文）的文本
	textarea.SetValue(strings.Join([]string{"你好你好", "Hello"}, "\n"))

	// 将光标移到第一行的第2列
	textarea.row = 0
	textarea.col = 2

	// 你好|你好
	// Hell|o
	// 1234|

	// 假设我们的光标在第一行的管道位置。
	// 我们按下向下箭头键移动到下一行。
	// 问题是，如果我们保持光标在相同的列，光标会跳到`e`之后。
	//
	// 你好|你好
	// He|llo
	//
	// 但这是错误的，因为视觉上我们在第4个字符的位置，
	// 因为第一行包含双宽度字符。
	// 我们希望光标保持在相同的视觉列。
	//
	// 你好|你好
	// Hell|o
	//
	// 这个测试通过确保列偏移从2 -> 4，来验证光标保持在相同的视觉列。

	// 获取当前行信息
	lineInfo := textarea.LineInfo()
	// 验证光标位置：应该在第4个字符（因为第一行有两个双宽度字符）
	if lineInfo.CharOffset != 4 || lineInfo.ColumnOffset != 2 {
		t.Log(lineInfo.CharOffset)
		t.Log(lineInfo.ColumnOffset)
		t.Fatal("光标应该位于第4个字符的位置，因为第一行有两个双宽度字符。")
	}

	// 发送向下箭头键消息
	downMsg := tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(downMsg)

	// 获取新的行信息
	lineInfo = textarea.LineInfo()
	// 验证光标位置：应该在第4个字符（因为我们从第一行下来）
	if lineInfo.CharOffset != 4 || lineInfo.ColumnOffset != 4 {
		t.Log(lineInfo.CharOffset)
		t.Log(lineInfo.ColumnOffset)
		t.Fatal("光标应该位于第4个字符的位置，因为我们从第一行下来。")
	}
}

// 测试垂直导航时记住水平位置
// 验证在垂直导航时能否记住最后停留的水平位置，以及在水平移动后能否重置该位置
func TestVerticalNavigationShouldRememberPositionWhileTraversing(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(40) // 设置文本区域宽度为40个字符

	// 假设我们有一个包含以下内容的文本区域：
	//
	// Hello
	// World
	// This is a long line.
	//
	// 如果我们在最后一行的末尾并向上移动，应该到达第二行的末尾。
	// 如果再次向上移动，应该到达第一行的末尾。
	// 但如果我们再次向下移动两次，应该回到最后一行的末尾，
	// 而不是最后一行的第5个字符（第二行的长度）的位置。
	//
	// 换句话说，我们在垂直导航时应该记住最后停留的水平位置。

	// 设置多行文本，包含不同长度的行
	textarea.SetValue(strings.Join([]string{"Hello", "World", "This is a long line."}, "\n"))

	// 验证光标位置：应该在最后一行的第20个字符的位置
	if textarea.col != 20 || textarea.row != 2 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于最后一行的第20个字符的位置")
	}

	// 向上移动一行
	upMsg := tea.KeyMsg{Type: tea.KeyUp, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(upMsg)

	// 验证光标位置：应该在第二行的第5个字符的位置（"World"的末尾）
	if textarea.col != 5 || textarea.row != 1 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于第二行的第5个字符的位置")
	}

	// 再次向上移动一行
	textarea, _ = textarea.Update(upMsg)

	// 验证光标位置：应该在第一行的第5个字符的位置（"Hello"的末尾）
	if textarea.col != 5 || textarea.row != 0 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于第一行的第5个字符的位置")
	}

	// 向下移动两行
	downMsg := tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(downMsg)
	textarea, _ = textarea.Update(downMsg)

	// 验证光标位置：应该回到最后一行的第20个字符的位置
	if textarea.col != 20 || textarea.row != 2 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于最后一行的第20个字符的位置")
	}

	// 现在，为了正确的行为，如果我们左右移动光标，应该忘记（重置）保存的水平位置。
	// 因为我们假设用户希望将光标保持在当前的水平位置。这是大多数文本区域的工作方式。

	// 向上移动一行
	textarea, _ = textarea.Update(upMsg)
	// 向左移动一个字符
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(leftMsg)

	// 验证光标位置：应该在第二行的第4个字符的位置
	if textarea.col != 4 || textarea.row != 1 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于第二行的第4个字符的位置")
	}

	// 现在向下移动应该保持在第4列，因为我们已经向左移动并重置了水平位置的保存状态。
	textarea, _ = textarea.Update(downMsg)
	// 验证光标位置：应该在最后一行的第4个字符的位置
	if textarea.col != 4 || textarea.row != 2 {
		t.Log(textarea.col)
		t.Fatal("光标应该位于最后一行的第4个字符的位置")
	}
}

// 测试视图渲染
// 验证文本区域在不同配置下能否正确渲染视图
func TestView(t *testing.T) {
	t.Parallel() // 并行运行测试

	// 定义期望结果结构体
	type want struct {
		view      string // 期望的视图内容
		cursorRow int    // 期望的光标行
		cursorCol int    // 期望的光标列
	}

	// 定义测试用例
	tests := []struct {
		name      string            // 测试名称
		modelFunc func(Model) Model // 模型配置函数
		want      want              // 期望结果
	}{
		{
			name: "placeholder", // 占位符测试
			want: want{
				view: heredoc.Doc(`
					>   1 Hello, World!
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "single line", // 单行文本测试
			modelFunc: func(m Model) Model {
				m.SetValue("the first line") // 设置单行文本

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,  // 期望光标在第0行
				cursorCol: 14, // 期望光标在第14列
			},
		},
		{
			name: "multiple lines",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line\nthe second line\nthe third line")

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					>   2 the second line
					>   3 the third line
					>
					>
					>
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line without line numbers",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line")
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multipline lines without line numbers",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					> the second line
					> the third line
					>
					>
					>
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line and custom end of buffer character",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line")
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					> *
					> *
					> *
					> *
					> *
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines and custom end of buffer character",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					>   2 the second line
					>   3 the third line
					> *
					> *
					> *
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line without line numbers and custom end of buffer character",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line")
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					> *
					> *
					> *
					> *
					> *
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines without line numbers and custom end of buffer character",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					> the second line
					> the third line
					> *
					> *
					> *
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line and custom prompt",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line")
				m.Prompt = "* "

				return m
			},
			want: want{
				view: heredoc.Doc(`
					*   1 the first line
					*
					*
					*
					*
					*
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines and custom prompt",
			modelFunc: func(m Model) Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.Prompt = "* "

				return m
			},
			want: want{
				view: heredoc.Doc(`
					*   1 the first line
					*   2 the second line
					*   3 the third line
					*
					*
					*
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "type single line",
			modelFunc: func(m Model) Model {
				input := "foo"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "type multiple lines",
			modelFunc: func(m Model) Model {
				input := "foo\nbar\nbaz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo
					>   2 bar
					>   3 baz
					>
					>
					>
				`),
				cursorRow: 2,
				cursorCol: 3,
			},
		},
		{
			name: "softwrap",
			modelFunc: func(m Model) Model {
				m.ShowLineNumbers = false
				m.Prompt = ""
				m.SetWidth(5)

				input := "foo bar baz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					foo
					bar
					baz



				`),
				cursorRow: 2,
				cursorCol: 3,
			},
		},
		{
			name: "single line character limit",
			modelFunc: func(m Model) Model {
				m.CharLimit = 7

				input := "foo bar baz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo bar
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 7,
			},
		},
		{
			name: "multiple lines character limit",
			modelFunc: func(m Model) Model {
				m.CharLimit = 19

				input := "foo bar baz\nfoo bar baz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo bar baz
					>   2 foo bar
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 7,
			},
		},
		{
			name: "set width",
			modelFunc: func(m Model) Model {
				m.SetWidth(10)

				input := "12"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 2,
			},
		},
		{
			name: "set width max length text minus one",
			modelFunc: func(m Model) Model {
				m.SetWidth(10)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 123
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width max length text",
			modelFunc: func(m Model) Model {
				m.SetWidth(10)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width max length text plus one",
			modelFunc: func(m Model) Model {
				m.SetWidth(10)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>     5
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width set max width minus one",
			modelFunc: func(m Model) Model {
				m.MaxWidth = 10
				m.SetWidth(11)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 123
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width set max width",
			modelFunc: func(m Model) Model {
				m.MaxWidth = 10
				m.SetWidth(11)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width set max width plus one",
			modelFunc: func(m Model) Model {
				m.MaxWidth = 10
				m.SetWidth(11)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>     5
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width min width minus one",
			modelFunc: func(m Model) Model {
				m.SetWidth(6)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1
					>     2
					>     3
					>
					>
					>
				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width",
			modelFunc: func(m Model) Model {
				m.SetWidth(7)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1
					>     2
					>     3
					>
					>
					>
				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width no line numbers",
			modelFunc: func(m Model) Model {
				m.ShowLineNumbers = false
				m.SetWidth(0)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 1
					> 2
					> 3
					>
					>
					>
				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width no line numbers no prompt",
			modelFunc: func(m Model) Model {
				m.ShowLineNumbers = false
				m.Prompt = ""
				m.SetWidth(0)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					1
					2
					3



				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width plus one",
			modelFunc: func(m Model) Model {
				m.SetWidth(8)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12
					>     3
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width without line numbers max length text minus one",
			modelFunc: func(m Model) Model {
				m.ShowLineNumbers = false
				m.SetWidth(6)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width without line numbers max length text",
			modelFunc: func(m Model) Model {
				m.ShowLineNumbers = false
				m.SetWidth(6)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 1234
					>
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width without line numbers max length text plus one",
			modelFunc: func(m Model) Model {
				m.ShowLineNumbers = false
				m.SetWidth(6)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 1234
					> 5
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width with style",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.SetWidth(12)

				input := "1"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 1   │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 1,
			},
		},
		{
			name: "set width with style max width minus one",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.SetWidth(12)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 123 │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width with style max width",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.SetWidth(12)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 1234│
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width with style max width plus one",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.SetWidth(12)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 1234│
					│>     5   │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width without line numbers with style",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "123456"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 123456  │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 6,
			},
		},
		{
			name: "set width without line numbers with style max width minus one",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "1234567"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 1234567 │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 7,
			},
		},
		{
			name: "set width without line numbers with style max width",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "12345678"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 12345678│
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width without line numbers with style max width plus one",
			modelFunc: func(m Model) Model {
				m.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "123456789"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 12345678│
					│> 9       │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "placeholder min width",
			modelFunc: func(m Model) Model {
				m.SetWidth(0)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 H
					>     e
					>     l
					>     l
					>     o
					>     ,
				`),
			},
		},
		{
			name: "placeholder single line",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					>
					>
					>
					>
					>
					`),
			},
		},
		{
			name: "placeholder multiple lines",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					> placeholder the second line
					> placeholder the third line
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line with line numbers",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = true

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines with line numbers",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = true

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					>     placeholder the second line
					>     placeholder the third line
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line with end of buffer character",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					> *
					> *
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder multiple lines with with end of buffer character",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					> placeholder the second line
					> placeholder the third line
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder single line with line numbers and end of buffer character",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = true
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					> *
					> *
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder multiple lines with line numbers and end of buffer character",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = true
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					>     placeholder the second line
					>     placeholder the third line
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line that is longer than the max width"
				m.SetWidth(40)
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line that is
					> longer than the max width
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line that is longer than the max width\nplaceholder the second line that is longer than the max width"
				m.ShowLineNumbers = false
				m.SetWidth(40)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line that is
					> longer than the max width
					> placeholder the second line that is
					> longer than the max width
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width with line numbers",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line that is longer than the max width"
				m.ShowLineNumbers = true
				m.SetWidth(40)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line that is
					>     longer than the max width
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width with line numbers",
			modelFunc: func(m Model) Model {
				m.Placeholder = "placeholder the first line that is longer than the max width\nplaceholder the second line that is longer than the max width"
				m.ShowLineNumbers = true
				m.SetWidth(40)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line that is
					>     longer than the max width
					>     placeholder the second line that
					>     is longer than the max width
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width at limit",
			modelFunc: func(m Model) Model {
				m.Placeholder = "123456789012345678"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width at limit plus one",
			modelFunc: func(m Model) Model {
				m.Placeholder = "1234567890123456789"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					> 9
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width with line numbers at limit",
			modelFunc: func(m Model) Model {
				m.Placeholder = "12345678901234"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width with line numbers at limit plus one",
			modelFunc: func(m Model) Model {
				m.Placeholder = "123456789012345"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>     5
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width at limit",
			modelFunc: func(m Model) Model {
				m.Placeholder = "123456789012345678\n123456789012345678"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					> 123456789012345678
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width at limit plus one",
			modelFunc: func(m Model) Model {
				m.Placeholder = "1234567890123456789\n1234567890123456789"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					> 9
					> 123456789012345678
					> 9
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width with line numbers at limit",
			modelFunc: func(m Model) Model {
				m.Placeholder = "12345678901234\n12345678901234"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>     12345678901234
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width with line numbers at limit plus one",
			modelFunc: func(m Model) Model {
				m.Placeholder = "123456789012345\n123456789012345"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>     5
					>     12345678901234
					>     5
					>
					>
				`),
			},
		},
		{
			name: "placeholder chinese character",
			modelFunc: func(m Model) Model {
				m.Placeholder = "输入消息..."
				m.ShowLineNumbers = true
				m.SetWidth(20)
				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 输入消息...
					>
					>
					>
					>
					>

				`),
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			textarea := newTextArea()

			if tt.modelFunc != nil {
				textarea = tt.modelFunc(textarea)
			}

			view := stripString(textarea.View())
			wantView := stripString(tt.want.view)

			if view != wantView {
				t.Log(udiff.Unified("expected", "got", wantView, view))
				t.Fatalf("Want:\n%v\nGot:\n%v\n", wantView, view)
			}

			cursorRow := textarea.cursorLineNumber()
			cursorCol := textarea.LineInfo().ColumnOffset
			if tt.want.cursorRow != cursorRow || tt.want.cursorCol != cursorCol {
				format := "Want cursor at row: %v, col: %v Got: row: %v col: %v\n"
				t.Fatalf(format, tt.want.cursorRow, tt.want.cursorCol, cursorRow, cursorCol)
			}
		})
	}
}

func newTextArea() Model {
	textarea := New()

	textarea.Prompt = "> "
	textarea.Placeholder = "Hello, World!"

	textarea.Focus()

	textarea, _ = textarea.Update(nil)

	return textarea
}

func keyPress(key rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}, Alt: false}
}

func sendString(m Model, str string) Model {
	for _, k := range []rune(str) {
		m, _ = m.Update(keyPress(k))
	}

	return m
}

func stripString(str string) string {
	s := ansi.Strip(str)
	ss := strings.Split(s, "\n")

	var lines []string
	for _, l := range ss {
		trim := strings.TrimRightFunc(l, unicode.IsSpace)
		if trim != "" {
			lines = append(lines, trim)
		}
	}

	return strings.Join(lines, "\n")
}
