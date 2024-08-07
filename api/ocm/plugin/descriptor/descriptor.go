package descriptor

import (
	"encoding/json"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

const VERSION = "v1"

type Descriptor struct {
	Version        string `json:"version,omitempty"`
	PluginName     string `json:"pluginName"`
	PluginVersion  string `json:"pluginVersion"`
	Short          string `json:"shortDescription"`
	Long           string `json:"description"`
	ForwardLogging bool   `json:"forwardLogging"`

	Actions                  []ActionDescriptor                `json:"actions,omitempty"`
	AccessMethods            []AccessMethodDescriptor          `json:"accessMethods,omitempty"`
	Uploaders                List[UploaderDescriptor]          `json:"uploaders,omitempty"`
	Downloaders              List[DownloaderDescriptor]        `json:"downloaders,omitempty"`
	ValueMergeHandlers       List[ValueMergeHandlerDescriptor] `json:"valueMergeHandlers,omitempty"`
	LabelMergeSpecifications List[LabelMergeSpecification]     `json:"labelMergeSpecifications,omitempty"`
	ValueSets                List[ValueSetDescriptor]          `json:"valuesets,omitempty"`
	Commands                 List[CommandDescriptor]           `json:"commands,omitempty"`
	ConfigTypes              List[ConfigTypeDescriptor]        `json:"configTypes,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

func (d *Descriptor) Capabilities() []string {
	var caps []string
	if len(d.AccessMethods) > 0 {
		caps = append(caps, "Access Methods")
	}
	if len(d.Uploaders) > 0 {
		caps = append(caps, "Repository Uploaders")
	}
	if len(d.Downloaders) > 0 {
		caps = append(caps, "Resource Downloaders")
	}
	if len(d.Actions) > 0 {
		caps = append(caps, "Actions")
	}
	if len(d.ValueSets) > 0 {
		caps = append(caps, "Value Sets")
	}
	if len(d.ValueMergeHandlers) > 0 {
		caps = append(caps, "Value Merge Handlers")
	}
	if len(d.LabelMergeSpecifications) > 0 {
		caps = append(caps, "Label Merge Specs")
	}
	if len(d.Commands) > 0 {
		caps = append(caps, "CLI Commands")
	}
	if len(d.ConfigTypes) > 0 {
		caps = append(caps, "Config Types")
	}
	return caps
}

////////////////////////////////////////////////////////////////////////////////

type DownloaderKey = ArtifactContext

func NewDownloaderKey(arttype, mediatype string) DownloaderKey {
	return DownloaderKey{
		ArtifactType: arttype,
		MediaType:    mediatype,
	}
}

type DownloaderDescriptor struct {
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	Constraints      []DownloaderKey          `json:"constraints,omitempty"`
	ConfigScheme     string                   `json:"configScheme,omitempty"`
	AutoRegistration []DownloaderRegistration `json:"autoRegistration,omitempty"`
}

func (d DownloaderDescriptor) GetName() string {
	return d.Name
}

func (d DownloaderDescriptor) GetDescription() string {
	return d.Description
}

func (d DownloaderDescriptor) GetConstraints() []DownloaderKey {
	return d.Constraints
}

type DownloaderRegistration struct {
	DownloaderKey `json:",inline"`
	Priority      int `json:"priority,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

type UploaderDescriptor struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Constraints []UploaderKey `json:"constraints,omitempty"`
}

func (d UploaderDescriptor) GetName() string {
	return d.Name
}

func (d UploaderDescriptor) GetDescription() string {
	return d.Description
}

func (d UploaderDescriptor) GetConstraints() []UploaderKey {
	return d.Constraints
}

type AccessMethodDescriptor struct {
	ValueSetDefinition `json:",inline"`
}

////////////////////////////////////////////////////////////////////////////////

type ValueTypeDefinition struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description"`
	Format      string `json:"format"`
}

func (d ValueTypeDefinition) GetName() string {
	return d.Name
}

func (d ValueTypeDefinition) GetDescription() string {
	return d.Description
}

type ValueSetDescriptor struct {
	ValueSetDefinition `json:",inline"`
	Purposes           []string `json:"purposes"`
}

const PURPOSE_ROUTINGSLIP = "routingslip"

type ValueSetDefinition struct {
	ValueTypeDefinition
	CLIOptions []CLIOption `json:"options,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

type ValueMergeHandlerDescriptor struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (a ValueMergeHandlerDescriptor) GetName() string {
	return a.Name
}

func (a ValueMergeHandlerDescriptor) GetDescription() string {
	return a.Description
}

////////////////////////////////////////////////////////////////////////////////

type LabelMergeSpecification struct {
	Name                               string `json:"name"`
	Version                            string `json:"version,omitempty"`
	Description                        string `json:"description,omitempty"`
	metav1.MergeAlgorithmSpecification `json:",inline"`
}

func (a LabelMergeSpecification) GetName() string {
	if a.Version != "" {
		return a.Name + "@" + a.Version
	}
	return a.Name
}

func (a LabelMergeSpecification) GetDescription() string {
	return a.Description
}

func (a LabelMergeSpecification) GetAlgorithm() string {
	return a.Algorithm
}

func (a LabelMergeSpecification) GetConfig() json.RawMessage {
	return a.Config
}

////////////////////////////////////////////////////////////////////////////////

type ActionDescriptor struct {
	Name             string   `json:"name"`
	Versions         []string `json:"versions,omitempty"`
	Description      string   `json:"description,omitempty"`
	ConsumerType     string   `json:"consumerType,omitempty"`
	DefaultSelectors []string `json:"defaultSelectors,omitempty"`
}

func (a ActionDescriptor) GetName() string {
	return a.Name
}

func (a ActionDescriptor) GetDescription() string {
	return a.Description
}

////////////////////////////////////////////////////////////////////////////////

type CommandDescriptor struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	ObjectType        string `json:"objectName,omitempty"`
	Usage             string `json:"usage,omitempty"`
	Short             string `json:"short,omitempty"`
	Example           string `json:"example,omitempty"`
	Realm             string `json:"realm,omitempty"`
	Verb              string `json:"verb,omitempty"`
	CLIConfigRequired bool   `json:"cliconfig,omitempty"`
}

func (a CommandDescriptor) GetName() string {
	return a.Name
}

func (a CommandDescriptor) GetDescription() string {
	return a.Description
}

////////////////////////////////////////////////////////////////////////////////

type ConfigTypeDescriptor = ValueTypeDefinition

////////////////////////////////////////////////////////////////////////////////

type CLIOption struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}
