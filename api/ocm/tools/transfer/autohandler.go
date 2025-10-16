package transfer

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
)

func init() {
	SetDefaultTransferHandlerFactory(func() transferhandler.TransferHandler { h, _ := standard.New(); return h })
}

var (
	lock           sync.Mutex
	defaultHandler func() transferhandler.TransferHandler
)

func SetDefaultTransferHandlerFactory(f func() transferhandler.TransferHandler) {
	lock.Lock()
	defer lock.Unlock()
	defaultHandler = f
}

func NewDefaultTransferHandler() transferhandler.TransferHandler {
	lock.Lock()
	defer lock.Unlock()
	return defaultHandler()
}

// NewTransferHandler creates a transfer handler for the given set of transfer
// options. If there is no handler type supporting all the given options
// an ErrNotSupported error is returned.
// If no handler can be found according to the origin handlers of given options
// (an option always belongs to a dedicated option set of a dedicated handler
// but may be supported by other option sets, also) the options sets of all
// registered handler types with checked by trying the most significant
// handlers, first, to find the best matching handler for the given option set.
func NewTransferHandler(list ...transferhandler.TransferOption) (h transferhandler.TransferHandler, ferr error) {
	var opts transferhandler.TransferHandlerOptions

	created := -1
outer:
	for {
		ferr = nil
		for i := 0; i < len(list); i++ {
			o := list[i]
			if o == nil {
				continue
			}
			c, ok := o.(transferhandler.TransferOptionsCreator)
			if !ok {
				// option nor used for transfer handlers, just ignore it here.
				continue
			}
			if opts == nil {
				created = i
				opts = c.NewOptions()
			}
			if err := o.ApplyTransferOption(opts); err != nil {
				if errors.IsErrNotSupportedKind(err, transferhandler.KIND_TRANSFEROPTION) {
					if i <= created {
						// give a second chance to later options
						i = created
						ferr = err
						opts = nil
					} else {
						// try next options implementation
						created = i
						opts = c.NewOptions()
						continue outer
					}
				} else {
					return nil, err
				}
			}
		}
		if opts == nil {
			if ferr != nil {
				// third change: try registered handlers in order
				// most specific first
			next:
				for _, h := range transferhandler.OrderedTransferOptionCreators() {
					opts = h.NewOptions()
					for _, o := range list {
						if o == nil {
							continue
						}
						if err := o.ApplyTransferOption(opts); err != nil {
							if errors.IsErrNotSupportedKind(err, transferhandler.KIND_TRANSFEROPTION) {
								opts = nil
								continue next
							}
							return nil, err
						}
					}
				}
				if opts == nil {
					return nil, ferr
				}
				return opts.NewTransferHandler()
			}
			return NewDefaultTransferHandler(), nil
		}
		if ferr != nil {
			// continue second chance from beginning
			continue
		}
		// all options are accepted now, so just return the appropriate handler.
		return opts.NewTransferHandler()
	}
}
