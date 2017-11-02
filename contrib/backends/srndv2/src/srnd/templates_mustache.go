//
// templates.go
// template model interfaces
//
package srnd

import (
	"github.com/cbroglie/mustache"
	"io"
)

type mustacheDriver struct {
}

func (d *mustacheDriver) RenderString(templ string, obj interface{}) (string, error) {
	return mustache.Render(templ, obj)
}

func (d *mustacheDriver) Render(templ string, obj interface{}, w io.Writer) error {
	s, err := d.RenderString(templ, obj)
	if err == nil {
		_, err = io.WriteString(w, s)
	}
	return err
}

func (d *mustacheDriver) Ext() string {
	return ".mustache"
}
