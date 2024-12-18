package transfer

import (
	"fmt"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/spiff"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/transferhandleroption"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func ValidateHandler(o options.OptionSetProvider) error {
	t := transferhandleroption.From(o)
	if t.Path != "" {
		s := scriptoption.From(o)
		if len(s.ScriptData) > 0 {
			if t.Path != "ocm/spiff" {
				return fmt.Errorf("transfer handler %q not compatible with script option", t.Path)
			}
			if len(t.Config) > 0 {
				return fmt.Errorf("transfer handler %q with config not compatible with script option", t.Path)
			}
		}
	}
	return nil
}

func DetermineTransferHandler(ctx clictx.Context, o options.OptionSetProvider) (transferhandler.TransferHandler, error) {
	var (
		thdlr transferhandler.TransferHandler
		err   error
	)

	opts := options.FindOptions[transferhandler.TransferOption](o)
	th := transferhandleroption.From(o)
	if th.Path != "" {
		out.Outf(ctx, "using transfer handler %s\n", th.Path)
		thdlr, err = transferhandler.For(ctx).ByName(ctx, th.Path, opts...)
	} else {
		transferopts := &spiff.Options{}
		transferhandler.From(ctx.ConfigContext(), transferopts)
		transferhandler.ApplyOptions(transferopts, opts...)
		thdlr, err = spiff.New(transferopts)
	}
	if err != nil {
		return nil, err
	}
	return thdlr, nil
}
