package artifactset

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/utils/accessio"
)

type Options struct {
	accessio.StandardOptions

	FormatVersion string `json:"formatVersion,omitempty"`
	OCIBlobLayout bool   `json:"ociBlobLayout,omitempty"`
}

func NewOptions(olist ...accessio.Option) (*Options, error) {
	opts := &Options{}
	err := accessio.ApplyOptions(opts, olist...)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

type FormatVersionOption interface {
	SetFormatVersion(string)
	GetFormatVersion() string
}

func GetFormatVersion(opts accessio.Options) string {
	if o, ok := opts.(FormatVersionOption); ok {
		return o.GetFormatVersion()
	}
	return ""
}

var _ FormatVersionOption = (*Options)(nil)

func (o *Options) SetFormatVersion(s string) {
	o.FormatVersion = s
}

func (o *Options) GetFormatVersion() string {
	return o.FormatVersion
}

func (o *Options) ApplyOption(opts accessio.Options) error {
	err := o.StandardOptions.ApplyOption(opts)
	if err != nil {
		return err
	}
	if o.FormatVersion != "" {
		if s, ok := opts.(FormatVersionOption); ok {
			s.SetFormatVersion(o.FormatVersion)
		} else {
			return errors.ErrNotSupported("format version option")
		}
	}
	if o.OCIBlobLayout {
		if t, ok := opts.(*Options); ok {
			t.OCIBlobLayout = o.OCIBlobLayout
		}
	}
	return nil
}

type optFmt struct {
	format string
}

type optOCIBlobLayout bool

var _ accessio.Option = optOCIBlobLayout(false)

// OCIBlobLayout returns an option to enable OCI Image Layout blob paths.
func OCIBlobLayout(enabled bool) accessio.Option {
	return optOCIBlobLayout(enabled)
}

func (o optOCIBlobLayout) ApplyOption(opts accessio.Options) error {
	if t, ok := opts.(*Options); ok {
		t.OCIBlobLayout = bool(o)
	}
	return nil
}

var _ accessio.Option = (*optFmt)(nil)

func StructureFormat(fmt string) accessio.Option {
	return &optFmt{fmt}
}

func (o *optFmt) ApplyOption(opts accessio.Options) error {
	if s, ok := opts.(FormatVersionOption); ok {
		s.SetFormatVersion(o.format)
		return nil
	}
	return errors.ErrNotSupported("format version option")
}
