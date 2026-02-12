# 常见问题

以下是关于 `list` bubble（列表气泡）的一些最常见问题。

## 添加自定义项

要创建自定义项，您需要完成几个步骤。首先，它们需要实现 `list.Item` 和 `list.DefaultItem` 接口。

```go
// Item 是出现在列表中的一个项。
type Item interface {
	// FilterValue 是我们在过滤列表时针对此项进行过滤所使用的值。
	FilterValue() string
}
```

```go
// DefaultItem 描述了一个设计用于与 DefaultDelegate（默认委托）配合使用的项。
type DefaultItem interface {
	Item
	Title() string      // 返回项的标题
	Description() string // 返回项的描述
}
```

您可以在我们的 [Kancli][kancli] 项目中看到一个可运行的示例，该项目专门为 Bubble Tea 中的列表和组合视图教程而构建。

[VIDEO](https://youtu.be/ZA93qgdLUzM)

## 自定义样式

列表项的渲染（和行为）是通过 [`ItemDelegate`][itemDelegate] 接口完成的。起初可能会有点令人困惑，但它允许列表变得非常灵活和强大。

如果您只是想更改默认样式，可以这样做：

```go
import "github.com/purpose168/bubbles-cn/list"

// 创建一个新的默认委托
d := list.NewDefaultDelegate()

// 更改颜色
c := lipgloss.Color("#6f03fc") // 设置紫色
d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(c).BorderLeftForeground(c) // 设置选中标题的前景色和左边框颜色
d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy() // 在此处复用标题样式

// 使用我们的委托初始化列表模型
width, height := 80, 40 // 设置宽度和高度
l := list.New(listItems, d, width, height) // 创建列表

// 您也可以动态更改委托
l.SetDelegate(d) // 设置新的委托
```

这段代码将替换 [`list-default` 示例][listDefault]中的[这一行][replacedLine]。

要完全控制列表项的渲染方式，您也可以定义自己的 `ItemDelegate`（[示例][customDelegate]）。

[kancli]: https://github.com/charmbracelet/kancli/blob/main/main.go#L45
[itemDelegate]: https://pkg.go.dev/github.com/purpose168/bubbles-cn/list#ItemDelegate
[replacedLine]: https://github.com/purpose168/bubbletea-cn/blob/main/examples/list-default/main.go#L77
[listDefault]: https://github.com/purpose168/bubbletea-cn/tree/main/examples/list-default
[customDelegate]: https://github.com/purpose168/bubbletea-cn/blob/main/examples/list-simple/main.go#L29-L50
