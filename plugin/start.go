package plugin

import "github.com/jqs7/Jqs7Bot/conf"

type Start struct{ Default }

func (s *Start) Run() {
	if !s.isAuthed() {
		s.sendQuestion()
		return
	}
	s.NewMessage(s.ChatID,
		conf.List2StringInConf("help")).Send()
}
