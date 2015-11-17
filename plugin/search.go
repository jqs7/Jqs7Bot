package plugin

import (
	"strings"

	"github.com/jqs7/Jqs7Bot/conf"
)

type Search struct{ Default }

func (s *Search) Run() {
	if !s.isAuthed() {
		s.sendQuestion()
		return
	}

	if s.FromPrivate {
		if len(s.Args) > 1 {
			result := []string{}
			for _, v := range conf.Groups {
				arg := strings.ToLower(s.Args[1])
				lower := strings.ToLower(v)
				if strings.Contains(lower, arg) {
					result = append(result, v)
				}
			}
			s.NewMessage(s.ChatID, strings.Join(result, "\n")).Send()
		}
	}
}
