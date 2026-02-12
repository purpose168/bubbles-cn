package progress

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

const (
	AnsiReset = "\x1b[0m" // ANSI 重置序列
)

// TestGradient 测试渐变填充功能
func TestGradient(t *testing.T) {

	colA := "#FF0000" // 第一种颜色
	colB := "#00FF00" // 第二种颜色

	var p Model
	var descr string

	for _, scale := range []bool{false, true} {
		opts := []Option{
			WithColorProfile(termenv.TrueColor), WithoutPercentage(),
		}
		if scale {
			descr = "带缩放渐变的进度条"
			opts = append(opts, WithScaledGradient(colA, colB))
		} else {
			descr = "带渐变的进度条"
			opts = append(opts, WithGradient(colA, colB))
		}

		t.Run(descr, func(t *testing.T) {
			p = New(opts...)

			// 通过对空字符串着色然后截断随后的重置序列来构建期望的颜色
			sb := strings.Builder{}
			sb.WriteString(termenv.String("").Foreground(p.color(colA)).String())
			expFirst := strings.Split(sb.String(), AnsiReset)[0]
			sb.Reset()
			sb.WriteString(termenv.String("").Foreground(p.color(colB)).String())
			expLast := strings.Split(sb.String(), AnsiReset)[0]

			for _, width := range []int{3, 5, 50} {
				p.Width = width
				res := p.ViewAs(1.0)

				// 通过在 p.Full+AnsiReset 处分割来从进度条中提取颜色，
				// 这样我们就只剩下颜色序列
				colors := strings.Split(res, string(p.Full)+AnsiReset)

				// 丢弃最后一个颜色，因为它是空的（进度条最后一个字符之后没有新颜色）
				colors = colors[0 : len(colors)-1]

				if expFirst != colors[0] {
					t.Errorf("期望进度条的第一个颜色是第一种渐变颜色 %q，但得到了 %q", expFirst, colors[0])
				}

				if expLast != colors[len(colors)-1] {
					t.Errorf("期望进度条的最后一个颜色是第二种渐变颜色 %q，但得到了 %q", expLast, colors[len(colors)-1])
				}
			}
		})
	}

}
