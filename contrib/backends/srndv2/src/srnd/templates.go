//
// templates.go
// template model interfaces
//
package srnd

import (
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type TemplateDriver interface {
	RenderString(template string, obj interface{}) (string, error)
	Render(template string, obj interface{}, w io.Writer) error
	Ext() string
}

// get our default template dir
func defaultTemplateDir() string {
	p, _ := filepath.Abs(filepath.Join("contrib", "templates", "default"))
	return p
}

func templateDriverForFile(fname string) TemplateDriver {
	switch strings.ToLower(filepath.Ext(fname)) {
	case ".mustache":
		return new(mustacheDriver)
	case ".tmpl":
		return new(stdTemplateDriver)
	default:
		return nil
	}
}

func templateDriverFromDir(dir string) TemplateDriver {
	files, err := ioutil.ReadDir(dir)
	if err == nil && len(files) > 0 {
		for idx := range files {
			if !files[idx].IsDir() {
				return templateDriverForFile(files[idx].Name())
			}
		}
	}
	log.Println("no template found in ", dir, " ", err)
	return nil
}

func newTemplateEngine(dir string) *templateEngine {
	return &templateEngine{
		templates:    make(map[string]string),
		template_dir: dir,
		driver:       templateDriverFromDir(dir),
	}
}

var template = newTemplateEngine(defaultTemplateDir())

func ReloadTemplates() {
	log.Println("reload templates")
	template.reloadAllTemplates()
}
