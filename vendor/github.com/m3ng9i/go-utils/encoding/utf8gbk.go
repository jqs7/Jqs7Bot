package encoding

import "bytes"
import "io"
import "io/ioutil"
import "golang.org/x/text/encoding/simplifiedchinese"
import "golang.org/x/text/transform"


func GbkToUtf8Reader(r io.Reader) io.Reader {
    return transform.NewReader(r, simplifiedchinese.GBK.NewDecoder())
}


func Utf8ToGbkReader(r io.Reader) io.Reader {
    return transform.NewReader(r, simplifiedchinese.GBK.NewEncoder())
}


func GbkToUtf8BytesReader(b []byte) io.Reader {
    return GbkToUtf8Reader(bytes.NewReader(b))
}


func Utf8ToGbkBytesReader(b []byte) io.Reader {
    return Utf8ToGbkReader(bytes.NewReader(b))
}


func GbkToUtf8(b []byte) ([]byte, error) {
    d, e := ioutil.ReadAll(GbkToUtf8BytesReader(b))
    if e != nil {
        return nil, e
    }
    return d, nil
}


func Utf8ToGbk(b []byte) ([]byte, error) {
    d, e := ioutil.ReadAll(Utf8ToGbkBytesReader(b))
    if e != nil {
        return nil, e
    }
    return d, nil
}

