//
// policy.go
//
package srnd

import (
	"log"
	"regexp"
)

type FeedPolicy struct {
	rules map[string]string
}

// do we allow this newsgroup?
func (self *FeedPolicy) AllowsNewsgroup(newsgroup string) (result bool) {
	var k, v string
	for k, v = range self.rules {
		match, err := regexp.MatchString(k, newsgroup)
		if err != nil {
			log.Fatal(err)
		}
		if match {
			if v == "0" {
				return false
			} else if v == "1" {
				result = true
			}
		}
	}
	return result
}
