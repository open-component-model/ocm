package names

var (
	// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	ComponentArchive       = []string{"componentarchive", "comparch", "ca"}
	CommonTransportArchive = []string{"commontransportarchive", "ctf"}
	Components             = []string{"componentversions", "componentversion", "cv", "components", "component", "comps", "comp", "c"}
	CLI                    = []string{"cli", "ocmcli", "ocm-cli"}
	Configuration          = []string{"configuration", "config", "cfg"}
	ResourceConfig         = []string{"resource-configuration", "resourceconfig", "rsccfg", "rcfg"}
	SourceConfig           = []string{"source-configuration", "sourceconfig", "srccfg", "scfg"}
	Resources              = []string{"resources", "resource", "res", "r"}
	Sources                = []string{"sources", "source", "src", "s"}
	References             = []string{"references", "reference", "refs"}
	Versions               = []string{"versions", "vers", "v"}
	Plugins                = []string{"plugins", "plugin", "p"}
	Action                 = []string{"action"}
	RoutingSlips           = []string{"routingslips", "routingslip", "rs"}
	PubSub                 = []string{"pubsub", "ps"}
	Verified               = []string{"verified"}
)

var Aliases = map[string][]string{}

func init() {
	add(
		ComponentArchive,
		CommonTransportArchive,
		Components,
		CLI,
		Configuration,
		ResourceConfig,
		SourceConfig,
		Resources,
		Sources,
		References,
		Versions,
		Plugins,
		Action,
		RoutingSlips,
		PubSub,
		Verified,
	)
}

func add(aliases ...[]string) {
	for _, a := range aliases {
		Aliases[a[0]] = a
	}
}
