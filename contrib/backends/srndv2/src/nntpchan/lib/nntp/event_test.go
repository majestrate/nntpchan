package nntp

import (
	"testing"
)

func TestTAKETHISParse(t *testing.T) {
	msgid := GenMessageID("test.tld")
	ev := stream_cmd_TAKETHIS(msgid)
	t.Logf("event: %s", ev)
	if ev.MessageID() != msgid {
		t.Logf("%s != %s, event was %s", msgid, ev.MessageID(), ev)
		t.Fail()
	}
	if ev.Command() != "TAKETHIS" {
		t.Logf("%s != TAKETHIS, event was %s", ev.Command(), ev)
		t.Fail()
	}
	if !ev.Valid() {
		t.Logf("%s is invalid stream event", ev)
		t.Fail()
	}
}

func TestCHECKParse(t *testing.T) {
	msgid := GenMessageID("test.tld")
	ev := stream_cmd_CHECK(msgid)
	t.Logf("event: %s", ev)
	if ev.MessageID() != msgid {
		t.Logf("%s != %s, event was %s", msgid, ev.MessageID(), ev)
		t.Fail()
	}
	if ev.Command() != "CHECK" {
		t.Logf("%s != CHECK, event was %s", ev.Command(), ev)
		t.Fail()
	}
	if !ev.Valid() {
		t.Logf("%s is invalid stream event", ev)
		t.Fail()
	}
}

func TestInvalidStremEvent(t *testing.T) {
	str := "asd"
	ev := StreamEvent(str)
	t.Logf("invalid str=%s ev=%s", str, ev)
	if ev.Valid() {
		t.Logf("invalid CHECK command is valid? %s", ev)
		t.Fail()
	}

	str = "asd asd"
	ev = StreamEvent(str)
	t.Logf("invalid str=%s ev=%s", str, ev)

	if ev.Valid() {
		t.Logf("invalid CHECK command is valid? %s", ev)
		t.Fail()
	}
}
