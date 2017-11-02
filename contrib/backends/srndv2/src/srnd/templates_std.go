package srnd

import (
	"bytes"
	stdtemplate "html/template"
	"io"
)

type stdTemplateDriver struct {
}

func (d *stdTemplateDriver) Render(templ string, obj interface{}, w io.Writer) error {
	return stdtemplate.Must(stdtemplate.New("").Parse(templ)).Execute(w, obj)
}

func (d *stdTemplateDriver) RenderString(templ string, obj interface{}) (string, error) {
	buff := new(bytes.Buffer)
	err := d.Render(templ, obj, buff)
	if err == nil {
		return buff.String(), nil
	}
	return "", err
}

func (d *stdTemplateDriver) Ext() string {
	return ".tmpl"
}
