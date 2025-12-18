package yqlib

import (
	"sync"

	mlog "github.com/mandelsoft/logging"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	glog "gopkg.in/op/go-logging.v1"

	ocmlog "ocm.software/ocm/api/utils/logging"
)

var yqinit = sync.OnceFunc(func() {
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

	yqlib.InitExpressionParser()
})

func InitYq() {
	// 2025-11-19 d :
	// Prevent https://github.com/open-component-model/ocm-project/issues/773
	// Race condition with the setting of the log level 'yq-lib'
	yqinit()
}
