package plugin

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/jqs7/bb"
)

type Base64 struct{ bb.Base }

func (b *Base64) Run() {
	switch b.Args[0] {
	case "/e64":
		if b.Message.ReplyToMessage != nil &&
			b.Message.ReplyToMessage.Text != "" {
			b.NewMessage(b.ChatID, E64(b.Message.ReplyToMessage.Text)).
				Send()
		} else if len(b.Args) >= 2 {
			in := strings.Join(b.Args[1:], " ")
			b.NewMessage(b.ChatID, E64(in)).Send()
		}
	case "/d64":
		if len(b.Args) >= 2 {
			in := strings.Join(b.Args[1:], " ")
			b.NewMessage(b.ChatID, D64(in)).Send()
		}
	}
}

func E64(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}

func D64(in string) string {
	out, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "解码系统出现故障，请查看弹药是否填充无误"
	}
	if utf8.Valid(out) {
		return string(out)
	}
	return "解码结果包含不明物体，群组娘已将之上交国家"
}
