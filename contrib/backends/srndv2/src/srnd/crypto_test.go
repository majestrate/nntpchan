package srnd

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestVerify(t *testing.T) {
	d := filepath.Join("testdata", "article.test.txt")

	f, e := os.Open(d)
	if e != nil {
		t.Logf("os.Open returned error: %s", e)
		t.Fail()
		return
	}

	r := bufio.NewReader(f)

	msg, er := readMIMEHeader(r)
	if er != nil {
		t.Logf("readMIMEHeader returned error: %s", er)
		t.Fail()
		return
	}

	b := &io.LimitedReader{
		R: msg.Body,
		N: 1000000000,
	}

	err := read_message_body(b, msg.Header, nil, ioutil.Discard, true, func(msg NNTPMessage) {
		return
	})
	if err != nil {
		t.Logf("read_message_body returned error: %s", err)
		t.Fail()
		return
	}
}
