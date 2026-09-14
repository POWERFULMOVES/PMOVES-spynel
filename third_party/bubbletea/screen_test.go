package tea

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestClearScreenReestablishesAlternateOwnership(t *testing.T) {
	var output bytes.Buffer
	r := newRenderer(&output, false, 60).(*standardRenderer)
	r.handleMessages(WindowSizeMsg{Width: 80, Height: 24})
	r.enterAltScreen()
	r.write("application frame")
	r.flush()
	output.Reset()
	r.clearScreen()
	cleared := output.String()
	exit := strings.Index(cleared, ansi.ResetAltScreenSaveCursorMode)
	enter := strings.Index(cleared, ansi.SetAltScreenSaveCursorMode)
	if exit < 0 || enter <= exit || !strings.Contains(cleared[:exit], ansi.ResetStyle+ansi.CursorHomePosition+ansi.EraseScreenBelow) {
		t.Fatalf("alternate ownership was not cleared and restored: %q", cleared)
	}
	r.write("application frame")
	r.flush()
	if !r.altScreen() || !strings.Contains(output.String()[len(cleared):], "application frame") {
		t.Fatal("clearing did not retain alternate ownership and invalidate the frame")
	}
}

func TestClearMsg(t *testing.T) {
	tests := []struct {
		name     string
		cmds     sequenceMsg
		expected string
	}{
		{
			name:     "clear_screen",
			cmds:     []Cmd{ClearScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[2J\x1b[H\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "altscreen",
			cmds:     []Cmd{EnterAltScreen, ExitAltScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1049h\x1b[2J\x1b[H\x1b[?25l\x1b[m\x1b[H\x1b[J\x1b[?1049l\x1b[?25l\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "altscreen_autoexit",
			cmds:     []Cmd{EnterAltScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1049h\x1b[2J\x1b[H\x1b[?25l\x1b[H\rsuccess\x1b[K\r\n\x1b[K\x1b[2;H\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l\x1b[m\x1b[H\x1b[J\x1b[?1049l\x1b[?25h",
		},
		{
			name:     "mouse_cellmotion",
			cmds:     []Cmd{EnableMouseCellMotion},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1002h\x1b[?1006h\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "mouse_allmotion",
			cmds:     []Cmd{EnableMouseAllMotion},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1003h\x1b[?1006h\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "mouse_disable",
			cmds:     []Cmd{EnableMouseAllMotion, DisableMouse},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1003h\x1b[?1006h\x1b[?1002l\x1b[?1003l\x1b[?1006l\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "cursor_hide",
			cmds:     []Cmd{HideCursor},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?25l\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "cursor_hideshow",
			cmds:     []Cmd{HideCursor, ShowCursor},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?25l\x1b[?25h\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "bp_stop_start",
			cmds:     []Cmd{DisableBracketedPaste, EnableBracketedPaste},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?2004l\x1b[?2004h\rsuccess\x1b[K\r\n\x1b[K\r\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testModel{}
			p := NewProgram(m, WithInput(&in), WithOutput(&buf))

			test.cmds = append([]Cmd{func() Msg { return WindowSizeMsg{80, 24} }}, test.cmds...)
			test.cmds = append(test.cmds, Quit)
			go p.Send(test.cmds)

			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}

			if buf.String() != test.expected {
				t.Errorf("expected embedded sequence:\n%q\ngot:\n%q", test.expected, buf.String())
			}
		})
	}
}
