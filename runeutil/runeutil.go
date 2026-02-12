// Package runeutil 为 Bubbles 提供一个实用函数，
// 该函数可以处理包含符文的按键消息。
package runeutil

import (
	"unicode"
	"unicode/utf8"
)

// Sanitizer 是一个辅助工具，用于想要处理按键消息中符文的气泡小部件。
type Sanitizer interface {
	// Sanitize 从 KeyRunes 消息中移除控制字符，
	// 并可选择地用指定字符替换换行符/回车符/制表符。
	//
	// 如果可能，符文数组会被就地修改。在这种情况下，
	// 返回的切片是移除/翻译控制字符后缩短的原始切片。
	Sanitize(runes []rune) []rune
}

// NewSanitizer 构建一个符文清理器。
func NewSanitizer(opts ...Option) Sanitizer {
	s := sanitizer{
		replaceNewLine: []rune("\n"),
		replaceTab:     []rune("    "),
	}
	for _, o := range opts {
		s = o(s)
	}
	return &s
}

// Option 是可以传递给 Sanitize() 的选项类型。
type Option func(sanitizer) sanitizer

// ReplaceTabs 用指定字符串替换制表符。
func ReplaceTabs(tabRepl string) Option {
	return func(s sanitizer) sanitizer {
		s.replaceTab = []rune(tabRepl)
		return s
	}
}

// ReplaceNewlines 用指定字符串替换换行符。
func ReplaceNewlines(nlRepl string) Option {
	return func(s sanitizer) sanitizer {
		s.replaceNewLine = []rune(nlRepl)
		return s
	}
}

func (s *sanitizer) Sanitize(runes []rune) []rune {
	// dstrunes 是我们存储结果的地方。
	dstrunes := runes[:0:len(runes)]
	// copied 指示 dstrunes 是 runes 的别名还是副本。
	// 当 dst 超过 src 时我们需要副本。
	// 我们使用此作为优化，以避免在输出小于或等于输入的常见情况下分配新的符文切片。
	copied := false

	for src := 0; src < len(runes); src++ {
		r := runes[src]
		switch {
		case r == utf8.RuneError:
			// 跳过

		case r == '\r' || r == '\n':
			if len(dstrunes)+len(s.replaceNewLine) > src && !copied {
				dst := len(dstrunes)
				dstrunes = make([]rune, dst, len(runes)+len(s.replaceNewLine))
				copy(dstrunes, runes[:dst])
				copied = true
			}
			dstrunes = append(dstrunes, s.replaceNewLine...)

		case r == '\t':
			if len(dstrunes)+len(s.replaceTab) > src && !copied {
				dst := len(dstrunes)
				dstrunes = make([]rune, dst, len(runes)+len(s.replaceTab))
				copy(dstrunes, runes[:dst])
				copied = true
			}
			dstrunes = append(dstrunes, s.replaceTab...)

		case unicode.IsControl(r):
			// 其他控制字符：跳过。

		default:
			// 保留字符。
			dstrunes = append(dstrunes, runes[src])
		}
	}
	return dstrunes
}

// sanitizer 符文清理器结构体
type sanitizer struct {
	replaceNewLine []rune // 替换换行符
	replaceTab     []rune // 替换制表符
}
