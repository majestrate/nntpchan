// markup.go
// memeposting markup parser
//
package srnd

import (
	"github.com/mvdan/xurls"
	"html"
	"regexp"
	"strings"
)

// copypasted from https://stackoverflow.com/questions/161738/what-is-the-best-regular-expression-to-check-if-a-string-is-a-valid-url
// var re_external_link = regexp.MustCompile(`((?:(?:https?|ftp):\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,}))\.?)(?::\d{2,5})?(?:[/?#]\S*)?)`);
var re_external_link = xurls.Strict
var re_backlink = regexp.MustCompile(`>> ?([0-9a-f]+)`)
var re_boardlink = regexp.MustCompile(`>>> ?/([0-9a-zA-Z\.]+)/`)
var re_nntpboardlink = regexp.MustCompile(`news:([0-9a-zA-Z\.]+)`)

// find all backlinks in string
func findBacklinks(msg string) (cites []string) {
	re := re_backlink.Copy()
	cmap := make(map[string]string)
	for _, cite := range re.FindAllString(msg, -1) {
		cmap[cite] = cite
	}
	for _, c := range cmap {
		cites = append(cites, c)
	}
	return
}

// parse backlink
func backlink(word, prefix string) (markup string) {
	re := re_backlink.Copy()
	link := re.FindString(word)
	if len(link) > 2 {
		link = strings.Trim(link[2:], " ")
		if len(link) > 2 {
			url := template.findLink(prefix, link)
			if len(url) == 0 {
				return "<span class='memearrows'>&gt;&gt;" + link + "</span>"
			}
			// backlink exists
			parts := strings.Split(url, "#")
			longhash := ""
			if len(parts) > 1 {
				longhash = parts[1]
			}
			return `<a class='backlink' backlinkhash="` + longhash + `" href="` + url + `">&gt;&gt;` + link + "</a>"
		} else {
			return escapeline(word)
		}
	}
	return escapeline(word)
}

func boardlink(word, prefix string, r *regexp.Regexp) (markup string) {
	re := r.Copy()
	l := re.FindStringSubmatch(word)
	if len(l[1]) > 2 {
		link := strings.ToLower(l[1])
		markup = `<a class="boardlink" href="` + prefix + "b/" + link + `">&gt;&gt;&gt;/` + link + `/</a>`
		return
	}
	markup = escapeline(word)
	return
}

func escapeline(line string) (markup string) {
	markup = html.EscapeString(line)
	return
}

func formatline(line, prefix string) (markup string) {
	if len(line) > 0 {
		line_nospace := strings.Trim(line, " ")
		if strings.HasPrefix(line_nospace, ">") && !strings.HasPrefix(line_nospace, ">>") {
			// le ebin meme arrows
			markup += "<span class='memearrows'>"
			markup += escapeline(line)
			markup += "</span>"
		} else {
			// regular line
			// for each word
			for _, word := range strings.Split(line, " ") {
				if re_boardlink.MatchString(word) {
					markup += boardlink(word, prefix, re_boardlink)
				} else if re_nntpboardlink.MatchString(word) {
					markup += boardlink(word, prefix, re_nntpboardlink)
				} else if re_backlink.MatchString(word) {
					markup += backlink(word, prefix)
				} else {
					// linkify as needed
					word = escapeline(word)
					markup += re_external_link.ReplaceAllString(word, `<a href="$1">$1</a>`)
				}
				markup += " "
			}
		}
	}
	return
}

func MEMEPosting(src, prefix string) (markup string) {
	for _, line := range strings.Split(src, "\n") {
		line = strings.Trim(line, "\r")
		markup += formatline(line, prefix) + "\n"
	}
	return extraMemePosting(markup, prefix)
}
