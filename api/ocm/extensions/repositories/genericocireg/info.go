package genericocireg

import (
	"encoding/json"

	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/ociutils"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

func init() {
	ociutils.RegisterInfoHandler(componentmapping.ComponentDescriptorConfigMimeType, &handler{})
}

type handler struct{}

type ComponentVersionInfo struct {
	Error       string          `json:"error,omitempty"`
	Description string          `json:"description"`
	Unparsed    string          `json:"unparsed,omitempty"`
	Descriptor  json.RawMessage `json:"descriptor,omitempty"`
}

func (h handler) Info(m cpi.ManifestAccess, config []byte) interface{} {
	info := &ComponentVersionInfo{
		Description: "component version",
	}
	acc := NewStateAccess(nil, m)
	data, err := blobaccess.BlobData(acc.Get())
	if err != nil {
		info.Error = "cannot read component descriptor: " + err.Error()
		return info
	}
	var raw interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		info.Unparsed = string(data)
		return info
	}
	info.Descriptor = data
	return info
}

func (h handler) Description(pr common.Printer, m cpi.ManifestAccess, config []byte) {
	pr.Printf("component version:\n")
	acc := NewStateAccess(nil, m)
	data, err := blobaccess.BlobData(acc.Get())
	if err != nil {
		pr.Printf("  cannot read component descriptor: %s\n", err.Error())
		return
	}
	pr.Printf("  descriptor:\n")
	var raw interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &raw)
	if err != nil {
		pr.Printf("    data: %s\n", string(data))
		pr.Printf("  cannot get unmarshal component descriptor: %s\n", err.Error())
		return
	}

	form, err := json.MarshalIndent(raw, "  ", "    ")
	if err != nil {
		pr.Printf("    data: %s\n", string(data))
		pr.Printf("  cannot get marshal component descriptor: %s\n", err.Error())
		return
	}
	pr.Printf("%s\n", string(form))
}
