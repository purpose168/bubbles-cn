# Bubbles

<p>
  <img src="https://stuff.charm.sh/bubbles/bubbles-github.png" width="233" alt="Bubbles 标志">
</p>

[![最新版本](https://img.shields.io/github/release/charmbracelet/bubbles.svg)](https://github.com/purpose168/bubbles-cn/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/charmbracelet/bubbles)
[![构建状态](https://github.com/purpose168/bubbles-cn/workflows/build/badge.svg)](https://github.com/purpose168/bubbles-cn/actions)
[![Go ReportCard](https://goreportcard.com/badge/charmbracelet/bubbles)](https://goreportcard.com/report/charmbracelet/bubbles)

Bubble Tea 应用程序的一些组件。这些组件已在生产环境中用于 [Glow][glow] 和 [许多其他应用程序][otherstuff]。

[glow]: https://github.com/charmbracelet/glow
[otherstuff]: https://github.com/purpose168/bubbletea-cn/#bubble-tea-in-the-wild

## Spinner

<img src="https://stuff.charm.sh/bubbles-examples/spinner.gif" width="400" alt="Spinner 示例">

一个加载指示器，用于表示正在进行某种操作。有几个默认的样式，但你也可以传递自己的"帧"。

- [示例代码，基本 spinner](https://github.com/purpose168/bubbletea-cn/blob/main/examples/spinner/main.go)
- [示例代码，各种 spinner](https://github.com/purpose168/bubbletea-cn/blob/main/examples/spinners/main.go)

## 文本输入

<img src="https://stuff.charm.sh/bubbles-examples/textinput.gif" width="400" alt="文本输入示例">

一个文本输入字段，类似于 HTML 中的 `<input type="text">`。支持 Unicode、粘贴、当值超过元素宽度时的原位滚动，以及许多自定义选项。

- [示例代码，单个字段](https://github.com/purpose168/bubbletea-cn/blob/main/examples/textinput/main.go)
- [示例代码，多个字段](https://github.com/purpose168/bubbletea-cn/blob/main/examples/textinputs/main.go)

## 文本区域

<img src="https://stuff.charm.sh/bubbles-examples/textarea.gif" width="400" alt="文本区域示例">

一个文本区域字段，类似于 HTML 中的 `<textarea />`。允许跨多行输入。支持 Unicode、粘贴、当值超过元素宽度和高度时的垂直滚动，以及许多自定义选项。

- [示例代码，聊天输入](https://github.com/purpose168/bubbletea-cn/blob/main/examples/chat/main.go)
- [示例代码，故事时间输入](https://github.com/purpose168/bubbletea-cn/blob/main/examples/textarea/main.go)

## 表格

<img src="https://stuff.charm.sh/bubbles-examples/table.gif" width="400" alt="表格示例">

一个用于显示和导航表格数据（列和行）的组件。支持垂直滚动和许多自定义选项。

- [示例代码，国家和人口](https://github.com/purpose168/bubbletea-cn/blob/main/examples/table/main.go)

## 进度条

<img src="https://stuff.charm.sh/bubbles-examples/progress.gif" width="800" alt="进度条示例">

一个简单、可定制的进度指示器，可通过 [Harmonica][harmonica] 实现可选的动画效果。支持纯色和渐变填充。空和填充的字符可以设置为你喜欢的任何内容。百分比读数可自定义，也可以完全省略。

- [动画示例](https://github.com/purpose168/bubbletea-cn/blob/main/examples/progress-animated/main.go)
- [静态示例](https://github.com/purpose168/bubbletea-cn/blob/main/examples/progress-static/main.go)

[harmonica]: https://github.com/charmbracelet/harmonica

## 分页器

<img src="https://stuff.charm.sh/bubbles-examples/pagination.gif" width="200" alt="分页器示例">

一个用于处理分页逻辑并可选绘制分页 UI 的组件。支持"点样式"分页（类似于你在 iOS 上看到的）和数字页码，但你也可以仅使用此组件的逻辑并以任何你喜欢的方式可视化分页。

- [示例代码](https://github.com/purpose168/bubbletea-cn/blob/main/examples/paginator/main.go)

## 视口

<img src="https://stuff.charm.sh/bubbles-examples/viewport.gif" width="600" alt="视口示例">

一个用于垂直滚动内容的视口。可选包含标准分页器键绑定和鼠标滚轮支持。对于使用备用屏幕缓冲区的应用程序，提供高性能模式。

- [示例代码](https://github.com/purpose168/bubbletea-cn/blob/main/examples/pager/main.go)

此组件与 [Reflow][reflow] 配合使用效果良好，可实现 ANSI 感知的缩进和文本换行。

[reflow]: https://github.com/muesli/reflow

## 列表

<img src="https://stuff.charm.sh/bubbles-examples/list.gif" width="600" alt="列表示例">

一个可定制、功能齐全的组件，用于浏览一组项目。具有分页、模糊过滤、自动生成帮助、活动指示器和状态消息等功能，所有这些功能都可以根据需要启用和禁用。源自 [Glow][glow]。

- [示例代码，默认列表](https://github.com/purpose168/bubbletea-cn/blob/main/examples/list-default/main.go)
- [示例代码，简单列表](https://github.com/purpose168/bubbletea-cn/blob/main/examples/list-simple/main.go)
- [示例代码，所有功能](https://github.com/purpose168/bubbletea-cn/blob/main/examples/list-fancy/main.go)

## 文件选择器

<img src="https://vhs.charm.sh/vhs-yET2HNiJNEbyqaVfYuLnY.gif" width="600" alt="文件选择器示例">

一个用于从文件系统中选择文件的可定制组件。可以浏览目录并选择文件，可选限制为特定文件扩展名。

- [示例代码](https://github.com/purpose168/bubbletea-cn/blob/main/examples/file-picker/main.go)

## 计时器

一个简单、灵活的倒计时组件。更新频率和输出可以根据你的需要进行自定义。

<img src="https://stuff.charm.sh/bubbles-examples/timer.gif" width="400" alt="计时器示例">

- [示例代码](https://github.com/purpose168/bubbletea-cn/blob/main/examples/timer/main.go)

## 秒表

<img src="https://stuff.charm.sh/bubbles-examples/stopwatch.gif" width="400" alt="秒表示例">

一个简单、灵活的计时组件。更新频率和输出可以根据你的需要进行自定义。

- [示例代码](https://github.com/purpose168/bubbletea-cn/blob/main/examples/stopwatch/main.go)

## 帮助

<img src="https://stuff.charm.sh/bubbles-examples/help.gif" width="500" alt="帮助示例">

一个可定制的水平迷你帮助视图，可根据你的键绑定自动生成。它具有单行和多行模式，用户可以选择在两者之间切换。如果终端对于内容来说太宽，它会优雅地截断。

- [示例代码](https://github.com/purpose168/bubbletea-cn/blob/main/examples/help/main.go)

## 按键

一个用于管理键绑定的非可视化组件。它对于允许用户重新映射键绑定以及生成与你的键绑定相对应的帮助视图非常有用。

```go
// KeyMap 定义了应用程序的键绑定映射
type KeyMap struct {
    Up key.Binding    // 向上移动的键绑定
    Down key.Binding  // 向下移动的键绑定
}

// DefaultKeyMap 是默认的键绑定映射
var DefaultKeyMap = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("k", "up"),        // 实际的键绑定
        key.WithHelp("↑/k", "向上移动"), // 对应的帮助文本
    ),
    Down: key.NewBinding(
        key.WithKeys("j", "down"),      // 实际的键绑定
        key.WithHelp("↓/j", "向下移动"),  // 对应的帮助文本
    ),
}

// Update 处理消息并更新模型状态
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, DefaultKeyMap.Up):
            // 用户按下了向上键
        case key.Matches(msg, DefaultKeyMap.Down):
            // 用户按下了向下键
        }
    }
    return m, nil
}
```

## 更多精彩内容

要查看社区维护的 Bubbles，请访问 [Charm & Friends](https://github.com/charm-and-friends/additional-bubbles)。制作了一个很酷的 Bubble 想分享？欢迎提交 [PR](https://github.com/charm-and-friends/additional-bubbles?tab=readme-ov-file#what-is-a-complete-project)！

## 贡献

请参阅 [contributing][contribute]。

[contribute]: https://github.com/purpose168/bubbles-cn/contribute

## 反馈

我们很想听听你对这个项目的想法。随时给我们留言！

- [Twitter](https://twitter.com/charmcli)
- [联邦宇宙](https://mastodon.social/@charmcli)
- [Discord](https://charm.sh/chat)

## 许可证

[MIT](https://github.com/purpose168/bubbletea-cn/raw/main/LICENSE)


