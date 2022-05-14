package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
	"os"
)

var (
	boxStyle = gloss.NewStyle().
			Border(gloss.RoundedBorder()).
			BorderForeground(gloss.Color("196")).
			Padding(2)
	previewStyle  = gloss.NewStyle()
	SelectedStyle = previewStyle.Copy().
			Background(gloss.Color("52")).
			Foreground(gloss.Color("196"))
	subtle = gloss.Color("235")
)

type Model struct {
	runeset Runeset
	path    string
	view    string
	fx, fy  int
	editing int
	ex, ey  int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			switch m.view {
			case "all":
				_ = WriteRunesetFile(m.runeset, m.path)
				return m, tea.Quit
			case "edit":
				m.view = "all"
			}

		case "enter", " ":
			switch m.view {
			case "all":
				m.view = "edit"
				m.editing = m.fy*32 + m.fx
			case "edit":
				m.runeset[m.editing][m.ey] ^= 1 << m.ex
			}
		case "up":
			switch m.view {
			case "all":
				m.fy -= 1
			case "edit":
				m.ey -= 1
			}
		case "down":
			switch m.view {
			case "all":
				m.fy += 1
			case "edit":
				m.ey += 1
			}
		case "left":
			switch m.view {
			case "all":
				m.fx -= 1
			case "edit":
				m.ex -= 1
			}
		case "right":
			switch m.view {
			case "all":
				m.fx += 1
			case "edit":
				m.ex += 1
			}
		}
	}
	switch m.view {
	case "all":
		if m.fx >= 32 {
			m.fx -= 32
		} else if m.fx < 0 {
			m.fx += 32
		}

		if m.fy >= 8 {
			m.fy -= 8
		} else if m.fy < 0 {
			m.fy += 8
		}
	case "edit":
		if m.ex >= 8 {
			m.ex -= 8
		} else if m.ex < 0 {
			m.ex += 8
		}

		if m.ey >= 8 {
			m.ey -= 8
		} else if m.ey < 0 {
			m.ey += 8
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := ""

	tx, ty, _ := term.GetSize(int(os.Stdout.Fd()))

	if tx < 5*32 || ty < 5*8 {
		return gloss.Place(tx, ty, gloss.Center, gloss.Center,
			boxStyle.Render(fmt.Sprintf("Terminal window too small! (%dx%d)", tx, ty)),
			gloss.WithWhitespaceChars("▞"), gloss.WithWhitespaceForeground(gloss.Color("235")))
	}

	switch m.view {
	case "all":
		spacerH := gloss.NewStyle().Foreground(subtle).Render("▞\n▞\n▞\n▞")
		spacerV := gloss.NewStyle().Foreground(subtle).Render("▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞▞")

		if m.view == "all" {
			for j := 0; j < 8; j++ {
				line := ""
				for i := 0; i < 32; i++ {
					preview, _ := m.runeset.Preview(j*32 + i)

					if i == m.fx && j == m.fy {
						preview = SelectedStyle.Render(preview)
					} else {
						preview = previewStyle.Render(preview)
					}

					line = gloss.JoinHorizontal(gloss.Top, line, spacerH, preview)
				}
				if j == 0 {
					s = line
					continue
				}

				s = gloss.JoinVertical(gloss.Left, s, spacerV, line)
			}
		}

		s += fmt.Sprintf("\n0x%02x (%02d, %02d)", m.fy*32+m.fx, m.fx, m.fy)

		return gloss.Place(tx, ty, gloss.Center, gloss.Center, s,
			gloss.WithWhitespaceChars("▞"), gloss.WithWhitespaceForeground(subtle))
	case "edit":
		img, _ := m.runeset.ToImg(m.editing)

		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				v := ""
				if img[i][j] {
					v = "█"
				} else {
					v = " "
				}

				if m.ex == j && m.ey == i {
					v = SelectedStyle.Render(v)
				}

				s += v

				if j != 7 {
					s += " │ "
				}
			}
			if i != 7 {
				s += "\n──┼───┼───┼───┼───┼───┼───┼──\n"
			}
		}

		return gloss.Place(tx, ty, gloss.Center, gloss.Center, boxStyle.Render(s),
			gloss.WithWhitespaceChars("▞"), gloss.WithWhitespaceForeground(subtle))
	}

	return fmt.Sprintf("Error: View `%s` does not exist...", m.view)
}

func main() {
	var path string

	flag.StringVar(&path, "p", "", "The path to the file you want to create or edit.")
	flag.Parse()

	fmt.Println("Reading file:", path)

	r, err, found := ReadRunesetFile(path)

	switch err {
	case nil:
		break
	case FileReadError:
		fmt.Println("Could not read file!")
		return
	case BytesLengthError:
		fmt.Println("Oh no! This shouldn't have happened!")
		return
	default:
		fmt.Println(err)
		return
	}

	if !found {
		fmt.Println("File not found, creating new runeset...")
		err := WriteRunesetFile(r, path)

		if err != nil {
			fmt.Println("Failed to create new runeset file!")
			return
		} else {
			fmt.Println("Runeset created successfully...")
		}
	} else {
		fmt.Println("Successfully read file...")
	}

	m := Model{
		runeset: r,
		path:    path,
		view:    "all",
	}

	fmt.Println("Starting program...")

	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Println(err)
		return
	}
}
