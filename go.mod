module github.com/purpose168/bubbles-cn

go 1.24.2

replace (
	github.com/purpose168/bubbletea-cn => ../bubbletea-cn
	github.com/purpose168/charm-experimental-packages-cn/ansi => ../charm-experimental-packages-cn/ansi
	github.com/purpose168/charm-experimental-packages-cn/cellbuf => ../charm-experimental-packages-cn/cellbuf
	github.com/purpose168/charm-experimental-packages-cn/exp/golden => ../charm-experimental-packages-cn/exp/golden
	github.com/purpose168/charm-experimental-packages-cn/term => ../charm-experimental-packages-cn/term
	github.com/purpose168/lipgloss-cn => ../lipgloss-cn
)

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/atotto/clipboard v0.1.4
	github.com/aymanbagabas/go-udiff v0.3.1
	github.com/charmbracelet/harmonica v0.2.0
	github.com/dustin/go-humanize v1.0.1
	github.com/lucasb-eyer/go-colorful v1.3.0
	github.com/mattn/go-runewidth v0.0.19
	github.com/muesli/termenv v0.16.0
	github.com/purpose168/bubbletea-cn v0.0.0-00010101000000-000000000000
	github.com/purpose168/charm-experimental-packages-cn/ansi v0.10.2
	github.com/purpose168/charm-experimental-packages-cn/exp/golden v0.0.0-20250609102027-b60490452b30
	github.com/purpose168/lipgloss-cn v0.0.0-00010101000000-000000000000
	github.com/rivo/uniseg v0.4.7
	github.com/sahilm/fuzzy v0.1.1
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.4.1 // indirect
	github.com/charmbracelet/x/ansi v0.11.3 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.10.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.6.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/purpose168/charm-experimental-packages-cn/cellbuf v0.0.13 // indirect
	github.com/purpose168/charm-experimental-packages-cn/term v0.2.1 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
