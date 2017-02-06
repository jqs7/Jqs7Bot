package plugin

import (
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
)

type About struct{ bb.Base }

func (a *About) Run() {
	a.NewMessage(a.Message.From.ID,
		conf.List2StringInConf("about")).Send()
}
