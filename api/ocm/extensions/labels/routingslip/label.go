package routingslip

import (
	"sort"

	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/maplistmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/simplemapmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
	"ocm.software/ocm/api/utils"
)

const NAME = "routing-slips"

type LabelValue map[string]HistoryEntries

var spec = utils.Must(hpi.NewSpecification(
	simplemapmerge.ALGORITHM,
	simplemapmerge.NewConfig(
		"",
		utils.Must(hpi.NewSpecification(
			maplistmerge.ALGORITHM,
			maplistmerge.NewConfig("digest", maplistmerge.MODE_INBOUND),
		)),
	)),
)

func init() {
	hpi.Assign(hpi.LabelHint(NAME), spec)
}

func (l LabelValue) Has(name string) bool {
	return l[name] != nil
}

func (l LabelValue) Get(name string) (*RoutingSlip, error) {
	return NewRoutingSlip(name, l)
}

func (l LabelValue) Query(name string) (*RoutingSlip, error) {
	a := l[name]
	if a == nil {
		return nil, nil
	}
	return l.Get(name)
}

func (l LabelValue) Leaves() []Link {
	var links []Link

	for k := range l {
		s, err := l.Get(k)
		if err == nil {
			for _, d := range s.Leaves() {
				links = append(links, Link{
					Name:   k,
					Digest: d,
				})
			}
		}
	}
	sort.Slice(links, func(i, j int) bool { return links[i].Compare(links[j]) < 0 })
	return links
}

func (l LabelValue) Set(slip *RoutingSlip) {
	l[slip.name] = slip.entries
}

func AddEntry(cv cpi.ComponentVersionAccess, name string, algo string, e Entry, links []Link, parent ...digest.Digest) (*HistoryEntry, error) {
	var label LabelValue
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return nil, err
	}
	if label == nil {
		label = LabelValue{}
	}
	slip, err := label.Get(name)
	if err != nil {
		return nil, err
	}
	entry, err := slip.Add(cv.GetContext(), name, algo, e, links, parent...)
	if err != nil {
		return nil, err
	}
	label.Set(slip)

	err = Set(cv, label)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func Get(cv cpi.ComponentVersionAccess) (LabelValue, error) {
	var label LabelValue
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func Set(cv cpi.ComponentVersionAccess, label LabelValue) error {
	return cv.GetDescriptor().Labels.SetValue(NAME, label)
}
