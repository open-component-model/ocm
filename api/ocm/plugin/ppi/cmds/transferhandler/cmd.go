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

The question arguments are passed as jdon structure on *stdin*, 
the result has to be returned as JSON document
on *stdout*.

There are several questions a handler can answer:
- <code>enforcetransport</code>: This action answers the question, whether
  a component version shall be transported as if it is not yet present
  in the target repository. The argument is a ComponentVersionQuestion.

There are several types of questions:
- <code>ComponentVersionQuestion</code>: this type of question refers to
  a complete component version. The given argument has the following fields:
  - <code>source</code>: 
  	- <code>component</code> the component name.
    - <code>version</code> the component version name.
    - <code>provider</code> the provider struct from the component descriptor
      restricted to the label entries selected by the transfer handler descriptor.
    - <code>labels</code>the labels of the component version restricted to the
      label entries selected by the transfer handler descriptor.
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
	Handler   string
	Question  string
	Arguments interface{}
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
			data, _ = json.Marshal(result)
			cmd.Printf("%s\n", string(data))
			return nil
		}
	}
	return fmt.Errorf("question %q not configured for transfer handler %q", opts.Question, opts.Handler)

}
