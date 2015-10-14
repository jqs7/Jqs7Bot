package plugin

import (
	"log"
	"strings"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
	"github.com/st3v/translator"
	"github.com/st3v/translator/microsoft"
)

type Trans struct{ bb.Base }

func (t *Trans) Run() {
	if t.Message.ReplyToMessage != nil &&
		t.Message.ReplyToMessage.Text != "" &&
		len(t.Args) < 2 {
		in := t.Message.ReplyToMessage.Text
		result := t.translator(in)
		t.NewMessage(t.ChatID, result).Send()
	} else if len(t.Args) >= 2 {
		in := strings.Join(t.Args[1:], " ")
		result := t.translator(in)
		t.NewMessage(t.ChatID, result).Send()
	}
}

func (t *Trans) translator(in string) string {
	result := make(chan string)
	typingChan := make(chan bool)
	go func() {
		t.NewChatAction(t.ChatID).Typing().Send()
		typingChan <- true
	}()
	go func() {
		result <- ZhTrans(in)
	}()
	<-typingChan
	return <-result
}

type MsTrans struct {
	t translator.Translator
}

func (m *MsTrans) New() {
	m.t = microsoft.NewTranslator(
		conf.GetItem("msTransId"),
		conf.GetItem("msTransSecret"))
}

func (m *MsTrans) Detect(in string) (string, error) {
	return m.t.Detect(in)
}

func (m *MsTrans) Trans(in, from, to string) (string, error) {
	return m.t.Translate(in, from, to)
}

func ZhTrans(in string) (out string) {
	m := &MsTrans{}
	m.New()
	from, err := m.Detect(in)
	if err != nil {
		log.Println(err.Error())
		return "警报！弹药系统过载！请放宽后重试"
	}
	switch from {
	case "zh-CHS", "zh-CHT":
		out, err = m.Trans(in, from, "en")
	default:
		out, err = m.Trans(in, from, "zh-CHS")
	}
	if err != nil {
		return "可怜的群组娘被母舰放逐了X﹏X"
	}
	return out
}
