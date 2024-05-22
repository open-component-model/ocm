package subst

import (
	"bytes"
	"container/list"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	mlog "github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	glog "gopkg.in/op/go-logging.v1"
	"gopkg.in/yaml.v3"

	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
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

var yqLibLogInit sync.Once

func Parse(data []byte) (SubstitutionTarget, error) {
	yqLibLogInit.Do(func() {
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
	})

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

func (f *fileinfo) SubstituteByData(path string, value []byte) error {
	var node interface{}
	err := runtime.DefaultYAMLEncoding.Unmarshal(value, &node)
	if err != nil {
		return err
	}
	if f.json {
		value, err = runtime.DefaultJSONEncoding.Marshal(node)
	} else {
		value, err = runtime.DefaultYAMLEncoding.Marshal(node)
	}
	if err != nil {
		return err
	}
	m := &yaml.Node{}
	err = yaml.Unmarshal(value, m)
	if err != nil {
		return err
	}

	if !f.json {
		var replaceFlowStyle func(*yaml.Node)
		replaceFlowStyle = func(nd *yaml.Node) {
			if nd.Style == yaml.FlowStyle {
				nd.Style = yaml.LiteralStyle
			}
			for _, chld := range nd.Content {
				replaceFlowStyle(chld)
			}
		}
		replaceFlowStyle(m)
	}

	nd := &yqlib.CandidateNode{}
	nd.SetDocument(0)
	nd.SetFilename("value")
	nd.SetFileIndex(0)

	if err = nd.UnmarshalYAML(m.Content[0], map[string]*yqlib.CandidateNode{}); err != nil {
		return err
	}
	return f.substituteByValue(path, nd)
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

	nd, err := yqlib.ExpressionParser.ParseExpression(expr)
	if err != nil {
		return err
	}

	ngvtr := yqlib.NewDataTreeNavigator()
	_, err = ngvtr.GetMatchingNodes(ctxt, nd)
	return err
}
