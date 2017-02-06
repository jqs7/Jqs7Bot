package plugin

import (
	"strings"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/helper"
)

type Search struct{ Default }

func (s *Search) Run() {
	if !s.isAuthed() {
		s.sendQuestion()
		return
	}

	if len(s.Args) > 1 {
		result := []string{}
		for _, v := range conf.Groups {
			arg := strings.ToLower(s.Args[1])
			lower := strings.ToLower(v.GroupName)
			if strings.Contains(lower, arg) {
				if v.GroupURL != "" {
					result = append(result, helper.ToMarkdownLink(v.GroupName, v.GroupURL))
				} else {
					result = append(result, v.GroupName)
				}
			}
		}
		if len(result) != 0 {
			s.NewMessage(s.Message.From.ID, strings.Join(result, "\n")).
				MarkdownMode().DisableWebPagePreview().Send()
		} else {
			s.NewMessage(s.Message.From.ID, "搜索大失败喵(/￣ˇ￣)/").Send()
		}
	}
}
