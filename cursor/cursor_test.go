package cursor

import (
	"sync"
	"testing"
	"time"
)

// TestBlinkCmdDataRace 测试 [Cursor.blinkTag] 上的数据竞争。
//
// 最初的 [Model.BlinkCmd] 实现返回一个闭包，该闭包捕获了指针接收器：
//
//	return func() tea.Msg {
//		defer cancel()
//		<-ctx.Done()
//		if ctx.Err() == context.DeadlineExceeded {
//			return BlinkMsg{id: m.id, tag: m.blinkTag}
//		}
//		return blinkCanceled{}
//	}
//
// 在以下情况下会发生“m.blinkTag”上的竞争：
//  1. [Model.BlinkCmd] 被调用，例如通过从
//     ["github.com/purpose168/bubbletea-cn".Model.Update] 调用 [Model.Focus]；
//  2. ["github.com/purpose168/bubbletea-cn".handleCommands] 足够繁忙，以至于它无法接收和
//     执行 [Model.BlinkCmd]，例如由于其他长时间运行的命令；
//  3. 至少经过 [Mode.BlinkSpeed] 时间；
//  4. [Model.BlinkCmd] 再次被调用；
//  5. ["github.com/purpose168/bubbletea-cn".handleCommands] 最终接收并执行原始
//     闭包。
//
// 即使这不是正式的竞争，获取的标签值在语义上也是不正确的（可能是当前值，而不是创建闭包时的值）。
func TestBlinkCmdDataRace(t *testing.T) {
	m := New()
	cmd := m.BlinkCmd()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		time.Sleep(m.BlinkSpeed * 3)
		cmd()
	}()
	go func() {
		defer wg.Done()
		time.Sleep(m.BlinkSpeed * 2)
		m.BlinkCmd()
	}()
	wg.Wait()
}
