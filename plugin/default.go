package plugin

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/helper"
	"github.com/jqs7/bb"
)

type Default struct{ bb.Base }

func (d *Default) Run() {
	if d.FromPrivate {
		switch d.getStatus() {
		case "auth":
			d.auth(d.Message.Text)
		case "broadcast":
			d.bc(d.Message.Text)
			d.setStatus("")
		case "trans":
			result := d.translator(d.Message.Text)
			d.NewMessage(d.ChatID, result).Send()
			d.setStatus("")
		default:
			if conf.CategoriesSet.Has(d.Message.Text) {
				// custom keyboard reply
				if !d.isAuthed() {
					d.sendQuestion()
					return
				}
				groups := conf.List2SliceInConf(d.Message.Text)
				result := make([]string, len(groups))
				for k, v := range groups {
					reg := regexp.MustCompile("^(.+) (http(s)?://(.*))$")
					strs := reg.FindAllStringSubmatch(v, -1)
					if !reg.MatchString(v) {
						result[k] = v
					}
					for _, v := range strs {
						result[k] = helper.ToMarkdown(v[1], v[2])
					}
				}
				msgContent := strings.Join(result, "\n")
				msgContent = strings.Replace(msgContent, "\\n", "", -1)
				d.NewMessage(d.ChatID, msgContent).
					MarkdownMode().DisableWebPagePreview().Send()
			} else {
				if len(d.Args) > 0 {
					d.turing(d.Message.Text)
					return
				}
				photo := d.Message.Photo
				if len(photo) > 0 {
					go d.NewChatAction(d.ChatID).UploadPhoto().Send()

					fileID := photo[len(photo)-1].FileID
					link, _ := d.GetLink(fileID)
					path := helper.Downloader(link, fileID)

					mime := helper.FileMime(path)
					size := helper.FileSize(path)
					bar := helper.BarCode(path)
					vcn := helper.Vim_cn_Uploader(path)
					os.Remove(path)

					s := fmt.Sprintf("%s %s\n%s\n%s", mime, size, vcn, bar)
					d.NewMessage(d.ChatID, s).
						DisableWebPagePreview().
						ReplyToMessageID(d.Message.MessageID).Send()
					return
				}
			}
		}
	}
}

func (d *Default) getStatus() string {
	return conf.Redis.Get("tgStatus:" + strconv.Itoa(d.Message.From.ID)).Val()
}

func (d *Default) auth(answer string) {
	qs := conf.GetQuestions()
	index := time.Now().Hour() % len(qs)
	answer = strings.ToLower(answer)
	answer = strings.TrimSpace(answer)
	if d.FromPrivate {
		if d.isAuthed() {
			d.NewMessage(d.ChatID,
				"已经验证过了，你还想验证，你是不是傻？⊂彡☆))д`)`").
				ReplyToMessageID(d.Message.MessageID).Send()
			return
		}

		if qs[index].A.Has(answer) {
			conf.Redis.SAdd("tgAuthUser", strconv.Itoa(d.Message.From.ID))
			log.Printf("%d --- %s Auth OK\n",
				d.Message.From.ID, d.Message.From.UserName)
			d.NewMessage(d.ChatID,
				"验证成功喵~！\n原来你不是外星人呢😊").Send()
			d.setStatus("")
			d.NewMessage(d.ChatID,
				conf.List2StringInConf("help")).Send()
		} else {
			log.Printf("%d --- %s Auth Fail\n",
				d.Message.From.ID, d.Message.From.UserName)
			d.NewMessage(d.ChatID,
				"答案不对不对！你一定是外星人！不跟你玩了喵！\n"+
					"重新验证一下吧\n请问："+qs[index].Q).Send()
		}
	}
}

func (d *Default) isAuthed() bool {
	if conf.Redis.SIsMember("tgAuthUser",
		strconv.Itoa(d.Message.From.ID)).Val() {
		return true
	}
	return false
}

func (d *Default) sendQuestion() {
	qs := conf.GetQuestions()
	index := time.Now().Hour() % len(qs)
	d.NewMessage(d.Message.From.ID,
		"需要通过中文验证之后才能使用本功能哟~\n请问："+
			qs[index].Q+"\n把答案发给奴家就可以了呢").
		Send()
	d.setStatus("auth")
}

func (d *Default) isMaster() bool {
	master := conf.GetItem("master")
	if d.Message.From.UserName == master {
		return true
	}
	return false
}

func (d *Default) setStatus(status string) {
	if status == "" {
		conf.Redis.Del("tgStatus:" + strconv.Itoa(d.Message.From.ID))
		return
	}
	conf.Redis.Set("tgStatus:"+strconv.Itoa(d.Message.From.ID), status, -1)
}
