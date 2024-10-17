## plugin transferhandler &mdash; Decide On A Question Related To A Component Version Transport

### Synopsis

```bash
plugin transferhandler <name> <question> [<options>]
```

### Options

```text
  -h, --help   help for transferhandler
```

### Description

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

### SEE ALSO

#### Parents

* [plugin](plugin.md)	 &mdash; OCM Plugin

