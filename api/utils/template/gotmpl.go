package template

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func init() {
	Register("go", func(_ vfs.FileSystem, _ TemplaterOptions) Templater { return NewGo() }, `go templating supports complex values.
<pre>
  key:
    subkey: "abc {{.MY_VAL}}"
</pre>
`)
}

type GoTemplater struct{}

func NewGo() Templater {
	return &GoTemplater{}
}

func (g GoTemplater) Process(data string, values Values) (string, error) {
	t, err := template.New("resourcespec").Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(data)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, values)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
