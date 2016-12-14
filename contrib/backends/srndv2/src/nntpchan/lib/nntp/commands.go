package nntp

type Command string

func (c Command) String() string {
	return string(c)
}

// command to list newsgroups
const CMD_Newsgroups = Command("NEWSGROUPS 0 0 GMT")

// create group command for a newsgroup
func CMD_Group(g Newsgroup) Command {
	return Command("GROUP " + g.String())
}

const CMD_XOver = Command("XOVER 0")

func CMD_Article(msgid MessageID) Command {
	return Command("ARTICLE " + msgid.String())
}

func CMD_Head(msgid MessageID) Command {
	return Command("HEAD " + msgid.String())
}

const CMD_Capabilities = Command("CAPABILITIES")
