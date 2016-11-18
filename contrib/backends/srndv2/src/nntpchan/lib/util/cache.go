package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
)

func GetThreadHashHTML(file string) (thread string) {
	exp := regexp.MustCompilePOSIX(`thread-([0-9a-f]+)\.html`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 2 {
		return ""
	}
	thread = matches[1]
	return
}

func GetGroupAndPageHTML(file string) (board string, page int) {
	exp := regexp.MustCompilePOSIX(`(.*)-([0-9]+)\.html`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 3 {
		return "", -1
	}
	var err error
	board = matches[1]
	tmp := matches[2]
	page, err = strconv.Atoi(tmp)
	if err != nil {
		page = -1
	}
	return
}

func GetGroupForCatalogHTML(file string) (group string) {
	exp := regexp.MustCompilePOSIX(`catalog-(.+)\.html`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 2 {
		return ""
	}
	group = matches[1]
	return
}

func GetFilenameForBoardPage(webroot_dir, boardname string, pageno int, json bool) string {
	var ext string
	if json {
		ext = "json"
	} else {
		ext = "html"
	}
	fname := fmt.Sprintf("%s-%d.%s", boardname, pageno, ext)
	return filepath.Join(webroot_dir, fname)
}

func GetFilenameForThread(webroot_dir, root_post_id string, json bool) string {
	var ext string
	if json {
		ext = "json"
	} else {
		ext = "html"
	}
	fname := fmt.Sprintf("thread-%s.%s", HashMessageID(root_post_id), ext)
	return filepath.Join(webroot_dir, fname)
}

func GetFilenameForCatalog(webroot_dir, boardname string) string {
	fname := fmt.Sprintf("catalog-%s.html", boardname)
	return filepath.Join(webroot_dir, fname)
}

func GetFilenameForIndex(webroot_dir string) string {
	return filepath.Join(webroot_dir, "index.html")
}

func GetFilenameForBoards(webroot_dir string) string {
	return filepath.Join(webroot_dir, "boards.html")
}

func GetFilenameForHistory(webroot_dir string) string {
	return filepath.Join(webroot_dir, "history.html")
}

func GetFilenameForUkko(webroot_dir string) string {
	return filepath.Join(webroot_dir, "ukko.html")
}
