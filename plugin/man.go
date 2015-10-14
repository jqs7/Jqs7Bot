package plugin

import (
	"strconv"
	"strings"

	"github.com/jqs7/Jqs7Bot/conf"
)

type Man struct{ Default }

func (m *Man) Run() {
	switch m.Args[0] {
	case "/setman":
		if len(m.Args) >= 3 {
			value := strings.Join(m.Args[2:], " ")
			if !m.isAuthed() {
				m.sendQuestion()
				return
			}

			if m.FromGroup {
				conf.Redis.HSet(
					"tgMan:"+strconv.Itoa(m.ChatID),
					m.Args[1], value)
				m.NewMessage(m.ChatID,
					m.Args[1]+":\n"+value).Send()
			}
		}
	case "/rmman":
		fields := m.Args[1:]
		for k := range fields {
			conf.Redis.HDel("tgMan:"+
				strconv.Itoa(m.ChatID), fields[k])
		}
		m.listMan()
	case "/man":
		if len(m.Args) == 1 {
			m.listMan()
		} else {
			if !m.FromGroup {
				return
			}
			if m.Args[1] == "man" && !conf.Redis.HExists(
				"tgMan:"+strconv.Itoa(m.ChatID), "man").Val() {
				m.NewMessage(m.ChatID, "你在慢慢个什么鬼啦！(σﾟ∀ﾟ)σ").Send()
				return
			}
			m.NewMessage(m.ChatID,
				conf.Redis.HGet("tgMan:"+strconv.Itoa(m.ChatID),
					m.Args[1]).Val()).Send()
		}
	}
}

func (m *Man) listMan() {
	if m.FromGroup {
		var result string
		fields := conf.Redis.HGetAllMap("tgMan:" +
			strconv.Itoa(m.ChatID)).Val()
		for k := range fields {
			result += k + "\n"
		}
		m.NewMessage(m.ChatID, result).Send()
	}
}
