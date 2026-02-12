// Package filepicker 为 Bubble Tea 应用程序提供文件选择器组件。
package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/dustin/go-humanize"
	"github.com/purpose168/bubbles-cn/key"
	tea "github.com/purpose168/bubbletea-cn"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

var lastID int64

// nextID 生成下一个唯一 ID。
func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// New 返回一个带有默认样式和键绑定的新文件选择器模型。
func New() Model {
	return Model{
		id:               nextID(),        // 生成唯一 ID
		CurrentDirectory: ".",             // 当前目录默认为当前工作目录
		Cursor:           ">",             // 光标默认样式
		AllowedTypes:     []string{},      // 允许的文件类型，默认为空（允许所有文件）
		selected:         0,               // 当前选中的文件索引
		ShowPermissions:  true,            // 是否显示文件权限
		ShowSize:         true,            // 是否显示文件大小
		ShowHidden:       false,           // 是否显示隐藏文件
		DirAllowed:       false,           // 是否允许选择目录
		FileAllowed:      true,            // 是否允许选择文件
		AutoHeight:       true,            // 是否自动调整高度
		Height:           0,               // 高度，默认为 0
		max:              0,               // 可视区域最大索引
		min:              0,               // 可视区域最小索引
		selectedStack:    newStack(),      // 选中索引栈，用于返回上一级目录时恢复选中状态
		minStack:         newStack(),      // 最小索引栈
		maxStack:         newStack(),      // 最大索引栈
		KeyMap:           DefaultKeyMap(), // 默认键映射
		Styles:           DefaultStyles(), // 默认样式
	}
}

// errorMsg 表示错误消息。
type errorMsg struct {
	err error
}

// readDirMsg 表示读取目录消息。
type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

const (
	marginBottom  = 5 // 底部边距
	fileSizeWidth = 7 // 文件大小显示宽度
	paddingLeft   = 2 // 左侧内边距
)

// KeyMap 定义每个用户操作的键绑定。
type KeyMap struct {
	GoToTop  key.Binding // 跳转到顶部
	GoToLast key.Binding // 跳转到底部
	Down     key.Binding // 向下移动
	Up       key.Binding // 向上移动
	PageUp   key.Binding // 向上翻页
	PageDown key.Binding // 向下翻页
	Back     key.Binding // 返回上一级目录
	Open     key.Binding // 打开文件或目录
	Select   key.Binding // 选择文件
}

// DefaultKeyMap 定义默认键绑定。
func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),                            // g 键跳转到顶部
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),                             // G 键跳转到底部
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("j", "down")),           // j/下箭头/ctrl+n 向下移动
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("k", "up")),               // k/上箭头/ctrl+p 向上移动
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),               // K/PageUp 向上翻页
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),         // J/PageDown 向下翻页
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("h", "back")), // h/退格/左箭头/Esc 返回上一级
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),           // l/右箭头/Enter 打开
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),                   // Enter 选择
	}
}

// Styles 定义文件选择器中样式的可能自定义项。
type Styles struct {
	DisabledCursor   lipgloss.Style // 禁用状态的光标样式
	Cursor           lipgloss.Style // 光标样式
	Symlink          lipgloss.Style // 符号链接样式
	Directory        lipgloss.Style // 目录样式
	File             lipgloss.Style // 文件样式
	DisabledFile     lipgloss.Style // 禁用状态的文件样式
	Permission       lipgloss.Style // 权限样式
	Selected         lipgloss.Style // 选中项样式
	DisabledSelected lipgloss.Style // 禁用状态的选中项样式
	FileSize         lipgloss.Style // 文件大小样式
	EmptyDirectory   lipgloss.Style // 空目录样式
}

// DefaultStyles 定义文件选择器的默认样式。
func DefaultStyles() Styles {
	return DefaultStylesWithRenderer(lipgloss.DefaultRenderer())
}

// DefaultStylesWithRenderer 定义文件选择器的默认样式，
// 使用给定的 Lip Gloss 渲染器。
func DefaultStylesWithRenderer(r *lipgloss.Renderer) Styles {
	return Styles{
		DisabledCursor:   r.NewStyle().Foreground(lipgloss.Color("247")),                                                               // 禁用光标颜色
		Cursor:           r.NewStyle().Foreground(lipgloss.Color("212")),                                                               // 光标颜色
		Symlink:          r.NewStyle().Foreground(lipgloss.Color("36")),                                                                // 符号链接颜色
		Directory:        r.NewStyle().Foreground(lipgloss.Color("99")),                                                                // 目录颜色
		File:             r.NewStyle(),                                                                                                 // 文件默认样式
		DisabledFile:     r.NewStyle().Foreground(lipgloss.Color("243")),                                                               // 禁用文件颜色
		DisabledSelected: r.NewStyle().Foreground(lipgloss.Color("247")),                                                               // 禁用选中项颜色
		Permission:       r.NewStyle().Foreground(lipgloss.Color("244")),                                                               // 权限颜色
		Selected:         r.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),                                                    // 选中项颜色和样式
		FileSize:         r.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),                    // 文件大小样式
		EmptyDirectory:   r.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("Bummer. No Files Found."), // 空目录提示
	}
}

// Model 表示文件选择器模型。
type Model struct {
	id int // 唯一 ID

	// Path 是用户通过文件选择器选择的路径。
	Path string

	// CurrentDirectory 是用户当前所在的目录。
	CurrentDirectory string

	// AllowedTypes 指定用户可以选择的文件类型。
	// 如果为空，用户可以选择任何文件。
	AllowedTypes []string

	KeyMap          KeyMap        // 键绑定
	files           []os.DirEntry // 文件列表
	ShowPermissions bool          // 是否显示权限
	ShowSize        bool          // 是否显示大小
	ShowHidden      bool          // 是否显示隐藏文件
	DirAllowed      bool          // 是否允许选择目录
	FileAllowed     bool          // 是否允许选择文件

	FileSelected  string // 选中的文件
	selected      int    // 当前选中的索引
	selectedStack stack  // 选中索引栈

	min      int   // 可视区域最小索引
	max      int   // 可视区域最大索引
	maxStack stack // 最大索引栈
	minStack stack // 最小索引栈

	// Height 是选择器的高度。
	//
	// Deprecated: 使用 [Model.SetHeight] 代替。
	Height     int  // 高度
	AutoHeight bool // 是否自动调整高度

	Cursor string // 光标样式
	Styles Styles // 样式
}

// stack 表示栈结构，用于存储目录导航历史。
type stack struct {
	Push   func(int)  // 入栈
	Pop    func() int // 出栈
	Length func() int // 获取栈长度
}

// newStack 创建一个新的栈。
func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

// pushView 将当前视图状态压入栈中。
func (m *Model) pushView(selected, minimum, maximum int) {
	m.selectedStack.Push(selected)
	m.minStack.Push(minimum)
	m.maxStack.Push(maximum)
}

// popView 从栈中弹出视图状态。
func (m *Model) popView() (int, int, int) {
	return m.selectedStack.Pop(), m.minStack.Pop(), m.maxStack.Pop()
}

// readDir 读取目录内容并返回命令。
func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		// 排序目录项：目录在前，文件在后，然后按名称排序
		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		if showHidden {
			return readDirMsg{id: m.id, entries: dirEntries}
		}

		// 过滤隐藏文件
		var sanitizedDirEntries []os.DirEntry
		for _, dirEntry := range dirEntries {
			isHidden, _ := IsHidden(dirEntry.Name())
			if isHidden {
				continue
			}
			sanitizedDirEntries = append(sanitizedDirEntries, dirEntry)
		}
		return readDirMsg{id: m.id, entries: sanitizedDirEntries}
	}
}

// Init 初始化文件选择器模型。
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory, m.ShowHidden)
}

// SetHeight 设置文件选择器的高度。
func (m *Model) SetHeight(height int) {
	m.Height = height
	if m.max > m.Height-1 {
		m.max = m.min + m.Height - 1
	}
}

// Update 处理文件选择器模型中的用户交互。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.files = msg.entries
		m.max = max(m.max, m.Height-1)
	case tea.WindowSizeMsg:
		if m.AutoHeight {
			m.Height = msg.Height - marginBottom
		}
		m.max = m.Height - 1
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
		case key.Matches(msg, m.KeyMap.GoToLast):
			m.selected = len(m.files) - 1
			m.min = len(m.files) - m.Height
			m.max = len(m.files) - 1
		case key.Matches(msg, m.KeyMap.Down):
			m.selected++
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			if m.selected > m.max {
				m.min++
				m.max++
			}
		case key.Matches(msg, m.KeyMap.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			if m.selected < m.min {
				m.min--
				m.max--
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			m.selected += m.Height
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			m.min += m.Height
			m.max += m.Height

			if m.max >= len(m.files) {
				m.max = len(m.files) - 1
				m.min = m.max - m.Height
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			m.selected -= m.Height
			if m.selected < 0 {
				m.selected = 0
			}
			m.min -= m.Height
			m.max -= m.Height

			if m.min < 0 {
				m.min = 0
				m.max = m.min + m.Height
			}
		case key.Matches(msg, m.KeyMap.Back):
			m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
			if m.selectedStack.Length() > 0 {
				m.selected, m.min, m.max = m.popView()
			} else {
				m.selected = 0
				m.min = 0
				m.max = m.Height - 1
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		case key.Matches(msg, m.KeyMap.Open):
			if len(m.files) == 0 {
				break
			}

			f := m.files[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if key.Matches(msg, m.KeyMap.Select) {
					// 选择当前路径作为选择结果
					m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				}
			}

			if !isDir {
				break
			}

			m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
			m.pushView(m.selected, m.min, m.max)
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		}
	}
	return m, nil
}

// View 返回文件选择器的视图。
func (m Model) View() string {
	if len(m.files) == 0 {
		return m.Styles.EmptyDirectory.Height(m.Height).MaxHeight(m.Height).String()
	}
	var s strings.Builder

	for i, f := range m.files {
		if i < m.min || i > m.max {
			continue
		}

		var symlinkPath string
		info, _ := f.Info()
		isSymlink := info.Mode()&os.ModeSymlink != 0
		size := strings.Replace(humanize.Bytes(uint64(info.Size())), " ", "", 1) //nolint:gosec
		name := f.Name()

		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, name))
		}

		disabled := !m.canSelect(name) && !f.IsDir()

		if m.selected == i { //nolint:nestif
			selected := ""
			if m.ShowPermissions {
				selected += " " + info.Mode().String()
			}
			if m.ShowSize {
				selected += fmt.Sprintf("%"+strconv.Itoa(m.Styles.FileSize.GetWidth())+"s", size)
			}
			selected += " " + name
			if isSymlink {
				selected += " → " + symlinkPath
			}
			if disabled {
				s.WriteString(m.Styles.DisabledSelected.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selected))
			} else {
				s.WriteString(m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selected))
			}
			s.WriteRune('\n')
			continue
		}

		style := m.Styles.File
		if f.IsDir() {
			style = m.Styles.Directory
		} else if isSymlink {
			style = m.Styles.Symlink
		} else if disabled {
			style = m.Styles.DisabledFile
		}

		fileName := style.Render(name)
		s.WriteString(m.Styles.Cursor.Render(" "))
		if isSymlink {
			fileName += " → " + symlinkPath
		}
		if m.ShowPermissions {
			s.WriteString(" " + m.Styles.Permission.Render(info.Mode().String()))
		}
		if m.ShowSize {
			s.WriteString(m.Styles.FileSize.Render(size))
		}
		s.WriteString(" " + fileName)
		s.WriteRune('\n')
	}

	// 填充剩余空间
	for i := lipgloss.Height(s.String()); i <= m.Height; i++ {
		s.WriteRune('\n')
	}

	return s.String()
}

// DidSelectFile 返回用户是否选择了文件（在此消息上）。
func (m Model) DidSelectFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// DidSelectDisabledFile 返回用户是否尝试选择禁用的文件
// （在此消息上）。只有当你想警告用户他们尝试选择禁用的文件时，这才是必要的。
func (m Model) DidSelectDisabledFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// didSelectFile 检查用户是否选择了文件。
func (m Model) didSelectFile(msg tea.Msg) (bool, string) {
	if len(m.files) == 0 {
		return false, ""
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 如果消息与 Select 键映射不匹配，则这不可能是选择操作。
		if !key.Matches(msg, m.KeyMap.Select) {
			return false, ""
		}

		// 按键是选择操作，让我们确认当前文件是否可以
		// 被选择或用于导航到更深层次的堆栈。
		f := m.files[m.selected]
		info, err := f.Info()
		if err != nil {
			return false, ""
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
			info, err := os.Stat(symlinkPath)
			if err != nil {
				break
			}
			if info.IsDir() {
				isDir = true
			}
		}

		if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) && m.Path != "" {
			return true, m.Path
		}

		// 如果消息不是 KeyMsg，则文件不可能在此迭代中被选择。
		// 只有 KeyMsg 可以选择文件。
	default:
		return false, ""
	}
	return false, ""
}

// canSelect 检查是否可以选择给定的文件。
func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) <= 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}
