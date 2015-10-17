package plugin

import (
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
)

type OtherResources struct{ bb.Base }

func (o *OtherResources) Run() {
	o.NewMessage(o.Message.From.ID,
		conf.List2StringInConf("其他资源")).Send()
}
