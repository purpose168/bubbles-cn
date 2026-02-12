package runeutil

import (
	"testing"
	"unicode/utf8"
)

// TestSanitize 测试 Sanitize 方法清理符文的功能
func TestSanitize(t *testing.T) {
	// 测试数据：输入字符串和期望的输出字符串
	td := []struct {
		input, output string // 输入和输出
	}{
		{"", ""},                     // 空字符串
		{"x", "x"},                   // 普通字符
		{"\n", "XX"},                 // 换行符
		{"\na\n", "XXaXX"},           // 换行符和字符
		{"\n\n", "XXXX"},             // 多个换行符
		{"\t", ""},                   // 制表符（默认替换为空）
		{"hello", "hello"},           // 普通字符串
		{"hel\nlo", "helXXlo"},       // 字符串中的换行符
		{"hel\rlo", "helXXlo"},       // 字符串中的回车符
		{"hel\tlo", "hello"},         // 字符串中的制表符（默认替换为空）
		{"he\n\nl\tlo", "heXXXXllo"}, // 多个控制字符
		{"he\tl\n\nlo", "helXXXXlo"}, // 混合控制字符
		{"hel\x1blo", "hello"},       // 无效的 UTF-8 字符
		{"hello\xc2", "hello"},       // 无效的 UTF-8 字符
	}

	for _, tc := range td {
		// 将输入字符串转换为符文切片
		runes := make([]rune, 0, len(tc.input))
		b := []byte(tc.input)
		for i, w := 0, 0; i < len(b); i += w {
			var r rune
			r, w = utf8.DecodeRune(b[i:])
			runes = append(runes, r)
		}
		t.Logf("输入符文: %+v", runes)
		// 创建清理器，用 "XX" 替换换行符，用空字符串替换制表符
		s := NewSanitizer(ReplaceNewlines("XX"), ReplaceTabs(""))
		result := s.Sanitize(runes)
		rs := string(result)
		if tc.output != rs {
			t.Errorf("%q: 期望 %q，但得到了 %q (%+v)", tc.input, tc.output, rs, result)
		}
	}
}
