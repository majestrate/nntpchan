package srnd

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const asdKey = "3c7850617b4fe116c98f4ed4a2eaf00ab219d16dd6351d9ee786f9fc710bad55"

func TestSeedToKeypair(t *testing.T) {
	seed := parseTripcodeSecret("asd")
	pk, _ := naclSeedToKeyPair(seed)
	hexpk := hexify(pk)
	if hexpk != asdKey {
		t.Logf("%s != %s", asdKey, hexpk)
		t.Fail()
	}
}

func TestSign(t *testing.T) {

	msgid := "<wut@wut.wut>"
	seed := randbytes(32)
	pk, sec := naclSeedToKeyPair(seed)
	sig := msgidFrontendSign(sec, msgid)
	t.Log(sig)
	if !verifyFrontendSig(hexify(pk), sig, msgid) {
		t.Fail()
	}
}

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
