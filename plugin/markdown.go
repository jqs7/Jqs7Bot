package plugin

import (
	"strings"

	"github.com/jqs7/bb"
)

type Markdown struct{ bb.Base }

func (m *Markdown) Run() {
	if len(m.Args) > 1 {
		s := strings.Join(m.Args[1:], " ")
		m.NewMessage(m.ChatID, s).
			DisableWebPagePreview().
			MarkdownMode().Send()
	}
}
