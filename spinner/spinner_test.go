package spinner_test

import (
	"testing"

	"github.com/purpose168/bubbles-cn/spinner"
)

// TestSpinnerNew 测试加载动画的创建功能
func TestSpinnerNew(t *testing.T) {
	// assertEqualSpinner 断言两个加载动画相等
	assertEqualSpinner := func(t *testing.T, exp, got spinner.Spinner) {
		t.Helper()

		if exp.FPS != got.FPS {
			t.Errorf("期望 %d FPS，但得到了 %d", exp.FPS, got.FPS)
		}

		if e, g := len(exp.Frames), len(got.Frames); e != g {
			t.Fatalf("期望 %d 帧，但得到了 %d", e, g)
		}

		for i, e := range exp.Frames {
			if g := got.Frames[i]; e != g {
				t.Errorf("期望帧索引 %d 的值为 %q，但得到了 %q", i, e, g)
			}
		}
	}
	// 测试默认加载动画
	t.Run("default", func(t *testing.T) {
		s := spinner.New()

		assertEqualSpinner(t, spinner.Line, s.Spinner)
	})

	// 测试自定义加载动画
	t.Run("WithSpinner", func(t *testing.T) {
		customSpinner := spinner.Spinner{
			Frames: []string{"a", "b", "c", "d"},
			FPS:    16,
		}

		s := spinner.New(spinner.WithSpinner(customSpinner))

		assertEqualSpinner(t, customSpinner, s.Spinner)
	})

	// 测试所有预定义的加载动画
	tests := map[string]spinner.Spinner{
		"Line":    spinner.Line,    // 线条加载动画
		"Dot":     spinner.Dot,     // 点加载动画
		"MiniDot": spinner.MiniDot, // 小点加载动画
		"Jump":    spinner.Jump,    // 跳跃加载动画
		"Pulse":   spinner.Pulse,   // 脉冲加载动画
		"Points":  spinner.Points,  // 点加载动画
		"Globe":   spinner.Globe,   // 地球加载动画
		"Moon":    spinner.Moon,    // 月亮加载动画
		"Monkey":  spinner.Monkey,  // 猴子加载动画
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			assertEqualSpinner(t, spinner.New(spinner.WithSpinner(s)).Spinner, s)
		})
	}
}
