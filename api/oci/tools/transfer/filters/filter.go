package filters

import (
	"encoding/json"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils"
)

type Filter interface {
	Accept(art cpi.ArtifactAccess, platform *artdesc.Platform) bool
}

type and struct {
	filters []Filter
}

func And(filter ...Filter) Filter {
	cnt := 0
	for _, f := range filter {
		if f != nil {
			cnt++
		}
	}
	if cnt == 0 {
		return nil
	}
	return &and{filter}
}

func (c *and) Accept(art cpi.ArtifactAccess, platform *artdesc.Platform) bool {
	for _, f := range c.filters {
		if f != nil && !f.Accept(art, platform) {
			return false
		}
	}
	return len(c.filters) > 0
}

type or struct {
	filters []Filter
}

func Or(filter ...Filter) Filter {
	cnt := 0
	for _, f := range filter {
		if f != nil {
			cnt++
		}
	}
	if cnt == 0 {
		return nil
	}
	return &or{filter}
}

func (c *or) Accept(art cpi.ArtifactAccess, platform *artdesc.Platform) bool {
	for _, f := range c.filters {
		if f != nil && f.Accept(art, platform) {
			return true
		}
	}
	return false
}

type not struct {
	filter Filter
}

func Not(filter Filter) Filter {
	return &not{filter}
}

func (c *not) Accept(art cpi.ArtifactAccess, platform *artdesc.Platform) bool {
	if c.filter != nil {
		return !c.filter.Accept(art, platform)
	}
	return false
}

type platform struct {
	os   string
	arch string
	excl bool // exclude artifacts without a platform
}

func Platform(os string, arch string, excl ...bool) Filter {
	return &platform{os, arch, utils.Optional(excl...)}
}

func (f *platform) Accept(art cpi.ArtifactAccess, platform *artdesc.Platform) bool {
	if art.IsIndex() {
		return false
	}

	if f.os == "" && f.arch == "" {
		return true
	}
	if platform == nil {
		cfg, err := art.ManifestAccess().GetConfigBlob()
		if err != nil {
			return false
		}
		if cfg.MimeType() != ociv1.MediaTypeImageConfig {
			return !f.excl
		}
		data, err := cfg.Get()
		if err != nil {
			return false
		}

		var im ociv1.Image
		err = json.Unmarshal(data, &im)
		if err != nil {
			return false
		}
		platform = &im.Platform
	}
	if f.os != "" && f.os != platform.OS {
		return false
	}
	if f.arch != "" && f.arch != platform.Architecture {
		return false
	}
	return true
}
