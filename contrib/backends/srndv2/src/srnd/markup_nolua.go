// +build !lua

package srnd

func extraMemePosting(src, prefix string) string {
	return src
}

func SetMarkupScriptFile(fname string) error {
	// does nothing for non lua
	return nil
}
