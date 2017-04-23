//
// policy.go
//
package srnd

import (
	"regexp"
	"strings"
)

type FeedPolicy struct {
	rules map[string]string
}

// do we allow this newsgroup?
func (self *FeedPolicy) AllowsNewsgroup(newsgroup string) (result bool) {
	var k, v string
	var allows int
	for k, v = range self.rules {
		v = strings.Trim(v, " ")
		match, err := regexp.MatchString(k, newsgroup)
		if err == nil {
			if match {
				if v == "1" {
					allows++
				} else if v == "0" {
					return false
				}
			}
		}
	}

	result = allows > 0

	return
}
