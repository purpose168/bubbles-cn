// Package list 提供了一个功能丰富的 Bubble Tea 组件，用于浏览通用项目列表。
// 它具有可选的过滤、分页、帮助、状态消息和用于指示活动的 spinner 等功能。
package list

import (
	"cmp"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	tea "github.com/purpose168/bubbletea-cn"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	lipgloss "github.com/purpose168/lipgloss-cn"
	"github.com/sahilm/fuzzy"

	"github.com/purpose168/bubbles-cn/help"
	"github.com/purpose168/bubbles-cn/key"
	"github.com/purpose168/bubbles-cn/paginator"
	"github.com/purpose168/bubbles-cn/spinner"
	"github.com/purpose168/bubbles-cn/textinput"
)

func clamp[T cmp.Ordered](v, low, high T) T {
	if low > high {
		low, high = high, low
	}
	return min(high, max(low, v))
}

// Item 是列表中显示的项目。
type Item interface {
	// FilterValue 是我们在过滤列表时用于与此项目进行过滤的值。
	FilterValue() string
}

// ItemDelegate 封装了所有列表项的通用功能。将此逻辑与项目本身分离的好处是，
// 您可以更改项目的功能而无需更改实际项目本身。
//
// 注意，如果委托还实现了 help.KeyMap 接口，与委托相关的帮助项将被添加到帮助视图中。
type ItemDelegate interface {
	// Render 渲染项目的视图。
	Render(w io.Writer, m Model, index int, item Item)

	// Height 是列表项的高度。
	Height() int

	// Spacing 是列表项之间的水平间隙大小（以单元格为单位）。
	Spacing() int

	// Update 是项目的更新循环。列表更新循环中的所有消息都将通过这里，
	// 除非用户正在设置过滤器。使用此方法执行适合此委托的项目级更新。
	Update(msg tea.Msg, m *Model) tea.Cmd
}

type filteredItem struct {
	index   int   // 未过滤列表中的索引
	item    Item  // 匹配的项目
	matches []int // 匹配项目的符文索引
}

type filteredItems []filteredItem

func (f filteredItems) items() []Item {
	agg := make([]Item, len(f))
	for i, v := range f {
		agg[i] = v.item
	}
	return agg
}

// FilterMatchesMsg 包含过滤期间匹配项的数据。该消息应路由到 Update 进行处理。
type FilterMatchesMsg []filteredItem

// FilterFunc 接受一个术语和一个要搜索的字符串列表（由 Item#FilterValue 定义）。
// 它应返回一个排序后的排名列表。
type FilterFunc func(string, []string) []Rank

// Rank 定义给定项目的排名。
type Rank struct {
	// 项目在原始输入中的索引。
	Index int
	// 与过滤术语匹配的实际单词的索引。
	MatchedIndexes []int
}

// DefaultFilter 使用 sahilm/fuzzy 来过滤列表。这是默认设置。
func DefaultFilter(term string, targets []string) []Rank {
	ranks := fuzzy.Find(term, targets)
	sort.Stable(ranks)
	result := make([]Rank, len(ranks))
	for i, r := range ranks {
		result[i] = Rank{
			Index:          r.Index,
			MatchedIndexes: r.MatchedIndexes,
		}
	}
	return result
}

// UnsortedFilter 使用 sahilm/fuzzy 来过滤列表。它不对结果进行排序。
func UnsortedFilter(term string, targets []string) []Rank {
	ranks := fuzzy.FindNoSort(term, targets)
	result := make([]Rank, len(ranks))
	for i, r := range ranks {
		result[i] = Rank{
			Index:          r.Index,
			MatchedIndexes: r.MatchedIndexes,
		}
	}
	return result
}

type statusMessageTimeoutMsg struct{}

// FilterState 描述模型上的当前过滤状态。
type FilterState int

// 可能的过滤状态。
const (
	Unfiltered    FilterState = iota // 未设置过滤器
	Filtering                        // 用户正在积极设置过滤器
	FilterApplied                    // 应用了过滤器且用户未编辑过滤器
)

// String 返回当前过滤状态的人类可读字符串。
func (f FilterState) String() string {
	return [...]string{
		"unfiltered",
		"filtering",
		"filter applied",
	}[f]
}

// Model 包含此组件的状态。
type Model struct {
	showTitle        bool
	showFilter       bool
	showStatusBar    bool
	showPagination   bool
	showHelp         bool
	filteringEnabled bool

	itemNameSingular string
	itemNamePlural   string

	Title             string
	Styles            Styles
	InfiniteScrolling bool

	// 用于导航列表的按键映射。
	KeyMap KeyMap

	// Filter 用于过滤列表。
	Filter FilterFunc

	disableQuitKeybindings bool

	// 简短和完整帮助视图的附加按键映射。这允许您在不重新实现帮助组件的情况下
	// 向帮助菜单添加附加按键映射。当然，如果您需要更多灵活性，
	// 也可以禁用列表的帮助组件并实现一个新的。
	AdditionalShortHelpKeys func() []key.Binding
	AdditionalFullHelpKeys  func() []key.Binding

	spinner     spinner.Model
	showSpinner bool
	width       int
	height      int
	Paginator   paginator.Model
	cursor      int
	Help        help.Model
	FilterInput textinput.Model
	filterState FilterState

	// 状态消息应保持可见的时间。默认情况下为 1 秒。
	StatusMessageLifetime time.Duration

	statusMessage      string
	statusMessageTimer *time.Timer

	// 我们正在处理的项目主集。
	items []Item

	// 我们当前显示的过滤项目。过滤、切换等将更改此切片，
	// 以便我们可以显示相关内容。因此，此字段应被视为临时的。
	filteredItems filteredItems

	delegate ItemDelegate
}

// New 返回一个具有合理默认值的新模型。
func New(items []Item, delegate ItemDelegate, width, height int) Model {
	styles := DefaultStyles()

	// 创建一个新的 spinner 模型
	sp := spinner.New()
	sp.Spinner = spinner.Line
	sp.Style = styles.Spinner

	// 创建一个新的文本输入模型用于过滤
	filterInput := textinput.New()
	filterInput.Prompt = "Filter: "
	filterInput.PromptStyle = styles.FilterPrompt
	filterInput.Cursor.Style = styles.FilterCursor
	filterInput.CharLimit = 64
	filterInput.Focus()

	// 创建一个新的分页器模型
	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = styles.ActivePaginationDot.String()
	p.InactiveDot = styles.InactivePaginationDot.String()

	// 创建并返回新的列表模型
	m := Model{
		showTitle:             true,
		showFilter:            true,
		showStatusBar:         true,
		showPagination:        true,
		showHelp:              true,
		itemNameSingular:      "item",
		itemNamePlural:        "items",
		filteringEnabled:      true,
		KeyMap:                DefaultKeyMap(),
		Filter:                DefaultFilter,
		Styles:                styles,
		Title:                 "List",
		FilterInput:           filterInput,
		StatusMessageLifetime: time.Second,

		width:     width,
		height:    height,
		delegate:  delegate,
		items:     items,
		Paginator: p,
		spinner:   sp,
		Help:      help.New(),
	}

	// 更新分页和按键绑定
	m.updatePagination()
	m.updateKeybindings()
	return m
}

// NewModel returns a new model with sensible defaults.
//
// Deprecated: use [New] instead.
var NewModel = New

// SetFilteringEnabled 启用或禁用过滤。注意这与 ShowFilter 不同，
// ShowFilter 仅仅隐藏或显示输入视图。
func (m *Model) SetFilteringEnabled(v bool) {
	m.filteringEnabled = v
	if !v {
		m.resetFiltering()
	}
	m.updateKeybindings()
}

// FilteringEnabled 返回是否启用了过滤。
func (m Model) FilteringEnabled() bool {
	return m.filteringEnabled
}

// SetShowTitle 显示或隐藏标题栏。
func (m *Model) SetShowTitle(v bool) {
	m.showTitle = v
	m.updatePagination()
}

// SetFilterText 显式设置过滤文本而不依赖用户输入。
// 它还将 filterState 设置为合理的默认值 FilterApplied，但这
// 可以通过 SetFilterState 更改。
func (m *Model) SetFilterText(filter string) {
	m.filterState = Filtering
	m.FilterInput.SetValue(filter)
	cmd := filterItems(*m)
	msg := cmd()
	fmm, _ := msg.(FilterMatchesMsg)
	m.filteredItems = filteredItems(fmm)
	m.filterState = FilterApplied
	m.GoToStart()
	m.FilterInput.CursorEnd()
	m.updatePagination()
	m.updateKeybindings()
}

// SetFilterState 允许手动设置过滤状态。
func (m *Model) SetFilterState(state FilterState) {
	m.GoToStart()
	m.filterState = state
	m.FilterInput.CursorEnd()
	m.FilterInput.Focus()
	m.updateKeybindings()
}

// ShowTitle 返回标题栏是否设置为渲染。
func (m Model) ShowTitle() bool {
	return m.showTitle
}

// SetShowFilter 显示或隐藏过滤栏。注意这不会禁用过滤，
// 它只是隐藏内置的过滤视图。这允许您使用 FilterInput 以不同的方式
// 渲染过滤 UI，而无需从头重新实现过滤。
//
// 要完全禁用过滤，请使用 EnableFiltering。
func (m *Model) SetShowFilter(v bool) {
	m.showFilter = v
	m.updatePagination()
}

// ShowFilter 返回过滤器是否设置为渲染。注意这与 FilteringEnabled 是分开的，
// 因此过滤可以被隐藏但仍然可以调用。这允许您以不同的方式渲染过滤，
// 而无需从头重新实现它。
func (m Model) ShowFilter() bool {
	return m.showFilter
}

// SetShowStatusBar 显示或隐藏显示列表元数据的视图，例如项目计数。
func (m *Model) SetShowStatusBar(v bool) {
	m.showStatusBar = v
	m.updatePagination()
}

// ShowStatusBar 返回状态栏是否设置为渲染。
func (m Model) ShowStatusBar() bool {
	return m.showStatusBar
}

// SetStatusBarItemName 定义项目标识符的替换。默认为 item/items。
func (m *Model) SetStatusBarItemName(singular, plural string) {
	m.itemNameSingular = singular
	m.itemNamePlural = plural
}

// StatusBarItemName 返回单数和复数状态栏项目名称。
func (m Model) StatusBarItemName() (string, string) {
	return m.itemNameSingular, m.itemNamePlural
}

// SetShowPagination 隐藏或显示分页器。注意分页仍然会活动，
// 只是不会显示。
func (m *Model) SetShowPagination(v bool) {
	m.showPagination = v
	m.updatePagination()
}

// ShowPagination 返回分页是否可见。
func (m *Model) ShowPagination() bool {
	return m.showPagination
}

// SetShowHelp 显示或隐藏帮助视图。
func (m *Model) SetShowHelp(v bool) {
	m.showHelp = v
	m.updatePagination()
}

// ShowHelp 返回帮助是否设置为渲染。
func (m Model) ShowHelp() bool {
	return m.showHelp
}

// Items 返回列表中的项目。
func (m Model) Items() []Item {
	return m.items
}

// SetItems 设置列表中可用的项目。这返回一个命令。
func (m *Model) SetItems(i []Item) tea.Cmd {
	var cmd tea.Cmd
	m.items = i

	// 如果当前处于过滤状态，则重新过滤项目
	if m.filterState != Unfiltered {
		m.filteredItems = nil
		cmd = filterItems(*m)
	}

	m.updatePagination()
	m.updateKeybindings()
	return cmd
}

// Select 选择列表的给定索引并转到其相应的页面。
func (m *Model) Select(index int) {
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage
}

// ResetSelected 将选定的项目重置为列表第一页的第一个项目。
func (m *Model) ResetSelected() {
	m.Select(0)
}

// ResetFilter 重置当前过滤状态。
func (m *Model) ResetFilter() {
	m.resetFiltering()
}

// SetItem 替换给定索引处的项目。这返回一个命令。
func (m *Model) SetItem(index int, item Item) tea.Cmd {
	var cmd tea.Cmd
	m.items[index] = item

	// 如果当前处于过滤状态，则重新过滤项目
	if m.filterState != Unfiltered {
		cmd = filterItems(*m)
	}

	m.updatePagination()
	return cmd
}

// InsertItem 在给定索引处插入一个项目。如果索引超出上界，
// 项目将被追加。这返回一个命令。
func (m *Model) InsertItem(index int, item Item) tea.Cmd {
	var cmd tea.Cmd
	m.items = insertItemIntoSlice(m.items, item, index)

	// 如果当前处于过滤状态，则重新过滤项目
	if m.filterState != Unfiltered {
		cmd = filterItems(*m)
	}

	m.updatePagination()
	m.updateKeybindings()
	return cmd
}

// RemoveItem 移除给定索引处的项目。如果索引超出范围，
// 这将是空操作。O(n) 复杂度，在 TUI 的情况下可能不会成为问题。
func (m *Model) RemoveItem(index int) {
	m.items = removeItemFromSlice(m.items, index)
	// 如果当前处于过滤状态，则从过滤结果中移除该项目
	if m.filterState != Unfiltered {
		m.filteredItems = removeFilterMatchFromSlice(m.filteredItems, index)
		if len(m.filteredItems) == 0 {
			m.resetFiltering()
		}
	}
	m.updatePagination()
}

// SetDelegate 设置项目委托。
func (m *Model) SetDelegate(d ItemDelegate) {
	m.delegate = d
	m.updatePagination()
}

// VisibleItems 返回可显示的总项目数。
func (m Model) VisibleItems() []Item {
	if m.filterState != Unfiltered {
		return m.filteredItems.items()
	}
	return m.items
}

// SelectedItem 返回列表中当前选定的项目。
func (m Model) SelectedItem() Item {
	i := m.Index()

	items := m.VisibleItems()
	if i < 0 || len(items) == 0 || len(items) <= i {
		return nil
	}

	return items[i]
}

// MatchesForItem 返回由当前过滤器匹配的符文位置（如果有）。
// 使用此方法来设置由活动过滤器匹配的符文的样式。
//
// 请参阅 DefaultItemView 以获取使用示例。
func (m Model) MatchesForItem(index int) []int {
	if m.filteredItems == nil || index >= len(m.filteredItems) {
		return nil
	}
	return m.filteredItems[index].matches
}

// Index 返回当前选定项目的索引，因为它存储在
// 过滤的项目列表中。
// 将此值与 SetItem() 一起使用可能不正确，请考虑使用
// GlobalIndex() 代替。
func (m Model) Index() int {
	return m.Paginator.Page*m.Paginator.PerPage + m.cursor
}

// GlobalIndex 返回当前选定项目的索引，因为它存储在
// 未过滤的项目列表中。此值可以与 SetItem() 一起使用。
func (m Model) GlobalIndex() int {
	index := m.Index()

	if m.filteredItems == nil || index >= len(m.filteredItems) {
		return index
	}

	return m.filteredItems[index].index
}

// Cursor 返回当前页面上光标的索引。
func (m Model) Cursor() int {
	return m.cursor
}

// CursorUp 向上移动光标。这也可以将状态移动到上一页。
func (m *Model) CursorUp() {
	m.cursor--

	// 如果我们在开始处，停止
	if m.cursor < 0 && m.Paginator.OnFirstPage() {
		// 如果启用了无限滚动，则转到最后一个项目
		if m.InfiniteScrolling {
			m.GoToEnd()
			return
		}
		m.cursor = 0
		return
	}

	// 正常移动光标
	if m.cursor >= 0 {
		return
	}

	// 转到上一页
	m.Paginator.PrevPage()
	m.cursor = m.maxCursorIndex()
}

// CursorDown 向下移动光标。这也可以将状态推进到下一页。
func (m *Model) CursorDown() {
	maxCursorIndex := m.maxCursorIndex()

	m.cursor++

	// 我们仍在当前页面的范围内，所以无需执行任何操作。
	if m.cursor <= maxCursorIndex {
		return
	}

	// 转到下一页
	if !m.Paginator.OnLastPage() {
		m.Paginator.NextPage()
		m.cursor = 0
		return
	}

	m.cursor = max(0, maxCursorIndex)

	// 如果启用了无限滚动，则转到第一个项目。
	if m.InfiniteScrolling {
		m.GoToStart()
	}
}

// GoToStart 移动到第一页，以及第一页上的第一个项目。
func (m *Model) GoToStart() {
	m.Paginator.Page = 0
	m.cursor = 0
}

// GoToEnd 移动到最后一页，以及最后一页上的最后一个项目。
func (m *Model) GoToEnd() {
	m.Paginator.Page = max(0, m.Paginator.TotalPages-1)
	m.cursor = m.maxCursorIndex()
}

// PrevPage 移动到上一页（如果可用）。
func (m *Model) PrevPage() {
	m.Paginator.PrevPage()
	m.cursor = clamp(m.cursor, 0, m.maxCursorIndex())
}

// NextPage 移动到下一页（如果可用）。
func (m *Model) NextPage() {
	m.Paginator.NextPage()
	m.cursor = clamp(m.cursor, 0, m.maxCursorIndex())
}

func (m *Model) maxCursorIndex() int {
	return max(0, m.Paginator.ItemsOnPage(len(m.VisibleItems()))-1)
}

// FilterState 返回当前过滤状态。
func (m Model) FilterState() FilterState {
	return m.filterState
}

// FilterValue 返回过滤器的当前值。
func (m Model) FilterValue() string {
	return m.FilterInput.Value()
}

// SettingFilter 返回用户当前是否正在编辑过滤值。
// 这纯粹是以下内容的便捷方法：
//
//	m.FilterState() == Filtering
//
// 包含在这里是因为在实现此组件时这是一个常见的检查项。
func (m Model) SettingFilter() bool {
	return m.filterState == Filtering
}

// IsFiltered 返回列表当前是否已过滤。
// 这纯粹是以下内容的便捷方法：
//
//	m.FilterState() == FilterApplied
func (m Model) IsFiltered() bool {
	return m.filterState == FilterApplied
}

// Width 返回当前宽度设置。
func (m Model) Width() int {
	return m.width
}

// Height 返回当前高度设置。
func (m Model) Height() int {
	return m.height
}

// SetSpinner 允许设置 spinner 样式。
func (m *Model) SetSpinner(spinner spinner.Spinner) {
	m.spinner.Spinner = spinner
}

// ToggleSpinner 切换 spinner。注意这也返回一个命令。
func (m *Model) ToggleSpinner() tea.Cmd {
	if !m.showSpinner {
		return m.StartSpinner()
	}
	m.StopSpinner()
	return nil
}

// StartSpinner 启动 spinner。注意这也返回一个命令。
func (m *Model) StartSpinner() tea.Cmd {
	m.showSpinner = true
	return m.spinner.Tick
}

// StopSpinner 停止 spinner。
func (m *Model) StopSpinner() {
	m.showSpinner = false
}

// DisableQuitKeybindings 是一个辅助函数，用于禁用用于退出的按键绑定，
// 以防您想在应用程序的其他地方处理此操作。
func (m *Model) DisableQuitKeybindings() {
	m.disableQuitKeybindings = true
	m.KeyMap.Quit.SetEnabled(false)
	m.KeyMap.ForceQuit.SetEnabled(false)
}

// NewStatusMessage 设置一个新的状态消息，该消息将显示有限的时间。
// 注意这也返回一个命令。
func (m *Model) NewStatusMessage(s string) tea.Cmd {
	m.statusMessage = s
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}

	m.statusMessageTimer = time.NewTimer(m.StatusMessageLifetime)

	// 等待超时
	return func() tea.Msg {
		<-m.statusMessageTimer.C
		return statusMessageTimeoutMsg{}
	}
}

// SetWidth 设置此组件的宽度。
func (m *Model) SetWidth(v int) {
	m.SetSize(v, m.height)
}

// SetHeight 设置此组件的高度。
func (m *Model) SetHeight(v int) {
	m.SetSize(m.width, v)
}

// SetSize 设置此组件的宽度和高度。
func (m *Model) SetSize(width, height int) {
	promptWidth := lipgloss.Width(m.Styles.Title.Render(m.FilterInput.Prompt))

	m.width = width
	m.height = height
	m.Help.Width = width
	m.FilterInput.Width = width - promptWidth - lipgloss.Width(m.spinnerView())
	m.updatePagination()
	m.updateKeybindings()
}

func (m *Model) resetFiltering() {
	if m.filterState == Unfiltered {
		return
	}

	m.filterState = Unfiltered
	m.FilterInput.Reset()
	m.filteredItems = nil
	m.updatePagination()
	m.updateKeybindings()
}

func (m Model) itemsAsFilterItems() filteredItems {
	fi := make([]filteredItem, len(m.items))
	for i, item := range m.items {
		fi[i] = filteredItem{
			item: item,
		}
	}
	return fi
}

// 根据过滤状态设置按键绑定。
func (m *Model) updateKeybindings() {
	switch m.filterState { //nolint:exhaustive
	case Filtering:
		// 在过滤状态下禁用导航按键
		m.KeyMap.CursorUp.SetEnabled(false)
		m.KeyMap.CursorDown.SetEnabled(false)
		m.KeyMap.NextPage.SetEnabled(false)
		m.KeyMap.PrevPage.SetEnabled(false)
		m.KeyMap.GoToStart.SetEnabled(false)
		m.KeyMap.GoToEnd.SetEnabled(false)
		m.KeyMap.Filter.SetEnabled(false)
		m.KeyMap.ClearFilter.SetEnabled(false)
		m.KeyMap.CancelWhileFiltering.SetEnabled(true)
		m.KeyMap.AcceptWhileFiltering.SetEnabled(m.FilterInput.Value() != "")
		m.KeyMap.Quit.SetEnabled(false)
		m.KeyMap.ShowFullHelp.SetEnabled(false)
		m.KeyMap.CloseFullHelp.SetEnabled(false)

	default:
		// 默认状态下的按键绑定
		hasItems := len(m.items) != 0
		m.KeyMap.CursorUp.SetEnabled(hasItems)
		m.KeyMap.CursorDown.SetEnabled(hasItems)

		hasPages := m.Paginator.TotalPages > 1
		m.KeyMap.NextPage.SetEnabled(hasPages)
		m.KeyMap.PrevPage.SetEnabled(hasPages)

		m.KeyMap.GoToStart.SetEnabled(hasItems)
		m.KeyMap.GoToEnd.SetEnabled(hasItems)

		m.KeyMap.Filter.SetEnabled(m.filteringEnabled && hasItems)
		m.KeyMap.ClearFilter.SetEnabled(m.filterState == FilterApplied)
		m.KeyMap.CancelWhileFiltering.SetEnabled(false)
		m.KeyMap.AcceptWhileFiltering.SetEnabled(false)
		m.KeyMap.Quit.SetEnabled(!m.disableQuitKeybindings)

		if m.Help.ShowAll {
			m.KeyMap.ShowFullHelp.SetEnabled(true)
			m.KeyMap.CloseFullHelp.SetEnabled(true)
		} else {
			minHelp := countEnabledBindings(m.FullHelp()) > 1
			m.KeyMap.ShowFullHelp.SetEnabled(minHelp)
			m.KeyMap.CloseFullHelp.SetEnabled(minHelp)
		}
	}
}

// 根据当前状态的项目数量更新分页。
func (m *Model) updatePagination() {
	index := m.Index()
	availHeight := m.height

	// 减去标题栏的高度
	if m.showTitle || (m.showFilter && m.filteringEnabled) {
		availHeight -= lipgloss.Height(m.titleView())
	}
	// 减去状态栏的高度
	if m.showStatusBar {
		availHeight -= lipgloss.Height(m.statusView())
	}
	// 减去分页器的高度
	if m.showPagination {
		availHeight -= lipgloss.Height(m.paginationView())
	}
	// 减去帮助视图的高度
	if m.showHelp {
		availHeight -= lipgloss.Height(m.helpView())
	}

	// 计算每页可以显示的项目数量
	m.Paginator.PerPage = max(1, availHeight/(m.delegate.Height()+m.delegate.Spacing()))

	// 设置总页数
	if pages := len(m.VisibleItems()); pages < 1 {
		m.Paginator.SetTotalPages(1)
	} else {
		m.Paginator.SetTotalPages(pages)
	}

	// 恢复索引
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage

	// 确保页面保持在范围内
	if m.Paginator.Page >= m.Paginator.TotalPages-1 {
		m.Paginator.Page = max(0, m.Paginator.TotalPages-1)
	}
}

func (m *Model) hideStatusMessage() {
	m.statusMessage = ""
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}
}

// Update 是 Bubble Tea 更新循环。
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理强制退出按键
		if key.Matches(msg, m.KeyMap.ForceQuit) {
			return m, tea.Quit
		}

	case FilterMatchesMsg:
		// 处理过滤匹配消息
		m.filteredItems = filteredItems(msg)
		return m, nil

	case spinner.TickMsg:
		// 处理 spinner 滴答消息
		newSpinnerModel, cmd := m.spinner.Update(msg)
		m.spinner = newSpinnerModel
		if m.showSpinner {
			cmds = append(cmds, cmd)
		}

	case statusMessageTimeoutMsg:
		// 处理状态消息超时
		m.hideStatusMessage()
	}

	// 根据过滤状态处理消息
	if m.filterState == Filtering {
		cmds = append(cmds, m.handleFiltering(msg))
	} else {
		cmds = append(cmds, m.handleBrowsing(msg))
	}

	return m, tea.Batch(cmds...)
}

// 当用户浏览列表时的更新。
func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// 注意：我们在退出之前匹配清除过滤器，因为默认情况下，
		// 它们都映射到 escape。
		case key.Matches(msg, m.KeyMap.ClearFilter):
			m.resetFiltering()

		case key.Matches(msg, m.KeyMap.Quit):
			return tea.Quit

		case key.Matches(msg, m.KeyMap.CursorUp):
			m.CursorUp()

		case key.Matches(msg, m.KeyMap.CursorDown):
			m.CursorDown()

		case key.Matches(msg, m.KeyMap.PrevPage):
			m.Paginator.PrevPage()

		case key.Matches(msg, m.KeyMap.NextPage):
			m.Paginator.NextPage()

		case key.Matches(msg, m.KeyMap.GoToStart):
			m.GoToStart()

		case key.Matches(msg, m.KeyMap.GoToEnd):
			m.GoToEnd()

		case key.Matches(msg, m.KeyMap.Filter):
			m.hideStatusMessage()
			// 仅当过滤器为空时，才用所有项目填充过滤器。
			if m.FilterInput.Value() == "" {
				m.filteredItems = m.itemsAsFilterItems()
			}
			m.GoToStart()
			m.filterState = Filtering
			m.FilterInput.CursorEnd()
			m.FilterInput.Focus()
			m.updateKeybindings()
			return textinput.Blink

		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
			m.updatePagination()
		}
	}

	// 调用委托的更新方法
	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	// 确保光标在有效范围内
	m.cursor = clamp(m.cursor, 0, m.maxCursorIndex())

	return tea.Batch(cmds...)
}

// 当用户在过滤编辑界面中时的更新。
func (m *Model) handleFiltering(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// 处理按键
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.KeyMap.CancelWhileFiltering):
			m.resetFiltering()
			m.KeyMap.Filter.SetEnabled(true)
			m.KeyMap.ClearFilter.SetEnabled(false)

		case key.Matches(msg, m.KeyMap.AcceptWhileFiltering):
			m.hideStatusMessage()

			if len(m.items) == 0 {
				break
			}

			h := m.VisibleItems()

			// 如果我们过滤后什么都没有，则清除过滤器
			if len(h) == 0 {
				m.resetFiltering()
				break
			}

			m.FilterInput.Blur()
			m.filterState = FilterApplied
			m.updateKeybindings()

			if m.FilterInput.Value() == "" {
				m.resetFiltering()
			}
		}
	}

	// 更新过滤文本输入组件
	newFilterInputModel, inputCmd := m.FilterInput.Update(msg)
	filterChanged := m.FilterInput.Value() != newFilterInputModel.Value()
	m.FilterInput = newFilterInputModel
	cmds = append(cmds, inputCmd)

	// 如果过滤输入已更改，则请求更新的过滤
	if filterChanged {
		cmds = append(cmds, filterItems(*m))
		m.KeyMap.AcceptWhileFiltering.SetEnabled(m.FilterInput.Value() != "")
	}

	// 更新分页
	m.updatePagination()

	return tea.Batch(cmds...)
}

// ShortHelp 返回要在缩略帮助视图中显示的绑定。这是
// help.KeyMap 接口的一部分。
func (m Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.CursorUp,
		m.KeyMap.CursorDown,
	}

	filtering := m.filterState == Filtering

	// 如果委托实现了 help.KeyMap 接口，则将简短帮助项
	// 添加到光标移动键之后的简短帮助中。
	if !filtering {
		if b, ok := m.delegate.(help.KeyMap); ok {
			kb = append(kb, b.ShortHelp()...)
		}
	}

	kb = append(kb,
		m.KeyMap.Filter,
		m.KeyMap.ClearFilter,
		m.KeyMap.AcceptWhileFiltering,
		m.KeyMap.CancelWhileFiltering,
	)

	if !filtering && m.AdditionalShortHelpKeys != nil {
		kb = append(kb, m.AdditionalShortHelpKeys()...)
	}

	return append(kb,
		m.KeyMap.Quit,
		m.KeyMap.ShowFullHelp,
	)
}

// FullHelp 返回要显示完整帮助视图的绑定。这是
// help.KeyMap 接口的一部分。
func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.KeyMap.CursorUp,
		m.KeyMap.CursorDown,
		m.KeyMap.NextPage,
		m.KeyMap.PrevPage,
		m.KeyMap.GoToStart,
		m.KeyMap.GoToEnd,
	}}

	filtering := m.filterState == Filtering

	// 如果委托实现了 help.KeyMap 接口，则将完整帮助按键绑定
	// 添加到完整帮助的特殊部分。
	if !filtering {
		if b, ok := m.delegate.(help.KeyMap); ok {
			kb = append(kb, b.FullHelp()...)
		}
	}

	listLevelBindings := []key.Binding{
		m.KeyMap.Filter,
		m.KeyMap.ClearFilter,
		m.KeyMap.AcceptWhileFiltering,
		m.KeyMap.CancelWhileFiltering,
	}

	if !filtering && m.AdditionalFullHelpKeys != nil {
		listLevelBindings = append(listLevelBindings, m.AdditionalFullHelpKeys()...)
	}

	return append(kb,
		listLevelBindings,
		[]key.Binding{
			m.KeyMap.Quit,
			m.KeyMap.CloseFullHelp,
		})
}

// View 渲染组件。
func (m Model) View() string {
	var (
		sections    []string
		availHeight = m.height
	)

	// 渲染标题栏或过滤器
	if m.showTitle || (m.showFilter && m.filteringEnabled) {
		v := m.titleView()
		sections = append(sections, v)
		availHeight -= lipgloss.Height(v)
	}

	// 渲染状态栏
	if m.showStatusBar {
		v := m.statusView()
		sections = append(sections, v)
		availHeight -= lipgloss.Height(v)
	}

	var pagination string
	// 渲染分页器
	if m.showPagination {
		pagination = m.paginationView()
		availHeight -= lipgloss.Height(pagination)
	}

	var help string
	// 渲染帮助视图
	if m.showHelp {
		help = m.helpView()
		availHeight -= lipgloss.Height(help)
	}

	// 渲染主要内容
	content := lipgloss.NewStyle().Height(availHeight).Render(m.populatedView())
	sections = append(sections, content)

	// 添加分页器
	if m.showPagination {
		sections = append(sections, pagination)
	}

	// 添加帮助视图
	if m.showHelp {
		sections = append(sections, help)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) titleView() string {
	var (
		view          string
		titleBarStyle = m.Styles.TitleBar

		// 我们需要考虑 spinner 的大小，即使我们不渲染它，
		// 也要为它预留一些空间，以防我们稍后打开它。
		spinnerView    = m.spinnerView()
		spinnerWidth   = lipgloss.Width(spinnerView)
		spinnerLeftGap = " "
		spinnerOnLeft  = titleBarStyle.GetPaddingLeft() >= spinnerWidth+lipgloss.Width(spinnerLeftGap) && m.showSpinner
	)

	// 如果过滤器正在显示，则绘制它。否则绘制标题。
	if m.showFilter && m.filterState == Filtering {
		view += m.FilterInput.View()
	} else if m.showTitle {
		if m.showSpinner && spinnerOnLeft {
			view += spinnerView + spinnerLeftGap
			titleBarGap := titleBarStyle.GetPaddingLeft()
			titleBarStyle = titleBarStyle.PaddingLeft(titleBarGap - spinnerWidth - lipgloss.Width(spinnerLeftGap))
		}

		view += m.Styles.Title.Render(m.Title)

		// 状态消息
		if m.filterState != Filtering {
			view += "  " + m.statusMessage
			view = ansi.Truncate(view, m.width-spinnerWidth, ellipsis)
		}
	}

	// Spinner
	if m.showSpinner && !spinnerOnLeft {
		// 将 spinner 放在右侧
		availSpace := m.width - lipgloss.Width(m.Styles.TitleBar.Render(view))
		if availSpace > spinnerWidth {
			view += strings.Repeat(" ", availSpace-spinnerWidth)
			view += spinnerView
		}
	}

	if len(view) > 0 {
		return titleBarStyle.Render(view)
	}
	return view
}

func (m Model) statusView() string {
	var status string

	totalItems := len(m.items)
	visibleItems := len(m.VisibleItems())

	var itemName string
	if visibleItems != 1 {
		itemName = m.itemNamePlural
	} else {
		itemName = m.itemNameSingular
	}

	itemsDisplay := fmt.Sprintf("%d %s", visibleItems, itemName)

	if m.filterState == Filtering { //nolint:nestif
		// 过滤结果
		if visibleItems == 0 {
			status = m.Styles.StatusEmpty.Render("Nothing matched")
		} else {
			status = itemsDisplay
		}
	} else if len(m.items) == 0 {
		// 未过滤：没有项目。
		status = m.Styles.StatusEmpty.Render("No " + m.itemNamePlural)
	} else {
		// 正常状态
		filtered := m.FilterState() == FilterApplied

		if filtered {
			f := strings.TrimSpace(m.FilterInput.Value())
			f = ansi.Truncate(f, 10, "…") //nolint:mnd
			status += fmt.Sprintf("“%s” ", f)
		}

		status += itemsDisplay
	}

	numFiltered := totalItems - visibleItems
	if numFiltered > 0 {
		status += m.Styles.DividerDot.String()
		status += m.Styles.StatusBarFilterCount.Render(fmt.Sprintf("%d filtered", numFiltered))
	}

	return m.Styles.StatusBar.Render(status)
}

func (m Model) paginationView() string {
	if m.Paginator.TotalPages < 2 { //nolint:mnd
		return ""
	}

	s := m.Paginator.View()

	// 如果点分页比窗口宽度宽，
	// 则使用阿拉伯数字分页器。
	if ansi.StringWidth(s) > m.width {
		m.Paginator.Type = paginator.Arabic
		s = m.Styles.ArabicPagination.Render(m.Paginator.View())
	}

	style := m.Styles.PaginationStyle
	if m.delegate.Spacing() == 0 && style.GetMarginTop() == 0 {
		style = style.MarginTop(1)
	}

	return style.Render(s)
}

func (m Model) populatedView() string {
	items := m.VisibleItems()

	var b strings.Builder

	// 空状态
	if len(items) == 0 {
		if m.filterState == Filtering {
			return ""
		}
		return m.Styles.NoItems.Render("No " + m.itemNamePlural + ".")
	}

	if len(items) > 0 {
		start, end := m.Paginator.GetSliceBounds(len(items))
		docs := items[start:end]

		for i, item := range docs {
			m.delegate.Render(&b, m, i+start, item)
			if i != len(docs)-1 {
				fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
			}
		}
	}

	// 如果没有足够的项目来填充此页面（总是最后一页），
	// 那么我们需要添加一些换行符来填充本应有项目的空间。
	itemsOnPage := m.Paginator.ItemsOnPage(len(items))
	if itemsOnPage < m.Paginator.PerPage {
		n := (m.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())
		if len(items) == 0 {
			n -= m.delegate.Height() - 1
		}
		fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

func (m Model) helpView() string {
	return m.Styles.HelpStyle.Render(m.Help.View(m))
}

func (m Model) spinnerView() string {
	return m.spinner.View()
}

func filterItems(m Model) tea.Cmd {
	return func() tea.Msg {
		// 如果过滤器为空或未处于过滤状态，则返回所有项目
		if m.FilterInput.Value() == "" || m.filterState == Unfiltered {
			return FilterMatchesMsg(m.itemsAsFilterItems()) // return nothing
		}

		items := m.items
		targets := make([]string, len(items))

		// 获取所有项目的过滤值
		for i, t := range items {
			targets[i] = t.FilterValue()
		}

		// 使用过滤器过滤项目
		filterMatches := []filteredItem{}
		for _, r := range m.Filter(m.FilterInput.Value(), targets) {
			filterMatches = append(filterMatches, filteredItem{
				index:   r.Index,
				item:    items[r.Index],
				matches: r.MatchedIndexes,
			})
		}

		return FilterMatchesMsg(filterMatches)
	}
}

func insertItemIntoSlice(items []Item, item Item, index int) []Item {
	if items == nil {
		return []Item{item}
	}
	if index >= len(items) {
		return append(items, item)
	}

	index = max(0, index)

	items = append(items, nil)
	copy(items[index+1:], items[index:])
	items[index] = item
	return items
}

// 从给定索引处的项目切片中移除一个项目。这在 O(n) 中运行。
func removeItemFromSlice(i []Item, index int) []Item {
	if index >= len(i) {
		return i // noop
	}
	copy(i[index:], i[index+1:])
	i[len(i)-1] = nil
	return i[:len(i)-1]
}

func removeFilterMatchFromSlice(i []filteredItem, index int) []filteredItem {
	if index >= len(i) {
		return i // noop
	}
	copy(i[index:], i[index+1:])
	i[len(i)-1] = filteredItem{}
	return i[:len(i)-1]
}

func countEnabledBindings(groups [][]key.Binding) (agg int) {
	for _, group := range groups {
		for _, kb := range group {
			if kb.Enabled() {
				agg++
			}
		}
	}
	return agg
}
