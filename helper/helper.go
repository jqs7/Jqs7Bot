package helper

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bieber/barcode"
	"github.com/franela/goreq"
	"github.com/pyk/byten"
	"gopkg.in/h2non/filetype.v0"
)

func To2dSlice(in []string, x, y int) [][]string {
	out := [][]string{}
	var begin, end int
	for i := 0; i < y; i++ {
		end += x
		if end >= len(in) {
			out = append(out, in[begin:])
			break
		}
		out = append(out, in[begin:end])
		begin = end
	}
	return out
}

func Downloader(link, fileName string) string {
	resp, err := goreq.Request{
		Method: "GET",
		Uri:    link,
	}.Do()
	if err != nil {
		return ""
	}

	filePath := filepath.Join(os.TempDir(), fileName)
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return ""
	}
	io.Copy(f, resp.Body)
	return filePath
}

func FileMime(filePath string) string {
	buf, _ := ioutil.ReadFile(filePath)
	kind, err := filetype.Match(buf)
	if err != nil {
		return ""
	}
	return kind.MIME.Value
}

func FileSize(filePath string) string {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return ""
	}
	s, err := f.Stat()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", HumanByte(s.Size())...)
}

func BarCode(filePath string) string {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return ""
	}
	switch FileMime(filePath) {
	case "image/jpeg":
		src, err := jpeg.Decode(f)
		if err != nil {
			return ""
		}

		img := barcode.NewImage(src)
		scanner := barcode.NewScanner().SetEnabledAll(true)

		symbols, _ := scanner.ScanImage(img)
		var buf bytes.Buffer
		for _, s := range symbols {
			buf.WriteString(s.Data + " ")
		}
		return buf.String()
	default:
		return ""
	}
}

func Vim_cn_Uploader(filePath string) string {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return ""
	}
	client := new(http.Client)
	req, err := FileUploadReq("https://img.vim-cn.com/", "name", f)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	s := strings.TrimSpace(body.String())

	return s
}

func FileUploadReq(uri, paramName string, file *os.File) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, file.Name())
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	//for k, v := range params {
	//writer.WriteField(k, v)
	//}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}

func HumanByte(in ...interface{}) (out []interface{}) {
	for _, v := range in {
		switch v.(type) {
		case int, int64, uint64:
			s := fmt.Sprintf("%d", v)
			i, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, byten.Size(i))
		case float64:
			s := fmt.Sprintf("%.2f", v)
			out = append(out, s)
		default:
			out = append(out, v)
		}
	}
	return out
}
