package plugin

import (
	"strings"

	"github.com/jqs7/bb"
)

type Markdown struct{ bb.Base }

func (m *Markdown) Run() {
	if len(m.Args) > 1 {
		s := strings.Join(m.Args[1:], " ")
		err := m.NewMessage(m.ChatID, s).
			DisableWebPagePreview().
			MarkdownMode().Send().Err
		if err != nil {
			m.NewMessage(m.ChatID, err.Error()).Send()
		}
	}
}
