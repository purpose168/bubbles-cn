package viewport

import (
	"strings"
	"testing"
)

const defaultHorizontalStep = 6 // 默认水平滚动步长

// TestNew 测试创建新的视口模型
func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("通过 New 创建时的默认值", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		m.horizontalStep = defaultHorizontalStep // v2 版本时移除

		if !m.initialized {
			t.Errorf("通过 New 创建时，模型应该已初始化")
		}

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("默认 horizontalStep 应为 %d，实际为 %d", defaultHorizontalStep, m.horizontalStep)
		}

		if m.MouseWheelDelta != 3 {
			t.Errorf("默认 MouseWheelDelta 应为 3，实际为 %d", m.MouseWheelDelta)
		}

		if !m.MouseWheelEnabled {
			t.Error("鼠标滚轮默认应启用")
		}
	})
}

// TestSetInitialValues 测试设置初始值
func TestSetInitialValues(t *testing.T) {
	t.Parallel()

	t.Run("默认 horizontalStep", func(t *testing.T) {
		t.Parallel()

		m := Model{}
		m.horizontalStep = defaultHorizontalStep // v2 版本时移除
		m.setInitialValues()

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("默认 horizontalStep 应为 %d，实际为 %d", defaultHorizontalStep, m.horizontalStep)
		}
	})
}

// TestSetHorizontalStep 测试设置水平步长
func TestSetHorizontalStep(t *testing.T) {
	t.Parallel()

	t.Run("修改默认值", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		m.horizontalStep = defaultHorizontalStep // v2 版本时移除

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("默认 horizontalStep 应为 %d，实际为 %d", defaultHorizontalStep, m.horizontalStep)
		}

		newStep := 8
		m.SetHorizontalStep(newStep)
		if m.horizontalStep != newStep {
			t.Errorf("horizontalStep 应为 %d，实际为 %d", newStep, m.horizontalStep)
		}
	})

	t.Run("不允许负值", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		m.horizontalStep = defaultHorizontalStep // v2 版本时移除

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("默认 horizontalStep 应为 %d，实际为 %d", defaultHorizontalStep, m.horizontalStep)
		}

		zero := 0
		m.SetHorizontalStep(-1)
		if m.horizontalStep != zero {
			t.Errorf("horizontalStep 应为 %d，实际为 %d", zero, m.horizontalStep)
		}
	})
}

// TestScrollLeft 测试向左滚动
func TestScrollLeft(t *testing.T) {
	t.Parallel()

	zeroPosition := 0 // 零位置

	t.Run("零位置", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		m.longestLineWidth = 100
		if m.xOffset != zeroPosition {
			t.Errorf("默认缩进应为 %d，实际为 %d", zeroPosition, m.xOffset)
		}

		m.ScrollLeft(m.horizontalStep)
		if m.xOffset != zeroPosition {
			t.Errorf("缩进应为 %d，实际为 %d", zeroPosition, m.xOffset)
		}
	})

	t.Run("滚动", func(t *testing.T) {
		t.Parallel()
		m := New(10, 10)
		m.horizontalStep = defaultHorizontalStep // v2 版本时移除
		m.longestLineWidth = 100
		if m.xOffset != zeroPosition {
			t.Errorf("默认缩进应为 %d，实际为 %d", zeroPosition, m.xOffset)
		}

		m.xOffset = defaultHorizontalStep * 2
		m.ScrollLeft(m.horizontalStep)
		newIndent := defaultHorizontalStep
		if m.xOffset != newIndent {
			t.Errorf("缩进应为 %d，实际为 %d", newIndent, m.xOffset)
		}
	})
}

// TestScrollRight 测试向右滚动
func TestScrollRight(t *testing.T) {
	t.Parallel()

	t.Run("滚动", func(t *testing.T) {
		t.Parallel()

		zeroPosition := 0

		m := New(10, 10)
		m.SetHorizontalStep(defaultHorizontalStep)
		m.SetContent("Some line that is longer than width")
		if m.xOffset != zeroPosition {
			t.Errorf("默认缩进应为 %d，实际为 %d", zeroPosition, m.xOffset)
		}

		m.ScrollRight(m.horizontalStep)
		newIndent := defaultHorizontalStep
		if m.xOffset != newIndent {
			t.Errorf("缩进应为 %d，实际为 %d", newIndent, m.xOffset)
		}
	})
}

// TestResetIndent 测试重置缩进
func TestResetIndent(t *testing.T) {
	t.Parallel()

	t.Run("重置", func(t *testing.T) {
		t.Parallel()

		zeroPosition := 0

		m := New(10, 10)
		m.xOffset = 500

		m.SetXOffset(0)
		if m.xOffset != zeroPosition {
			t.Errorf("缩进应为 %d，实际为 %d", zeroPosition, m.xOffset)
		}
	})
}

// TestVisibleLines 测试可见行
func TestVisibleLines(t *testing.T) {
	t.Parallel()

	// 默认测试列表：来自《空洞骑士》游戏中 Zote 的57条戒律
	defaultList := []string{
		`57 Precepts of narcissistic comedy character Zote from an awesome "Hollow knight" game (https://store.steampowered.com/app/367520/Hollow_Knight/).`,
		`戒律一：'永远赢得你的战斗'。输掉战斗不会让你获得任何东西，也不会教会你什么。赢得你的战斗，或者干脆不要参与其中！`,
		`戒律二：'永远不要让他们嘲笑你'。傻瓜们嘲笑一切，甚至嘲笑他们的上位者。但要注意，笑声并非无害！笑声像疾病一样传播，很快每个人都会嘲笑你。你需要迅速打击这种变态欢乐的根源，以阻止其蔓延。`,
		`戒律三：'永远保持休息'。战斗和冒险会对你的身体造成损伤。当你休息时，你的身体会变得更强壮并自我修复。休息时间越长，你就越强大。`,
		`戒律四：'忘记你的过去'。过去是痛苦的，思考过去只会给你带来痛苦。相反，思考一些其他的事情，比如未来，或者一些食物。`,
		`戒律五：'力量战胜力量'。你的对手很强？没关系！只需用更多的力量来克服他们的力量，他们很快就会被击败。`,
		`戒律六：'选择你自己的命运'。我们的长辈教导说，我们的命运在我们出生之前就已经被选择好了。我不同意。`,
		`戒律七：'不要为死者哀悼'。当我们死后，事情对我们来说是变得更好还是更糟？没有办法知道，所以我不应该为此哀悼。或者为此庆祝。`,
		`戒律八：'独自旅行'。你不能依赖任何人，也没有人会永远忠诚。因此，不应该有人成为你永久的伴侣。`,
		`戒律九：'保持你的家整洁'。你的家是你保存最珍贵财产的地方——你自己。因此，你应该努力让它保持整洁。`,
		`戒律十：'保持你的武器锋利'。我确保我的武器 '生命终结者' 始终保持锋利。这使得切割东西变得更容易。`,
		`戒律十一：'母亲总会背叛你'。这条戒律不言自明。`,
		`戒律十二：'保持你的斗篷干燥'。如果你的斗篷湿了，尽快把它弄干。穿湿斗篷令人不快，并可能导致疾病。`,
		`戒律十三：'永远不要害怕'。恐惧只会阻碍你。面对恐惧需要巨大的努力。因此，你一开始就不应该害怕。`,
		`戒律十四：'尊重你的上位者'。如果有人在力量或智力上或两者都比你强，你需要向他们表示尊重。不要忽视或嘲笑他们。`,
		`戒律十五：'一个敌人，一击'。你应该只用一击来击败敌人。任何更多的都是浪费。此外，通过在战斗中计算你的击打次数，你将知道你击败了多少敌人。`,
		`...`,
	}

	t.Run("空列表", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		list := m.visibleLines()

		if len(list) != 0 {
			t.Errorf("列表应为空，实际有 %d 个元素", len(list))
		}
	})

	t.Run("空列表：带缩进", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		list := m.visibleLines()
		m.xOffset = 5

		if len(list) != 0 {
			t.Errorf("列表应为空，实际有 %d 个元素", len(list))
		}
	})

	t.Run("列表", func(t *testing.T) {
		t.Parallel()
		numberOfLines := 10

		m := New(10, numberOfLines)
		m.SetContent(strings.Join(defaultList, "\n"))

		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("列表应有 %d 行，实际有 %d 行", numberOfLines, len(list))
		}

		lastItemIdx := numberOfLines - 1
		// 如果行不符合视口宽度，我们对其进行修剪
		shouldGet := defaultList[lastItemIdx][:m.Width]
		if list[lastItemIdx] != shouldGet {
			t.Errorf(`第 %d 个列表项应为 '%s'，实际为 '%s'`, lastItemIdx, shouldGet, list[lastItemIdx])
		}
	})

	t.Run("列表：带 Y 偏移", func(t *testing.T) {
		t.Parallel()
		numberOfLines := 10

		m := New(10, numberOfLines)
		m.SetContent(strings.Join(defaultList, "\n"))
		m.YOffset = 5

		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("列表应有 %d 行，实际有 %d 行", numberOfLines, len(list))
		}

		if list[0] == defaultList[0] {
			t.Error("由于 Y 偏移，列表的第一项不应与初始列表的第一项相同")
		}

		lastItemIdx := numberOfLines - 1
		// 如果行不符合视口宽度，我们对其进行修剪
		shouldGet := defaultList[m.YOffset+lastItemIdx][:m.Width]
		if list[lastItemIdx] != shouldGet {
			t.Errorf(`第 %d 个列表项应为 '%s'，实际为 '%s'`, lastItemIdx, shouldGet, list[lastItemIdx])
		}
	})

	t.Run("列表：带 Y 偏移：水平滚动", func(t *testing.T) {
		t.Parallel()
		numberOfLines := 10

		m := New(10, numberOfLines)
		m.horizontalStep = defaultHorizontalStep // v2 版本时移除
		m.SetContent(strings.Join(defaultList, "\n"))
		m.SetYOffset(7)

		// 默认列表
		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("列表应有 %d 行，实际有 %d 行", numberOfLines, len(list))
		}

		lastItem := numberOfLines - 1
		defaultLastItem := len(defaultList) - 1
		if list[lastItem] != defaultList[defaultLastItem] {
			t.Errorf("第 %d 个列表项应与第 %d 个默认列表项相同", lastItem, defaultLastItem)
		}

		perceptPrefix := "戒律"
		if !strings.HasPrefix(list[0], perceptPrefix) {
			t.Errorf("第一个列表项必须有前缀 %s", perceptPrefix)
		}

		// 向右滚动
		m.ScrollRight(m.horizontalStep)
		list = m.visibleLines()

		newPrefix := perceptPrefix[m.xOffset:]
		if !strings.HasPrefix(list[0], newPrefix) {
			t.Errorf("第一个列表项必须有前缀 %s，实际为 %s", newPrefix, list[0])
		}

		if list[lastItem] != "" {
			t.Errorf("最后一项应为空，实际为 '%s'", list[lastItem])
		}

		// 向左滚动
		m.ScrollLeft(m.horizontalStep)
		list = m.visibleLines()
		if !strings.HasPrefix(list[0], perceptPrefix) {
			t.Errorf("第一个列表项必须有前缀 %s", perceptPrefix)
		}

		if list[lastItem] != defaultList[defaultLastItem] {
			t.Errorf("第 %d 个列表项应与第 %d 个默认列表项相同", lastItem, defaultLastItem)
		}
	})

	t.Run("列表：带 2 字符宽度符号：水平滚动", func(t *testing.T) {
		t.Parallel()

		const horizontalStep = 5

		initList := []string{
			"あいうえお",
			"Aあいうえお",
			"あいうえお",
			"Aあいうえお",
		}
		numberOfLines := len(initList)

		m := New(20, numberOfLines)
		m.lines = initList
		m.longestLineWidth = 30 // 技巧：不检查此测试用例的右侧过度滚动

		// 默认列表
		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("列表应有 %d 行，实际有 %d 行", numberOfLines, len(list))
		}

		lastItemIdx := numberOfLines - 1
		initLastItem := len(initList) - 1
		shouldGet := initList[initLastItem]
		if list[lastItemIdx] != shouldGet {
			t.Errorf("第 %d 个列表项应与第 %d 个默认列表项相同", lastItemIdx, initLastItem)
		}

		// 向右滚动
		m.ScrollRight(horizontalStep)
		list = m.visibleLines()

		for i := range list {
			cutLine := "うえお"
			if list[i] != cutLine {
				t.Errorf("行必须为 `%s`，实际为 `%s`", cutLine, list[i])
			}
		}

		// 向左滚动
		m.ScrollLeft(horizontalStep)
		list = m.visibleLines()
		for i := range list {
			if list[i] != initList[i] {
				t.Errorf("行必须为 `%s`，实际为 `%s`", list[i], initList[i])
			}
		}

		// 第二次向左滚动在缩进为 0 时不应改变列表
		m.xOffset = 0
		m.ScrollLeft(horizontalStep)
		list = m.visibleLines()
		for i := range list {
			if list[i] != initList[i] {
				t.Errorf("行必须为 `%s`，实际为 `%s`", list[i], initList[i])
			}
		}
	})
}

// TestRightOverscroll 测试右侧过度滚动
func TestRightOverscroll(t *testing.T) {
	t.Parallel()

	t.Run("防止右侧过度滚动", func(t *testing.T) {
		t.Parallel()
		content := "Content is short"
		m := New(len(content)+1, 5)
		m.SetContent(content)

		for i := 0; i < 10; i++ {
			m.ScrollRight(m.horizontalStep)
		}

		visibleLines := m.visibleLines()
		visibleLine := visibleLines[0]

		if visibleLine != content {
			t.Error("可见行应保持与内容相同")
		}
	})
}
