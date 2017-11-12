package srnd

import (
	"bytes"
	"errors"
	"log"
	"mime"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	stdtemplate "text/template"
)

var ErrNoThumbnailerFound = errors.New("no thumbnailer found")

type ThumbnailRule struct {
	AcceptsMimeType string
	CommandTemplate string
}

func (th *ThumbnailRule) GenerateThumbnail(infname, outfname string, env map[string]string) (err error) {
	env["infile"] = infname
	env["outfile"] = outfname
	buff := new(bytes.Buffer)
	err = stdtemplate.Must(stdtemplate.New("").Parse(th.CommandTemplate)).Execute(buff, env)
	if err == nil {
		args := strings.Split(buff.String(), " ")
		cmd := exec.Command(args[0], args[1:]...)
		var out []byte
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Println(buff.String(), string(out))
		}
	}
	return
}

func (th *ThumbnailRule) Accepts(mimetype string) bool {
	return th.AcceptsMimeType == "*" || regexp.MustCompilePOSIX(th.AcceptsMimeType).MatchString(mimetype)
}

func (th *ThumbnailConfig) Load(opts map[string]string) {
	for k, v := range opts {
		th.rules = append(th.rules, ThumbnailRule{
			AcceptsMimeType: k,
			CommandTemplate: v,
		})
	}
	sort.Sort(th)
}

func (th *ThumbnailConfig) Len() int {
	return len(th.rules)
}

func (th *ThumbnailConfig) Swap(i, j int) {
	th.rules[i], th.rules[j] = th.rules[j], th.rules[i]
}

func (th *ThumbnailConfig) Less(i, j int) bool {
	return th.rules[i].AcceptsMimeType >= th.rules[j].AcceptsMimeType
}

func (th *ThumbnailConfig) FindRulesFor(mimetype string) (rules []ThumbnailRule) {
	for _, rule := range th.rules {
		if rule.Accepts(mimetype) {
			rules = append(rules, rule)
		}
	}
	return
}

func (th *ThumbnailConfig) GenerateThumbnail(infname, outfname string, env map[string]string) (err error) {
	mimeType := mime.TypeByExtension(filepath.Ext(infname))
	rules := th.FindRulesFor(mimeType)
	for _, rule := range rules {
		err = rule.GenerateThumbnail(infname, outfname, env)
		if err == nil {
			log.Println("made thumbnail for", infname)
			return
		}
	}
	if len(rules) == 0 {
		err = ErrNoThumbnailerFound
	}
	return
}
