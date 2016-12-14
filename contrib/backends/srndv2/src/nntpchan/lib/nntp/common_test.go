package nntp

import (
	"testing"
)

func TestGenMessageID(t *testing.T) {
	msgid := GenMessageID("test.tld")
	t.Logf("generated id %s", msgid)
	if !msgid.Valid() {
		t.Logf("invalid generated message-id %s", msgid)
		t.Fail()
	}
	msgid = GenMessageID("<><><>")
	t.Logf("generated id %s", msgid)
	if msgid.Valid() {
		t.Logf("generated valid message-id when it should've been invalid %s", msgid)
		t.Fail()
	}
}

func TestMessageIDHash(t *testing.T) {
	msgid := GenMessageID("test.tld")
	lh := msgid.LongHash()
	sh := msgid.ShortHash()
	bh := msgid.Blake2Hash()
	t.Logf("long=%s short=%s blake2=%s", lh, sh, bh)
}

func TestValidNewsgroup(t *testing.T) {
	g := Newsgroup("overchan.test")
	if !g.Valid() {
		t.Logf("%s is invalid?", g)
		t.Fail()
	}
}

func TestInvalidNewsgroup(t *testing.T) {
	g := Newsgroup("asd.asd.asd.&&&")
	if g.Valid() {
		t.Logf("%s should be invalid", g)
		t.Fail()
	}
}
