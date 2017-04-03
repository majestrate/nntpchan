package srnd

import (
	"github.com/majestrate/configparser"
	"golang.org/x/text/language"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type i18n struct {
	locale language.Tag
	// loaded translations
	translations map[string]string
	// loaded formats
	formats map[string]string
	// root directory for translations
	translation_dir string
}

var i18nProvider *i18n = nil

//Read all .ini files in dir, where the filenames are BCP 47 tags
//Use the language matcher to get the best match for the locale preference
func InitI18n(locale, dir string) {
	pref := language.Make(locale) // falls back to en-US on parse error

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	serverLangs := make([]language.Tag, 1)
	serverLangs[0] = language.AmericanEnglish // en-US fallback
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ini" {
			name := strings.TrimSuffix(file.Name(), ".ini")
			tag, err := language.Parse(name)
			if err == nil {
				serverLangs = append(serverLangs, tag)
			}
		}
	}
	matcher := language.NewMatcher(serverLangs)
	tag, _, _ := matcher.Match(pref)

	fname := filepath.Join(dir, tag.String()+".ini")
	conf, err := configparser.Read(fname)
	if err != nil {
		log.Fatal("cannot read translation file for", tag.String(), err)
	}

	formats, err := conf.Section("formats")
	if err != nil {
		log.Fatal("Cannot read formats sections in translations for", tag.String(), err)
	}
	translations, err := conf.Section("strings")
	if err != nil {
		log.Fatal("Cannot read strings sections in translations for", tag.String(), err)
	}

	i18nProvider = &i18n{
		translation_dir: dir,
		formats:         formats.Options(),
		translations:    translations.Options(),
		locale:          tag,
	}
}

func (self *i18n) Translate(key string) string {
	return self.translations[key]
}

func (self *i18n) Format(key string) string {
	return self.formats[key]
}

//this signature seems to be expected by mustache
func (self *i18n) Translations() (map[string]string, error) {
	return self.translations, nil
}

func (self *i18n) Formats() (map[string]string, error) {
	return self.formats, nil
}
