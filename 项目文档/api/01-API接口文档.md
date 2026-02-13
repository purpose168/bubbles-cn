# Bubbles-CN API 接口文档

## 概述

本文档详细说明了 Bubbles-CN 各组件的 API 接口定义，包括接口路径、请求方法、参数说明、返回格式及错误码定义。Bubbles-CN 是一个 Go 语言库，其 API 主要通过函数和方法调用提供。

## 接口分类

### 按模块分类

| 模块 | 接口类型 | 说明 |
|------|---------|------|
| **key** | 函数接口 | 键绑定管理 |
| **cursor** | 方法接口 | 光标管理 |
| **runeutil** | 函数接口 | 字符工具 |
| **textinput** | 方法接口 | 文本输入 |
| **textarea** | 方法接口 | 文本区域 |
| **spinner** | 方法接口 | 加载指示器 |
| **progress** | 方法接口 | 进度条 |
| **table** | 方法接口 | 表格 |
| **list** | 方法接口 | 列表 |
| **viewport** | 方法接口 | 视口 |
| **paginator** | 方法接口 | 分页器 |
| **filepicker** | 方法接口 | 文件选择器 |
| **help** | 方法接口 | 帮助 |
| **timer** | 方法接口 | 计时器 |
| **stopwatch** | 方法接口 | 秒表 |

## 基础模块接口

### 1. Key 模块

#### 1.1 NewBinding

创建新的键绑定。

**函数签名**:
```go
func NewBinding(opts ...Option) Binding
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| opts | []Option | 否 | 配置选项 |

**选项类型**:
| 选项 | 类型 | 说明 |
|------|------|------|
| WithKeys | func(...string) Option | 设置绑定的键 |
| WithHelp | func(string, string) Option | 设置帮助文本 |
| WithDisabled | func() Option | 禁用绑定 |
| WithEnabled | func() Option | 启用绑定 |

**返回值**:
```go
type Binding struct {
    keys      []string  // 绑定的键列表
    help      string    // 帮助文本
    enabled   bool      // 是否启用
    disabled  bool      // 是否禁用
}
```

**示例**:
```go
binding := key.NewBinding(
    key.WithKeys("k", "up"),
    key.WithHelp("↑/k", "向上移动"),
)
```

#### 1.2 Matches

检查按键消息是否匹配任何绑定。

**函数签名**:
```go
func Matches(msg tea.KeyMsg, bindings ...Binding) bool
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| msg | tea.KeyMsg | 是 | 按键消息 |
| bindings | ...Binding | 是 | 要检查的绑定列表 |

**返回值**:
| 类型 | 说明 |
|------|------|
| bool | 如果匹配返回 true，否则返回 false |

**示例**:
```go
switch msg := msg.(type) {
case tea.KeyMsg:
    if key.Matches(msg, m.keyMap.Up) {
        // 处理向上移动
    }
}
```

### 2. Cursor 模块

#### 2.1 New

创建新的光标模型。

**函数签名**:
```go
func New() Model
```

**返回值**:
```go
type Model struct {
    Hidden     bool          // 是否隐藏光标
    BlinkSpeed time.Duration // 闪烁速度
    Character  string        // 光标字符
}
```

**示例**:
```go
cursor := cursor.New()
cursor.BlinkSpeed = 500 * time.Millisecond
```

#### 2.2 View

渲染光标。

**函数签名**:
```go
func (m Model) View() string
```

**返回值**:
| 类型 | 说明 |
|------|------|
| string | 光标的 ANSI 转义序列 |

**示例**:
```go
cursorView := cursor.View()
```

### 3. Runeutil 模块

#### 3.1 StringWidth

计算字符串的显示宽度。

**函数签名**:
```go
func StringWidth(s string) int
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| s | string | 是 | 要计算的字符串 |

**返回值**:
| 类型 | 说明 |
|------|------|
| int | 字符串的显示宽度（考虑全角字符） |

**示例**:
```go
width := runeutil.StringWidth("你好世界") // 返回 8
```

#### 3.2 Truncate

截断字符串到指定宽度。

**函数签名**:
```go
func Truncate(s string, maxWidth int) string
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| s | string | 是 | 要截断的字符串 |
| maxWidth | int | 是 | 最大宽度 |

**返回值**:
| 类型 | 说明 |
|------|------|
| string | 截断后的字符串 |

**示例**:
```go
truncated := runeutil.Truncate("这是一个很长的文本", 10)
```

## 输入模块接口

### 4. TextInput 模块

#### 4.1 New

创建新的文本输入模型。

**函数签名**:
```go
func New() Model
```

**返回值**:
```go
type Model struct {
    Err            error         // 验证错误
    Prompt         string        // 提示符
    Placeholder    string        // 占位符文本
    EchoMode       EchoMode      // 回显模式
    EchoCharacter  rune          // 回显字符
    Cursor         cursor.Model  // 光标模型
    Value          string        // 当前值
    Width          int           // 宽度
    CharLimit      int           // 字符限制
    Validate       ValidateFunc  // 验证函数
}
```

**示例**:
```go
textInput := textinput.New()
textInput.Placeholder = "请输入用户名"
```

#### 4.2 Update

更新文本输入模型。

**函数签名**:
```go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| msg | tea.Msg | 是 | 消息 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 更新后的模型 |
| tea.Cmd | 命令 |

**示例**:
```go
var (
    textInput textinput.Model
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    textInput, cmd = textInput.Update(msg)
    return m, cmd
}
```

#### 4.3 View

渲染文本输入。

**函数签名**:
```go
func (m Model) View() string
```

**返回值**:
| 类型 | 说明 |
|------|------|
| string | 渲染后的文本输入 |

**示例**:
```go
view := textInput.View()
```

#### 4.4 Focus

聚焦文本输入。

**函数签名**:
```go
func (m Model) Focus() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 聚焦后的模型 |

**示例**:
```go
textInput = textInput.Focus()
```

#### 4.5 Blur

取消聚焦文本输入。

**函数签名**:
```go
func (m Model) Blur() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 取消聚焦后的模型 |

**示例**:
```go
textInput = textinput.Blur()
```

#### 4.6 SetValue

设置文本输入的值。

**函数签名**:
```go
func (m Model) SetValue(s string) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| s | string | 是 | 要设置的值 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置值后的模型 |

**示例**:
```go
textInput = textInput.SetValue("默认值")
```

### 5. TextArea 模块

#### 5.1 New

创建新的文本区域模型。

**函数签名**:
```go
func New() Model
```

**返回值**:
```go
type Model struct {
    Value           string        // 当前值
    Prompt          string        // 提示符
    Placeholder     string        // 占位符文本
    Cursor          cursor.Model  // 光标模型
    Width           int           // 宽度
    Height          int           // 高度
    CharLimit       int           // 字符限制
    ShowLineNumbers bool          // 是否显示行号
    KeyMap          KeyMap        // 键绑定
}
```

**示例**:
```go
textarea := textarea.New()
textarea.Placeholder = "请输入消息..."
textarea.SetHeight(10)
```

#### 5.2 SetWidth

设置文本区域宽度。

**函数签名**:
```go
func (m Model) SetWidth(w int) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| w | int | 是 | 宽度 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置宽度后的模型 |

**示例**:
```go
textarea = textarea.SetWidth(80)
```

#### 5.3 SetHeight

设置文本区域高度。

**函数签名**:
```go
func (m Model) SetHeight(h int) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| h | int | 是 | 高度 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置高度后的模型 |

**示例**:
```go
textarea = textarea.SetHeight(20)
```

## 显示模块接口

### 6. Spinner 模块

#### 6.1 New

创建新的加载指示器模型。

**函数签名**:
```go
func New() Model
```

**返回值**:
```go
type Model struct {
    Type  SpinnerType    // 加载指示器类型
    Style lipgloss.Style // 样式
}
```

**示例**:
```go
spinner := spinner.New()
spinner.Type = spinner.Dot
```

#### 6.2 View

渲染加载指示器。

**函数签名**:
```go
func (m Model) View() string
```

**返回值**:
| 类型 | 说明 |
|------|------|
| string | 渲染后的加载指示器 |

**示例**:
```go
view := spinner.View()
```

#### 6.3 Tick

创建定时器命令。

**函数签名**:
```go
func (m Model) Tick() tea.Cmd
```

**返回值**:
| 类型 | 说明 |
|------|------|
| tea.Cmd | 定时器命令 |

**示例**:
```go
cmd := spinner.Tick()
```

### 7. Progress 模块

#### 7.1 New

创建新的进度条模型。

**函数签名**:
```go
func New(opts ...Option) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| opts | []Option | 否 | 配置选项 |

**选项类型**:
| 选项 | 类型 | 说明 |
|------|------|------|
| WithDefaultGradient | func() Option | 使用默认渐变 |
| WithSolidPercentage | func(string) Option | 设置纯色百分比 |
| WithWidth | func(int) Option | 设置宽度 |
| WithoutPercentage | func() Option | 不显示百分比 |

**返回值**:
```go
type Model struct {
    Percent         float64        // 进度百分比 (0.0-1.0)
    Width           int            // 宽度
    Full            *lipgloss.Style // 填充样式
    Empty           *lipgloss.Style // 空样式
    Head            *lipgloss.Style // 头部样式
    ShowPercentage  bool           // 是否显示百分比
}
```

**示例**:
```go
progress := progress.New(
    progress.WithDefaultGradient(),
    progress.WithWidth(40),
)
```

#### 7.2 SetPercent

设置进度百分比。

**函数签名**:
```go
func (m Model) SetPercent(p float64) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| p | float64 | 是 | 进度百分比 (0.0-1.0) |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置进度后的模型 |

**示例**:
```go
progress = progress.SetPercent(0.5) // 50%
```

#### 7.3 IncrPercent

增加进度百分比。

**函数签名**:
```go
func (m Model) IncrPercent(delta float64) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| delta | float64 | 是 | 增加的百分比 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 增加进度后的模型 |

**示例**:
```go
progress = progress.IncrPercent(0.1) // 增加 10%
```

### 8. Table 模块

#### 8.1 New

创建新的表格模型。

**函数签名**:
```go
func New(cols []Column) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| cols | []Column | 是 | 列定义 |

**返回值**:
```go
type Model struct {
    KeyMap  KeyMap      // 键绑定
    Help    help.Model  // 帮助模型
    cols    []Column    // 列定义
    rows    []Row       // 行数据
    cursor  int         // 光标位置
    focus   bool        // 是否聚焦
    styles  Styles      // 样式
}

type Column struct {
    Title string // 列标题
    Width int    // 列宽度
}

type Row []string
```

**示例**:
```go
columns := []table.Column{
    {Title: "名称", Width: 20},
    {Title: "大小", Width: 10},
    {Title: "修改时间", Width: 20},
}
t := table.New(columns)
```

#### 8.2 SetRows

设置表格行数据。

**函数签名**:
```go
func (m Model) SetRows(rows []Row) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| rows | []Row | 是 | 行数据 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置行数据后的模型 |

**示例**:
```go
rows := []table.Row{
    {"文件1.txt", "1024", "2024-01-01"},
    {"文件2.txt", "2048", "2024-01-02"},
}
t = t.SetRows(rows)
```

#### 8.3 SetCursor

设置光标位置。

**函数签名**:
```go
func (m Model) SetCursor(i int) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| i | int | 是 | 光标位置（行索引） |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置光标后的模型 |

**示例**:
```go
t = t.SetCursor(0)
```

### 9. List 模块

#### 9.1 New

创建新的列表模型。

**函数签名**:
```go
func New(opts ...Option) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| opts | []Option | 否 | 配置选项 |

**选项类型**:
| 选项 | 类型 | 说明 |
|------|------|------|
| WithItems | func([]Item) Option | 设置列表项 |
| WithTitle | func(string) Option | 设置标题 |
| WithShowHelp | func(bool) Option | 是否显示帮助 |
| WithShowPagination | func(bool) Option | 是否显示分页 |
| WithFilteringEnabled | func(bool) Option | 是否启用过滤 |

**返回值**:
```go
type Model struct {
    KeyMap       key.KeyMap   // 键绑定
    Title        string       // 标题
    Items        []Item       // 列表项
    Cursor       int          // 光标位置
    SelectedItem Item         // 选中的项
    FilterState  FilterState  // 过滤状态
    Paginator    paginator.Model // 分页器
    ShowHelp     bool         // 是否显示帮助
}

type Item interface {
    FilterValue() string
}
```

**示例**:
```go
items := []list.Item{
    list.NewDefaultItem("项目1"),
    list.NewDefaultItem("项目2"),
    list.NewDefaultItem("项目3"),
}
l := list.New(
    list.WithItems(items),
    list.WithTitle("选择项目"),
)
```

#### 9.2 SetItems

设置列表项。

**函数签名**:
```go
func (m Model) SetItems(i []Item) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| i | []Item | 是 | 列表项 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置列表项后的模型 |

**示例**:
```go
items := []list.Item{
    list.NewDefaultItem("选项1"),
    list.NewDefaultItem("选项2"),
}
l = l.SetItems(items)
```

#### 9.3 SetDelegate

设置项目委托。

**函数签名**:
```go
func (m Model) SetDelegate(d ItemDelegate) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| d | ItemDelegate | 是 | 项目委托 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置委托后的模型 |

**示例**:
```go
delegate := list.NewDefaultDelegate()
l = l.SetDelegate(delegate)
```

## 导航模块接口

### 10. Viewport 模块

#### 10.1 New

创建新的视口模型。

**函数签名**:
```go
func New(width, height int) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| width | int | 是 | 视口宽度 |
| height | int | 是 | 视口高度 |

**返回值**:
```go
type Model struct {
    KeyMap  KeyMap       // 键绑定
    YOffset int          // Y 偏移量
    YPosition int        // Y 位置
    Height  int          // 高度
    Width   int          // 宽度
    HighPerformanceViewport bool // 高性能模式
}
```

**示例**:
```go
viewport := viewport.New(80, 20)
```

#### 10.2 SetContent

设置视口内容。

**函数签名**:
```go
func (m Model) SetContent(s string) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| s | string | 是 | 内容 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置内容后的模型 |

**示例**:
```go
viewport = viewport.SetContent(longText)
```

#### 10.3 GotoTop

跳转到顶部。

**函数签名**:
```go
func (m Model) GotoTop() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 跳转后的模型 |

**示例**:
```go
viewport = viewport.GotoTop()
```

#### 10.4 GotoBottom

跳转到底部。

**函数签名**:
```go
func (m Model) GotoBottom() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 跳转后的模型 |

**示例**:
```go
viewport = viewport.GotoBottom()
```

### 11. Paginator 模块

#### 11.1 New

创建新的分页器模型。

**函数签名**:
```go
func New() Model
```

**返回值**:
```go
type Model struct {
    Type          Type           // 分页类型
    PerPage       int            // 每页项数
    TotalPages    int            // 总页数
    Page          int            // 当前页
    ActiveDot     lipgloss.Style // 激活点样式
    InactiveDot   lipgloss.Style // 非激活点样式
}

type Type int

const (
    Dots Type = iota // 点样式
    Arabic           // 阿拉伯数字样式
)
```

**示例**:
```go
paginator := paginator.New()
paginator.PerPage = 10
```

#### 11.2 SetTotalPages

设置总页数。

**函数签名**:
```go
func (m Model) SetTotalPages(n int) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| n | int | 是 | 总页数 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置总页数后的模型 |

**示例**:
```go
paginator = paginator.SetTotalPages(5)
```

#### 11.3 NextPage

下一页。

**函数签名**:
```go
func (m Model) NextPage() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 下一页后的模型 |

**示例**:
```go
paginator = paginator.NextPage()
```

#### 11.4 PrevPage

上一页。

**函数签名**:
```go
func (m Model) PrevPage() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 上一页后的模型 |

**示例**:
```go
paginator = paginator.PrevPage()
```

### 12. FilePicker 模块

#### 12.1 New

创建新的文件选择器模型。

**函数签名**:
```go
func New() Model
```

**返回值**:
```go
type Model struct {
    CurrentDirectory string        // 当前目录
    SelectedFiles    []string      // 选中的文件
    ShowPermissions  bool           // 是否显示权限
    ShowSize         bool           // 是否显示大小
    ShowHidden       bool           // 是否显示隐藏文件
    AutoHeight       bool           // 自动高度
    Cursor           cursor.Model   // 光标模型
    FileAllowed      FileFilter     // 文件过滤器
    DirAllowed       FileFilter     // 目录过滤器
}

type FileFilter func(os.FileInfo) bool
```

**示例**:
```go
filepicker := filepicker.New()
filepicker.CurrentDirectory, _ = os.Getwd()
```

#### 12.2 SetCurrentDirectory

设置当前目录。

**函数签名**:
```go
func (m Model) SetCurrentDirectory(path string) (Model, error)
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| path | string | 是 | 目录路径 |

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 设置目录后的模型 |
| error | 错误信息 |

**示例**:
```go
filepicker, err := filepicker.SetCurrentDirectory("/home/user")
```

#### 12.3 SelectedFile

获取选中的文件。

**函数签名**:
```go
func (m Model) SelectedFile() string
```

**返回值**:
| 类型 | 说明 |
|------|------|
| string | 选中的文件路径 |

**示例**:
```go
selectedFile := filepicker.SelectedFile()
```

## 时间模块接口

### 13. Timer 模块

#### 13.1 New

创建新的计时器模型。

**函数签名**:
```go
func New(opts ...Option) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| opts | []Option | 否 | 配置选项 |

**选项类型**:
| 选项 | 类型 | 说明 |
|------|------|------|
| WithInterval | func(time.Duration) Option | 设置更新间隔 |
| WithTimeout | func(time.Duration) Option | 设置超时时间 |

**返回值**:
```go
type Model struct {
    Timeout      time.Duration   // 超时时间
    Interval     time.Duration   // 更新间隔
    StartTime    time.Time       // 开始时间
    Style        lipgloss.Style  // 样式
}
```

**示例**:
```go
timer := timer.New(
    timer.WithInterval(time.Second),
    timer.WithTimeout(5*time.Minute),
)
```

#### 13.2 Start

启动计时器。

**函数签名**:
```go
func (m Model) Start() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 启动后的模型 |

**示例**:
```go
timer = timer.Start()
```

#### 13.3 Stop

停止计时器。

**函数签名**:
```go
func (m Model) Stop() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 停止后的模型 |

**示例**:
```go
timer = timer.Stop()
```

### 14. Stopwatch 模块

#### 14.1 New

创建新的秒表模型。

**函数签名**:
```go
func New(opts ...Option) Model
```

**参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| opts | []Option | 否 | 配置选项 |

**选项类型**:
| 选项 | 类型 | 说明 |
|------|------|------|
| WithInterval | func(time.Duration) Option | 设置更新间隔 |

**返回值**:
```go
type Model struct {
    Interval     time.Duration   // 更新间隔
    Elapsed      time.Duration   // 已经过的时间
    StartTime    time.Time       // 开始时间
    Style        lipgloss.Style  // 样式
}
```

**示例**:
```go
stopwatch := stopwatch.New(
    stopwatch.WithInterval(time.Millisecond),
)
```

#### 14.2 Start

启动秒表。

**函数签名**:
```go
func (m Model) Start() Model
```

**返回值**:
| 类型 | 说明 |
|------|------|
| Model | 启动后的模型 |

**示例**:
```go
stopwatch = stopwatch.Start()
```

## 错误码定义

### 通用错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|---------|
| **ERR_INVALID_INPUT** | 无效输入 | 检查输入参数 |
| **ERR_OUT_OF_RANGE** | 超出范围 | 检查数值范围 |
| **ERR_NOT_FOUND** | 未找到 | 检查资源是否存在 |
| **ERR_PERMISSION_DENIED** | 权限拒绝 | 检查文件权限 |
| **ERR_TIMEOUT** | 超时 | 增加超时时间或优化操作 |

### TextInput 错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|---------|
| **ERR_VALIDATION_FAILED** | 验证失败 | 检查验证函数 |
| **ERR_CHAR_LIMIT_EXCEEDED** | 超过字符限制 | 增加字符限制或减少输入 |

### FilePicker 错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|---------|
| **ERR_DIRECTORY_NOT_FOUND** | 目录不存在 | 检查目录路径 |
| **ERR_PERMISSION_DENIED** | 权限拒绝 | 检查文件权限 |
| **ERR_NOT_A_DIRECTORY** | 不是目录 | 检查路径是否为目录 |

### Viewport 错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|---------|
| **ERR_INVALID_CONTENT** | 无效内容 | 检查内容格式 |
| **ERR_CONTENT_TOO_LONG** | 内容过长 | 分批处理内容 |

## 消息类型

### Bubble Tea 消息

| 消息类型 | 说明 |
|---------|------|
| **tea.KeyMsg** | 键盘消息 |
| **tea.MouseMsg** | 鼠标消息 |
| **tea.WindowSizeMsg** | 窗口大小消息 |
| **tea.QuitMsg** | 退出消息 |

### 组件内部消息

| 消息类型 | 说明 |
|---------|------|
| **spinner.TickMsg** | Spinner 定时器消息 |
| **timer.TickMsg** | Timer 定时器消息 |
| **timer.TimeoutMsg** | Timer 超时消息 |
| **stopwatch.TickMsg** | Stopwatch 定时器消息 |
| **textinput.PasteMsg** | 粘贴成功消息 |
| **textinput.PasteErrMsg** | 粘贴失败消息 |

## 使用示例

### 示例 1: 完整的文本输入组件

```go
package main

import (
    tea "github.com/purpose168/bubbletea-cn"
    "github.com/purpose168/bubbles-cn/textinput"
)

type model struct {
    textInput textinput.Model
    err       error
}

func initialModel() model {
    ti := textinput.New()
    ti.Placeholder = "请输入内容"
    ti.Focus()
    ti.CharLimit = 156
    ti.Width = 20

    return model{
        textInput: ti,
        err:       nil,
    }
}

func (m model) Init() tea.Cmd {
    return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEnter:
            return m, tea.Quit
        case tea.KeyCtrlC, tea.KeyEsc:
            return m, tea.Quit
        }
    case errMsg:
        m.err = msg
        return m, nil
    }

    m.textInput, cmd = m.textInput.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return "\n" + m.textInput.View() + "\n"
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

### 示例 2: 列表组件

```go
package main

import (
    tea "github.com/purpose168/bubbletea-cn"
    "github.com/purpose168/bubbles-cn/list"
)

type model struct {
    list list.Model
}

func initialModel() model {
    items := []list.Item{
        list.NewDefaultItem("选项1"),
        list.NewDefaultItem("选项2"),
        list.NewDefaultItem("选项3"),
    }

    l := list.New(items, list.WithShowHelp(true))
    l.Title = "选择选项"

    return model{list: l}
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEnter:
            item := m.list.SelectedItem()
            return m, tea.Printf("选择了: %v", item)
        case tea.KeyCtrlC:
            return m, tea.Quit
        }
    }

    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.list.View()
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

## 总结

Bubbles-CN 提供了一套完整的终端 UI 组件 API，每个组件都有清晰的接口定义和丰富的配置选项。开发者可以根据需要选择合适的组件，并通过组合使用构建复杂的终端用户界面。

### API 设计原则

1. **一致性**: 所有组件遵循相同的接口模式
2. **可组合**: 组件之间可以自由组合
3. **可扩展**: 提供丰富的配置选项
4. **易用性**: 提供默认配置和简单 API
5. **类型安全**: 利用 Go 的类型系统确保安全
