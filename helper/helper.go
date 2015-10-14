package helper

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
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

func Vim_cn_Uploader(f *os.File) string {
	client := new(http.Client)
	req, err := FileUploadReq("https://img.vim-cn.com/", "name", f)
	resp, err := client.Do(req)
	if err != nil {
		return "嘟嘟！希望号已失联~( ＞﹏＜  )"
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "生命转换系统发生了[叮咚~]"
	}
	defer resp.Body.Close()
	s := strings.TrimSpace(body.String())

	return fmt.Sprintf("[img.vim-cn.com/%s](%s)", s[len(s)-9:], s)
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
