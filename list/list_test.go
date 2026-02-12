package list

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	tea "github.com/purpose168/bubbletea-cn"
)

// item 是一个简单的字符串类型，实现了 Item 接口
type item string

// FilterValue 返回项目的过滤值
func (i item) FilterValue() string { return string(i) }

// itemDelegate 是一个简单的委托实现
type itemDelegate struct{}

// Height 返回委托的高度
func (d itemDelegate) Height() int { return 1 }

// Spacing 返回委托的间距
func (d itemDelegate) Spacing() int { return 0 }

// Update 处理委托的更新
func (d itemDelegate) Update(msg tea.Msg, m *Model) tea.Cmd { return nil }

// Render 渲染列表项
func (d itemDelegate) Render(w io.Writer, m Model, index int, listItem Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)
	fmt.Fprint(w, m.Styles.TitleBar.Render(str))
}

// TestStatusBarItemName 测试状态栏中的项目名称显示
func TestStatusBarItemName(t *testing.T) {
	// 创建一个包含两个项目的列表
	list := New([]Item{item("foo"), item("bar")}, itemDelegate{}, 10, 10)
	expected := "2 items"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}

	// 更新为包含一个项目的列表
	list.SetItems([]Item{item("foo")})
	expected = "1 item"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}
}

// TestStatusBarWithoutItems 测试没有项目时的状态栏显示
func TestStatusBarWithoutItems(t *testing.T) {
	// 创建一个空列表
	list := New([]Item{}, itemDelegate{}, 10, 10)

	expected := "No items"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}
}

// TestCustomStatusBarItemName 测试自定义状态栏项目名称
func TestCustomStatusBarItemName(t *testing.T) {
	// 创建一个包含两个项目的列表
	list := New([]Item{item("foo"), item("bar")}, itemDelegate{}, 10, 10)
	// 设置自定义项目名称
	list.SetStatusBarItemName("connection", "connections")

	expected := "2 connections"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}

	// 更新为包含一个项目的列表
	list.SetItems([]Item{item("foo")})
	expected = "1 connection"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}

	// 更新为空列表
	list.SetItems([]Item{})
	expected = "No connections"
	if !strings.Contains(list.statusView(), expected) {
		t.Fatalf("Error: expected view to contain %s", expected)
	}
}

// TestSetFilterText 测试设置过滤文本
func TestSetFilterText(t *testing.T) {
	tc := []Item{item("foo"), item("bar"), item("baz")}

	// 创建列表并设置过滤文本
	list := New(tc, itemDelegate{}, 10, 10)
	list.SetFilterText("ba")

	// 测试未过滤状态
	list.SetFilterState(Unfiltered)
	expected := tc
	// TODO: 当项目迁移到 go1.18 或更高版本时，替换为 slices.Equal()
	if !reflect.DeepEqual(list.VisibleItems(), expected) {
		t.Fatalf("Error: expected view to contain only %s", expected)
	}

	// 测试过滤中状态
	list.SetFilterState(Filtering)
	expected = []Item{item("bar"), item("baz")}
	if !reflect.DeepEqual(list.VisibleItems(), expected) {
		t.Fatalf("Error: expected view to contain only %s", expected)
	}

	// 测试已应用过滤状态
	list.SetFilterState(FilterApplied)
	if !reflect.DeepEqual(list.VisibleItems(), expected) {
		t.Fatalf("Error: expected view to contain only %s", expected)
	}
}

// TestSetFilterState 测试设置过滤状态
func TestSetFilterState(t *testing.T) {
	tc := []Item{item("foo"), item("bar"), item("baz")}

	// 创建列表并设置过滤文本
	list := New(tc, itemDelegate{}, 10, 10)
	list.SetFilterText("ba")

	// 测试未过滤状态
	list.SetFilterState(Unfiltered)
	expected, notExpected := "up", "clear filter"

	lines := strings.Split(list.View(), "\n")
	footer := lines[len(lines)-1]

	if !strings.Contains(footer, expected) || strings.Contains(footer, notExpected) {
		t.Fatalf("Error: expected view to contain '%s' not '%s'", expected, notExpected)
	}

	// 测试过滤中状态
	list.SetFilterState(Filtering)
	expected, notExpected = "filter", "more"

	lines = strings.Split(list.View(), "\n")
	footer = lines[len(lines)-1]

	if !strings.Contains(footer, expected) || strings.Contains(footer, notExpected) {
		t.Fatalf("Error: expected view to contain '%s' not '%s'", expected, notExpected)
	}

	// 测试已应用过滤状态
	list.SetFilterState(FilterApplied)
	expected = "clear"

	lines = strings.Split(list.View(), "\n")
	footer = lines[len(lines)-1]

	if !strings.Contains(footer, expected) {
		t.Fatalf("Error: expected view to contain '%s'", expected)
	}
}
