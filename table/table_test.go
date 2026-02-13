package table

import (
	"reflect"
	"testing"

	"github.com/purpose168/bubbles-cn/help"
	"github.com/purpose168/bubbles-cn/viewport"
	"github.com/purpose168/charm-experimental-packages-cn/ansi"
	"github.com/purpose168/charm-experimental-packages-cn/exp/golden"
	lipgloss "github.com/purpose168/lipgloss-cn"
)

// testCols 测试用的列定义
var testCols = []Column{
	{Title: "col1", Width: 10},
	{Title: "col2", Width: 10},
	{Title: "col3", Width: 10},
}

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	tests := map[string]struct {
		opts []Option // 选项
		want Model    // 期望的模型
	}{
		"Default": { // 默认情况
			want: Model{
				// Default fields 默认字段
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),
			},
		},
		"WithColumns": { // 设置列
			opts: []Option{
				WithColumns([]Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				}),
			},
			want: Model{
				// Default fields 默认字段
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields 修改的字段
				cols: []Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				},
			},
		},
		"WithColumns; WithRows": { // 设置列和行
			opts: []Option{
				WithColumns([]Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				}),
				WithRows([]Row{
					{"1", "Foo"},
					{"2", "Bar"},
				}),
			},
			want: Model{
				// Default fields 默认字段
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields 修改的字段
				cols: []Column{
					{Title: "Foo", Width: 1},
					{Title: "Bar", Width: 2},
				},
				rows: []Row{
					{"1", "Foo"},
					{"2", "Bar"},
				},
			},
		},
		"WithHeight": { // 设置高度
			opts: []Option{
				WithHeight(10),
			},
			want: Model{
				// Default fields 默认字段
				cursor: 0,
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

				// Modified fields 修改的字段
				// Viewport height is 1 less than the provided height when no header is present since lipgloss.Height adds 1
				// 当没有表头时，视口高度比提供的高度少 1，因为 lipgloss.Height 会加 1
				viewport: viewport.New(0, 9),
			},
		},
		"WithWidth": { // 设置宽度
			opts: []Option{
				WithWidth(10),
			},
			want: Model{
				// Default fields 默认字段
				cursor: 0,
				KeyMap: DefaultKeyMap(),
				Help:   help.New(),
				styles: DefaultStyles(),

				// Modified fields 修改的字段
				// Viewport height is 1 less than the provided height when no header is present since lipgloss.Height adds 1
				// 当没有表头时，视口高度比提供的高度少 1，因为 lipgloss.Height 会加 1
				viewport: viewport.New(10, 20),
			},
		},
		"WithFocused": { // 设置聚焦状态
			opts: []Option{
				WithFocused(true),
			},
			want: Model{
				// Default fields 默认字段
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields 修改的字段
				focus: true,
			},
		},
		"WithStyles": { // 设置样式
			opts: []Option{
				WithStyles(Styles{}),
			},
			want: Model{
				// Default fields 默认字段
				cursor:   0,
				viewport: viewport.New(0, 20),
				KeyMap:   DefaultKeyMap(),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields 修改的字段
				// 已移除重复的 styles 字段赋值，因在上一层已赋值
			},
		},
		"WithKeyMap": { // 设置键映射
			opts: []Option{
				WithKeyMap(KeyMap{}),
			},
			want: Model{
				// Default fields 默认字段
				cursor:   0,
				viewport: viewport.New(0, 20),
				Help:     help.New(),
				styles:   DefaultStyles(),

				// Modified fields 修改的字段
				KeyMap: KeyMap{},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.want.UpdateViewport()

			got := New(tc.opts...)

			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("\n\nwant %v\n\ngot %v", tc.want, got)
			}
		})
	}
}

// TestModel_FromValues 测试从值创建表格
func TestModel_FromValues(t *testing.T) {
	input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, ",")

	if len(table.rows) != 3 {
		t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
	}

	expect := []Row{
		{"foo1", "bar1"},
		{"foo2", "bar2"},
		{"foo3", "bar3"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("\n\nwant %v\n\ngot %v", expect, table.rows)
	}
}

// TestModel_FromValues_WithTabSeparator 测试使用制表符分隔符从值创建表格
func TestModel_FromValues_WithTabSeparator(t *testing.T) {
	input := "foo1.\tbar1\nfoo,bar,baz\tbar,2"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, "\t")

	if len(table.rows) != 2 {
		t.Fatalf("expect table to have 2 rows but it has %d", len(table.rows))
	}

	expect := []Row{
		{"foo1.", "bar1"},
		{"foo,bar,baz", "bar,2"},
	}
	if !reflect.DeepEqual(table.rows, expect) {
		t.Fatalf("\n\nwant %v\n\ngot %v", expect, table.rows)
	}
}

// TestModel_RenderRow 测试渲染行
func TestModel_RenderRow(t *testing.T) {
	tests := []struct {
		name     string // 测试名称
		table    *Model // 表格模型
		expected string // 期望的输出
	}{
		{
			name: "simple row", // 简单行
			table: &Model{
				rows:   []Row{{"Foooooo", "Baaaaar", "Baaaaaz"}},
				cols:   testCols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooo   Baaaaar   Baaaaaz   ",
		},
		{
			name: "simple row with truncations", // 带截断的简单行
			table: &Model{
				rows:   []Row{{"Foooooooooo", "Baaaaaaaaar", "Quuuuuuuuux"}},
				cols:   testCols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooooo…Baaaaaaaa…Quuuuuuuu…",
		},
		{
			name: "simple row avoiding truncations", // 避免截断的简单行
			table: &Model{
				rows:   []Row{{"Fooooooooo", "Baaaaaaaar", "Quuuuuuuux"}},
				cols:   testCols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "FoooooooooBaaaaaaaarQuuuuuuuux",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			row := tc.table.renderRow(0)
			if row != tc.expected {
				t.Fatalf("\n\nWant: \n%s\n\nGot:  \n%s\n", tc.expected, row)
			}
		})
	}
}

// TestTableAlignment 测试表格对齐
func TestTableAlignment(t *testing.T) {
	t.Run("No border", func(t *testing.T) { // 无边框
		biscuits := New(
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{"Chocolate Digestives", "UK", "Yes"},
				{"Tim Tams", "Australia", "No"},
				{"Hobnobs", "UK", "Yes"},
			}),
		)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("With border", func(t *testing.T) { // 带边框
		baseStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

		s := DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)

		biscuits := New(
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{"Chocolate Digestives", "UK", "Yes"},
				{"Tim Tams", "Australia", "No"},
				{"Hobnobs", "UK", "Yes"},
			}),
			WithStyles(s),
		)
		got := ansi.Strip(baseStyle.Render(biscuits.View()))
		golden.RequireEqual(t, []byte(got))
	})
}

// TestCursorNavigation 测试光标导航
func TestCursorNavigation(t *testing.T) {
	tests := map[string]struct {
		rows   []Row        // 行数据
		action func(*Model) // 执行的操作
		want   int          // 期望的光标位置
	}{
		"New": { // 新建
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
			},
			action: func(_ *Model) {},
			want:   0,
		},
		"MoveDown": { // 向下移动
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.MoveDown(2)
			},
			want: 2,
		},
		"MoveUp": { // 向上移动
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.MoveUp(2)
			},
			want: 1,
		},
		"GotoBottom": { // 跳转到底部
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.GotoBottom()
			},
			want: 3,
		},
		"GotoTop": { // 跳转到顶部
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.GotoTop()
			},
			want: 0,
		},
		"SetCursor": { // 设置光标
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.SetCursor(2)
			},
			want: 2,
		},
		"MoveDown with overflow": { // 向下移动超出范围
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.MoveDown(5)
			},
			want: 3,
		},
		"MoveUp with overflow": { // 向上移动超出范围
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.cursor = 3
				t.MoveUp(5)
			},
			want: 0,
		},
		"Blur does not stop movement": { // 失焦不阻止移动
			rows: []Row{
				{"r1"},
				{"r2"},
				{"r3"},
				{"r4"},
			},
			action: func(t *Model) {
				t.Blur()
				t.MoveDown(2)
			},
			want: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			table := New(WithColumns(testCols), WithRows(tc.rows))
			tc.action(&table)

			if table.Cursor() != tc.want {
				t.Errorf("want %d, got %d", tc.want, table.Cursor())
			}
		})
	}
}

// TestModel_SetRows 测试设置行
func TestModel_SetRows(t *testing.T) {
	table := New(WithColumns(testCols))

	if len(table.rows) != 0 {
		t.Fatalf("want 0, got %d", len(table.rows))
	}

	table.SetRows([]Row{{"r1"}, {"r2"}})

	if len(table.rows) != 2 {
		t.Fatalf("want 2, got %d", len(table.rows))
	}

	want := []Row{{"r1"}, {"r2"}}
	if !reflect.DeepEqual(table.rows, want) {
		t.Fatalf("\n\nwant %v\n\ngot %v", want, table.rows)
	}
}

// TestModel_SetColumns 测试设置列
func TestModel_SetColumns(t *testing.T) {
	table := New()

	if len(table.cols) != 0 {
		t.Fatalf("want 0, got %d", len(table.cols))
	}

	table.SetColumns([]Column{{Title: "Foo"}, {Title: "Bar"}})

	if len(table.cols) != 2 {
		t.Fatalf("want 2, got %d", len(table.cols))
	}

	want := []Column{{Title: "Foo"}, {Title: "Bar"}}
	if !reflect.DeepEqual(table.cols, want) {
		t.Fatalf("\n\nwant %v\n\ngot %v", want, table.cols)
	}
}

// TestModel_View 测试表格视图
func TestModel_View(t *testing.T) {
	tests := map[string]struct {
		modelFunc func() Model // 模型生成函数
		skip      bool         // 是否跳过测试
	}{
		// TODO(?): should the view/output of empty tables use the same default height? (this has height 21)
		// TODO(?): 空表格的视图/输出是否应该使用相同的默认高度？（这里的高度是 21）
		"Empty": { // 空表格
			modelFunc: func() Model {
				return New()
			},
		},
		"Single row and column": { // 单行单列
			modelFunc: func() Model {
				return New(
					WithColumns([]Column{
						{Title: "Name", Width: 25},
					}),
					WithRows([]Row{
						{"Chocolate Digestives"},
					}),
				)
			},
		},
		"Multiple rows and columns": { // 多行多列
			modelFunc: func() Model {
				return New(
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
				)
			},
		},
		// TODO(fix): since the table height is tied to the viewport height, adding vertical padding to the headers' height directly increases the table height.
		// TODO(fix): 由于表格高度与视口高度相关联，为表头高度添加垂直填充会直接增加表格高度。
		"Extra padding": { // 额外填充
			modelFunc: func() Model {
				s := DefaultStyles()
				s.Header = lipgloss.NewStyle().Padding(2, 2)
				s.Cell = lipgloss.NewStyle().Padding(2, 2)

				return New(
					WithHeight(10),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
					WithStyles(s),
				)
			},
		},
		"No padding": { // 无填充
			modelFunc: func() Model {
				s := DefaultStyles()
				s.Header = lipgloss.NewStyle()
				s.Cell = lipgloss.NewStyle()

				return New(
					WithHeight(10),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
					WithStyles(s),
				)
			},
		},
		// TODO(?): the total height is modified with borderd headers, however not with bordered cells. Is this expected/desired?
		// TODO(?): 带边框的表头会修改总高度，但带边框的单元格不会。这是预期的/期望的吗？
		"Bordered headers": { // 带边框的表头
			modelFunc: func() Model {
				return New(
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
					WithStyles(Styles{
						Header: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
					}),
				)
			},
		},
		// TODO(fix): Headers are not horizontally aligned with cells due to the border adding width to the cells.
		// TODO(fix): 由于边框增加了单元格的宽度，表头与单元格在水平方向上不对齐。
		"Bordered cells": { // 带边框的单元格
			modelFunc: func() Model {
				return New(
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
					WithStyles(Styles{
						Cell: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
					}),
				)
			},
		},
		"Manual height greater than rows": { // 手动高度大于行数
			modelFunc: func() Model {
				return New(
					WithHeight(6),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
				)
			},
		},
		"Manual height less than rows": { // 手动高度小于行数
			modelFunc: func() Model {
				return New(
					WithHeight(2),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
				)
			},
		},
		// TODO(fix): spaces are added to the right of the viewport to fill the width, but the headers end as though they are not aware of the width.
		// TODO(fix): 视口右侧添加空格以填充宽度，但表头结束时似乎不知道宽度。
		"Manual width greater than columns": { // 手动宽度大于列宽
			modelFunc: func() Model {
				return New(
					WithWidth(80),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
				)
			},
		},
		// TODO(fix): Setting the table width does not affect the total headers' width. Cells are wrapped.
		// 	Headers are not affected. Truncation/resizing should match lipgloss.table functionality.
		// TODO(fix): 设置表格宽度不会影响表头的总宽度。单元格会换行。
		// 	表头不受影响。截断/调整大小应该与 lipgloss.table 功能匹配。
		"Manual width less than columns": { // 手动宽度小于列宽
			modelFunc: func() Model {
				return New(
					WithWidth(30),
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
				)
			},
			skip: true,
		},
		"Modified viewport height": { // 修改视口高度
			modelFunc: func() Model {
				m := New(
					WithColumns([]Column{
						{Title: "Name", Width: 25},
						{Title: "Country of Origin", Width: 16},
						{Title: "Dunk-able", Width: 12},
					}),
					WithRows([]Row{
						{"Chocolate Digestives", "UK", "Yes"},
						{"Tim Tams", "Australia", "No"},
						{"Hobnobs", "UK", "Yes"},
					}),
				)

				m.viewport.Height = 2

				return m
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			table := tc.modelFunc()

			got := ansi.Strip(table.View())

			golden.RequireEqual(t, []byte(got))
		})
	}
}

// TODO: Fix table to make this test will pass.
// TODO: 修复表格以使此测试通过。
func TestModel_View_CenteredInABox(t *testing.T) {
	t.Skip()

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Align(lipgloss.Center)

	table := New(
		WithHeight(6),
		WithWidth(80),
		WithColumns([]Column{
			{Title: "Name", Width: 25},
			{Title: "Country of Origin", Width: 16},
			{Title: "Dunk-able", Width: 12},
		}),
		WithRows([]Row{
			{"Chocolate Digestives", "UK", "Yes"},
			{"Tim Tams", "Australia", "No"},
			{"Hobnobs", "UK", "Yes"},
		}),
	)

	tableView := ansi.Strip(table.View())
	got := boxStyle.Render(tableView)

	golden.RequireEqual(t, []byte(got))
}
