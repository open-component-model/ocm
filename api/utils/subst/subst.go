package subst

import (
	"bytes"
	"container/list"
	"regexp"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	mlog "github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"go.yaml.in/yaml/v4"
	glog "gopkg.in/op/go-logging.v1"

	"ocm.software/ocm/api/utils"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/runtime"
)

type SubstitutionTarget interface {
	SubstituteByData(path string, value []byte) error
	SubstituteByValue(path string, value interface{}) error

	Content() ([]byte, error)
}

func ParseFile(file string, fss ...vfs.FileSystem) (SubstitutionTarget, error) {
	fs := utils.FileSystem(fss...)

	data, err := utils.ReadFile(file, fs)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file %q", file)
	}
	s, err := Parse(data)
	if err != nil {
		return nil, errors.Wrapf(err, "file %q", file)
	}
	return s, nil
}

func Parse(data []byte) (SubstitutionTarget, error) {
	sync.OnceFunc(func() {
		var lvl glog.Level
		switch ocmlog.Context().GetDefaultLevel() {
		case mlog.None:
			fallthrough
		case mlog.ErrorLevel:
			lvl = glog.ERROR
		case mlog.WarnLevel:
			lvl = glog.WARNING
		case mlog.InfoLevel:
			lvl = glog.INFO
		case mlog.DebugLevel:
			fallthrough
		case mlog.TraceLevel:
			lvl = glog.DEBUG
		}
		glog.SetLevel(lvl, "yq-lib")
	})()

	var (
		err error
		fi  fileinfo
	)

	fi.json = true
	rdr := bytes.NewBuffer(data)
	jsnDcdr := yqlib.NewJSONDecoder()

	if err := jsnDcdr.Init(rdr); err != nil {
		return nil, err
	}

	if fi.content, err = jsnDcdr.Decode(); err != nil {
		fi.json = false
		ymlPrfs := yqlib.NewDefaultYamlPreferences()
		ymlDcdr := yqlib.NewYamlDecoder(ymlPrfs)
		rdr = bytes.NewBuffer(data)
		if err := ymlDcdr.Init(rdr); err != nil {
			return nil, err
		}
		if fi.content, err = ymlDcdr.Decode(); err != nil {
			return nil, err
		}
	}

	fi.content.SetDocument(0)
	fi.content.SetFilename("substitution-target")
	fi.content.SetFileIndex(0)

	return &fi, nil
}

type fileinfo struct {
	content *yqlib.CandidateNode
	json    bool
}

func (f *fileinfo) Content() ([]byte, error) {
	var enc yqlib.Encoder
	if f.json {
		prfs := yqlib.NewDefaultJsonPreferences()
		prfs.ColorsEnabled = false
		enc = yqlib.NewJSONEncoder(prfs)
	} else {
		prfs := yqlib.NewDefaultYamlPreferences()
		enc = yqlib.NewYamlEncoder(prfs)
	}

	buf := bytes.NewBuffer([]byte{})
	pw := yqlib.NewSinglePrinterWriter(buf)
	p := yqlib.NewPrinter(enc, pw)
	inptLst := list.New()
	inptLst.PushBack(f.content)

	if err := p.PrintResults(inptLst); err == nil {
		return buf.Bytes(), nil
	} else {
		return nil, err
	}
}

var sniffJson = regexp.MustCompile(`^\s*(\{|\[|")`)

func (f *fileinfo) SubstituteByData(path string, value []byte) error {
	var err error

	if !f.json && sniffJson.Match(value) {
		// yaml is generally a superset of json so we could just insert the json value
		// into a yaml file and have a valid yaml.
		// However having a yaml file that looks like a mix of yaml and json is off putting.
		// So if the value looks like json and the target file is yaml we will first
		// attempt to re-enode the value as yaml before inserting into the target document.
		// However... we don't want to perform re-encoding for everything because if the
		// value is actually yaml with some snippets in json style for readability
		// purposes we don't want to unnecessarily lose that styling.  Hence the initial
		// sniff test for json instead of always re-encoding.
		var valueData interface{}
		if err = runtime.DefaultJSONEncoding.Unmarshal(value, &valueData); err == nil {
			if value, err = runtime.DefaultYAMLEncoding.Marshal(valueData); err != nil {
				return err
			}
		}
	}

	m := &yaml.Node{}
	if err = yaml.Unmarshal(value, m); err != nil {
		return err
	}

	n := &yqlib.CandidateNode{}
	n.SetDocument(0)
	n.SetFilename("value")
	n.SetFileIndex(0)

	if err = n.UnmarshalYAML(m.Content[0], map[string]*yqlib.CandidateNode{}); err != nil {
		return err
	}

	return f.substituteByValue(path, n)
}

func (f *fileinfo) SubstituteByValue(path string, value interface{}) error {
	var mrshl func(interface{}) ([]byte, error)

	if f.json {
		mrshl = runtime.DefaultJSONEncoding.Marshal
	} else {
		mrshl = runtime.DefaultYAMLEncoding.Marshal
	}

	if bval, err := mrshl(value); err != nil {
		return err
	} else {
		return f.SubstituteByData(path, bval)
	}
}

func (f *fileinfo) substituteByValue(path string, value *yqlib.CandidateNode) error {
	inptLst := list.New()
	inptLst.PushBack(f.content)

	vlLst := list.New()
	vlLst.PushBack(value)

	ctxt := yqlib.Context{MatchingNodes: inptLst}
	ctxt.SetVariable("newValue", vlLst)

	yqlib.InitExpressionParser()
	expr := "." + path + " |= $newValue"

	expressionNode, err := yqlib.ExpressionParser.ParseExpression(expr)
	if err != nil {
		return err
	}

	ngvtr := yqlib.NewDataTreeNavigator()
	_, err = ngvtr.GetMatchingNodes(ctxt, expressionNode)
	return err
}
