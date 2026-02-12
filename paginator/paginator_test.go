package paginator

import (
	"testing"

	tea "github.com/purpose168/bubbletea-cn"
)

// TestNew 测试 New 函数创建新模型的功能
func TestNew(t *testing.T) {
	model := New()

	if model.PerPage != 1 {
		t.Errorf("PerPage = %d, expected %d", model.PerPage, 1)
	}
	if model.TotalPages != 1 {
		t.Errorf("TotalPages = %d, expected %d", model.TotalPages, 1)
	}

	perPage := 42
	totalPages := 42

	model = New(
		WithPerPage(perPage),
		WithTotalPages(totalPages),
	)

	if model.PerPage != perPage {
		t.Errorf("PerPage = %d, expected %d", model.PerPage, perPage)
	}
	if model.TotalPages != totalPages {
		t.Errorf("TotalPages = %d, expected %d", model.TotalPages, totalPages)
	}
}

// TestSetTotalPages 测试 SetTotalPages 函数设置总页数的功能
func TestSetTotalPages(t *testing.T) {
	tests := []struct {
		name         string // 测试用例名称
		items        int    // 要设置的项目总数
		initialTotal int    // 测试用例的初始总页数
		expected     int    // 调用 SetTotalPages 函数后的期望值
	}{
		{"Less than one page", 5, 1, 5},        // 少于一页
		{"Exactly one page", 10, 1, 10},        // 恰好一页
		{"More than one page", 15, 1, 15},      // 多于一页
		{"negative value for page", -10, 1, 1}, // 页数为负值
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New()
			if model.TotalPages != tt.initialTotal {
				model.SetTotalPages(tt.initialTotal)
			}
			model.SetTotalPages(tt.items)
			if model.TotalPages != tt.expected {
				t.Errorf("TotalPages = %d, expected %d", model.TotalPages, tt.expected)
			}
		})
	}
}

// TestPrevPage 测试 PrevPage 函数向前翻页的功能
func TestPrevPage(t *testing.T) {
	tests := []struct {
		name       string // 测试用例名称
		totalPages int    // 为测试用例设置的总页数
		page       int    // 测试的初始页码
		expected   int    // 期望的页码
	}{
		{"Go to previous page", 10, 1, 0}, // 转到上一页
		{"Stay on first page", 5, 0, 0},   // 停留在第一页
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New()
			model.SetTotalPages(tt.totalPages)
			model.Page = tt.page

			model, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}})
			if model.Page != tt.expected {
				t.Errorf("PrevPage() = %d, expected %d", model.Page, tt.expected)
			}
		})
	}
}

// TestNextPage 测试 NextPage 函数向后翻页的功能
func TestNextPage(t *testing.T) {
	tests := []struct {
		name       string // 测试用例名称
		totalPages int    // 总页数
		page       int    // 初始页码
		expected   int    // 期望的页码
	}{
		{"Go to next page", 2, 0, 1},   // 转到下一页
		{"Stay on last page", 2, 1, 1}, // 停留在最后一页
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New()
			model.SetTotalPages(tt.totalPages)
			model.Page = tt.page

			model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}})
			if model.Page != tt.expected {
				t.Errorf("NextPage() = %d, expected %d", model.Page, tt.expected)
			}
		})
	}
}

// TestOnLastPage 测试 OnLastPage 函数判断是否在最后一页的功能
func TestOnLastPage(t *testing.T) {
	tests := []struct {
		name       string // 测试用例名称
		page       int    // 当前页码
		totalPages int    // 总页数
		expected   bool   // 期望的返回值
	}{
		{"On last page", 1, 2, true},      // 在最后一页
		{"Not on last page", 0, 2, false}, // 不在最后一页
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New()
			model.SetTotalPages(tt.totalPages)
			model.Page = tt.page

			if result := model.OnLastPage(); result != tt.expected {
				t.Errorf("OnLastPage() = %t, expected %t", result, tt.expected)
			}
		})
	}
}

// TestOnFirstPage 测试 OnFirstPage 函数判断是否在第一页的功能
func TestOnFirstPage(t *testing.T) {
	tests := []struct {
		name       string // 测试用例名称
		page       int    // 当前页码
		totalPages int    // 总页数
		expected   bool   // 期望的返回值
	}{
		{"On first page", 0, 2, true},      // 在第一页
		{"Not on first page", 1, 2, false}, // 不在第一页
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New()
			model.SetTotalPages(tt.totalPages)
			model.Page = tt.page

			if result := model.OnFirstPage(); result != tt.expected {
				t.Errorf("OnFirstPage() = %t, expected %t", result, tt.expected)
			}
		})
	}
}

// TestItemsOnPage 测试 ItemsOnPage 函数返回当前页项目数量的功能
func TestItemsOnPage(t *testing.T) {
	testCases := []struct {
		currentPage   int // 为测试用例设置的当前页码
		totalPages    int // 为测试用例设置的总页数
		totalItems    int // 总项目数
		expectedItems int // 当前页期望的项目数
	}{
		{1, 10, 10, 1},
		{3, 10, 10, 1},
		{7, 10, 10, 1},
	}

	for _, tc := range testCases {
		model := New()
		model.Page = tc.currentPage
		model.SetTotalPages(tc.totalPages)
		if actualItems := model.ItemsOnPage(tc.totalItems); actualItems != tc.expectedItems {
			t.Errorf("ItemsOnPage() returned %d, expected %d for total items %d", actualItems, tc.expectedItems, tc.totalItems)
		}
	}
}
