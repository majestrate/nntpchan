package srnd

import (
	"errors"
	"github.com/majestrate/configparser"
	"golang.org/x/text/language"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type I18N struct {
	locale language.Tag
	// loaded translations
	Translations map[string]string
	// loaded formats
	Formats map[string]string
	// root directory for translations
	translation_dir string
	// name of locale
	Name string
}

var I18nProvider *I18N = nil
var ErrNoLang = errors.New("no such language")

//Read all .ini files in dir, where the filenames are BCP 47 tags
//Use the language matcher to get the best match for the locale preference
func InitI18n(locale, dir string) {
	var err error
	I18nProvider, err = NewI18n(locale, dir)
	if err != nil {
		log.Fatal(err)
	}
}

func NewI18n(locale, dir string) (*I18N, error) {
	log.Println("get locale", locale)
	pref := language.Make(locale) // falls back to en-US on parse error
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	found := false
	serverLangs := make([]language.Tag, 1)
	serverLangs[0] = language.AmericanEnglish // en-US fallback
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ini" {
			name := strings.TrimSuffix(file.Name(), ".ini")
			tag, err := language.Parse(name)
			if err == nil {
				serverLangs = append(serverLangs, tag)
				found = true
			}
		}
	}
	if !found {
		return nil, ErrNoLang
	}

	matcher := language.NewMatcher(serverLangs)
	tag, _, _ := matcher.Match(pref)

	fname := filepath.Join(dir, tag.String()+".ini")
	conf, err := configparser.Read(fname)
	if err != nil {
		return nil, err
	}

	formats, err := conf.Section("formats")
	if err != nil {
		return nil, err
	}
	translations, err := conf.Section("strings")
	if err != nil {
		return nil, err
	}

	return &I18N{
		Name:            locale,
		translation_dir: dir,
		Formats:         formats.Options(),
		Translations:    translations.Options(),
		locale:          tag,
	}, nil
}

func (self *I18N) Translate(key string) string {
	return self.Translations[key]
}

func (self *I18N) Format(key string) string {
	return self.Formats[key]
}

//this signature seems to be expected by mustache
//func (self *I18N) Translations() (map[string]string, error) {
//	return self._translations, nil
//}

//func (self *I18N) Formats() (map[string]string, error) {
//	return self.formats, nil
//}
