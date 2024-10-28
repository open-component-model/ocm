package transferhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/plugin/ppi"
)

const Name = "transferhandler"

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}
	cmd := &cobra.Command{
		Use:   Name + " <name> <question>",
		Short: "decide on a question related to a component version transport",
		Long: `
The task of this command is to decide on questions related to the transport
of component versions, their resources and sources.

This command is only used for actions for which the transfer handler descriptor
enables the particular question.

The question arguments are passed as JSON structure on *stdin*, 
the result has to be returned as JSON document
on *stdout*.

There are several questions a handler can answer:
- <code>transferversion</code>: This action answers the question, whether
  a component version shall be transported at all and how it should be
  transported. The argument is a <code>ComponentVersionQuestion</code>
  and the result decision is extended by an optional transport context
  consisting, of a new transport handler description and transport options.
- <code>enforcetransport</code>: This action answers the question, whether
  a component version shall be transported as if it is not yet present
  in the target repository. The argument is a <code>ComponentVersionQuestion</code>.
- <code>updateversion</code>: Update non-signature relevant information.
  The argument is a <code>ComponentVersionQuestion</code>.
- <code>overwriteversion</code>: Override signature-relevant information.
  The argument is a <code>ComponentVersionQuestion</code>.
- <code>transferresource</code>: Transport resource as value. The argument
  is an <code>ArtifactQuestion</code>.
- <code>transfersource</code>: Transport source as value. The argument
  is an <code>ArtifactQuestion</code>.

For detailed types see <code>ocm.software/ocm/api/ocm/plugin/ppi/questions.go</code>.
`,
		Args: cobra.ExactArgs(2),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Command(p, cmd, &opts)
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}

type Options struct {
	// Handler is the handler name as provided by the plugin.
	Handler string
	// Question cis the name of the question
	Question string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	o.Handler = args[0]
	o.Question = args[1]
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	t := ppi.TransferHandlerQuestions[opts.Question]
	if t == nil {
		return fmt.Errorf("unknown question %q", opts.Question)
	}
	h := p.GetTransferHandler(opts.Handler)
	if h == nil {
		return fmt.Errorf("unknown transfer handler %q", opts.Handler)
	}
	for _, q := range h.GetQuestions() {
		if q.GetQuestion() == opts.Question {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}

			args := reflect.New(t).Interface()
			err = json.Unmarshal(data, args)
			if err != nil {
				return err
			}

			var result ppi.DecisionRequestResult
			result.Decision, err = q.DecideOn(p, args)
			if err != nil {
				result.Error = err.Error()
			}
			data, err = json.Marshal(result)
			if err != nil {
				return err
			}
			cmd.Printf("%s\n", string(data))
			return nil
		}
	}
	return fmt.Errorf("question %q not configured for transfer handler %q", opts.Question, opts.Handler)
}
