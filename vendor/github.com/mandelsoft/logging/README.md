# Logging for Go with context-specific Log Level Settings

This package provides a wrapper around the [logr](https://github.com/go-logr/logr)
logging system supporting a rule based approach to enable log levels
for dedicated message contexts specified at the logging location.

The rule set is configured for a logging context (`Context`). It holds
information about the rule set, log level settings, a standard 
[message context](#message-contexts-and-conditions) and the configured
base logger (a `logr.Logger`). With this information it is then used to
create `Logger` objects (optionally for sub message contexts), which can
be used to issue log messages for some standard levels.
The setting of the context decide together with the message context
of a logger about its active log level.

A new logging context can be created with:

```go
    ctx := logging.New(logrLogger)
```

Any `logr.Logger` can be passed here, the level for this logger
is used as base level for the `ErrorLevel` of loggers provided
by the logging context.
If the full control should be handed over to the logging context, 
the maximum log level should be used for the sink of this logger.

If the used base level should always be 0, the base logger has to 
be set with plain mode:

```go
    ctx.SetBaseLogger(logrLogger, true)
```

Now you can add rules controlling the accepted log levels for dedicated log 
locations. First, a default log level can be set:

```go
    ctx.SetDefaultLevel(logging.InfoLevel)
```

This level restriction is used, if no rule matches a dedicated log request.

Another way to achieve the same goal is to provide a generic level rule without any
condition:

```go
    ctx.AddRule(logging.NewConditionRule(logging.InfoLevel))
```

A first rule for influencing the log level could be a realm rule.
A *Realm* represents a dedicated logical area, a good practice could be 
to use package names as realms. Realms are hierarchical consisting of
name components separated by a slash (/).

```go
    ctx.AddRule(logging.NewConditionRule(logging.DebugLevel, logging.NewRealm("github.com/mandelsoft/spiff")))
```

Alternatively `NewRealmPrefix(...)` can be used to match a complete realm hierarchy.

A realm for the actual package can be defined as local variable by using the
`Package` function:

```go
var realm = logging.Package()
```

Instead of passing `Logger`s around, now the logging `Context` is used.
It provides a method to access a logger specific for a dedicated log
request, for example, for a dedicated realm.

```go
  ctx.Logger(realm).Info("my message")
```

The provided logger offers the level specific functions, `Error`, `Warn`, `Info`, `Debug` and `Trace`.
Depending on the rule set configured for the used logging context, the level
for the given message context decides, which message to pass to the log sink of
the initial `logr.Logger`.

Like a traditional `logr.Logger`, the logging messages take a string and an
optional list a key/value arguments to describe formalized logging fields
for a structured log output.

Instead of two separate arguments for key and value, the function `KeyValue`
can be used to provide a key/value pair as single argument. This function
can be used to define standard keys for key/value pairs for dedicated usage
scenarios (see package `keyvalue`, which provide some standards for errors, ids or names).

Alternatively a traditional `logr.Logger` for the given message context can be
obtained by using the `V` method:

```go
  ctx.V(logging.InfoLevel, realm).Info("my message")
```

Those loggers do NOT support the `KeyValue` argument described above.

The sink for this logger is configured to accept messages according to the
log level determined by th rule set of the logging context for the given
message context.

*Remark*: Returned `logr.Logger`s are always using a sink with the base level 0,
which is potentially shifted to the level of the base `logr.Logger`
used to set up the context, when forwarding to the original sink. This means
they are always directly using the log levels 0..*n*.

It is possible to get a logging context with a predefined message context
with

```go
  ctx.WithContext("my message")
```

All loggers obtained from such a context will implicitly use the given
message context.

If no rules are configured, the default logger of the context is used
independently of the  given arguments. The given message context information is
optionally passed to the provided logger, depending on the used 
message context type.

For example, the realm is added to the logger's name.

It is also possible to provide dedicated attributes for the rule matching
process:

```go
  ctx.Logger(realm, logging.NewAttribute("test", "value")).Info("my message")
```

Such an attribute can be used as rule condition, also. This way, logging
can be enabled, for dedicated argument values of a method/function.

Both sides, the rule conditions and the message context can be a list.
For the conditions, all specified conditions must be evaluated to true, to
enable the rule. A rule is evaluated against the complete message context of
the log requests.
The default `ConditionRule` evaluates the rules against the complete log
request and a condition is *true*, if it matches at least one argument.

The rules are evaluated in the reverse order of their definition.
The first matching rule defines the finally used log level restriction and log
sink.

A `Rule` has the complete control over composing an appropriate logger.
The default condition based rule just enables the specified log level,
if all conditions match the actual log request.

For more complex conditions it is possible to compose conditions
using an `Or`, `And`, or `Not` condition.

Because `Rule` and `Condition` are interfaces, any desired behaviour
can be provided by dedicated rule and/or condition implementations.

## Default Logging Environment

This logging library provides a default logging context, it can be obtained
by

```go
  ctx := logging.DefaultContext()
```

This way it can be configured, also. It can be used for logging requests
not related to a dedicated logging context.

There is a shortcut to provide a logger for a message context based on
this default context:

```go
  logging.Log(messageContext).Debug(...)
```

or

```go
  logging.Log().V(logging.DebugLevel).Info(...
```

## Attribution Context

An `AttributionContext` is some kind of lightweight logging context.
It based on a regular context and holds a message context and standard
value (key pair) settings for issued log messages, but no rule environment
for influencing the log output and no base logger. These elements are
inherited from the base logging context.

Like a logging context an attribution context can be used to obtain loggers,
whose activation level is determined from the base logging context and the 
additional message context provided by the attribution context.

Additionally, they provide the possibility to create sub context for
more specific settings, which will be forwarded to the created logger objects. 

```go
actx := logging.NewAttributionContext(ctx, logging.NewAttribute("name", "value")).Withvalues("key", "value")
logger := actx.Logger()
logger.Info("message", "otherkey", "othervalue")
```

In this example, the attribute setting and the key/value pair will be inherited
by the generated logger and added to the log messages issued using this logger.

## Configuration

It is possible to configure a logging context from a textual configuration
using `config.ConfigureWithData(ctx, bytedata)`:

```yaml
defaultLevel: Info
rules:
  - rule:
      level: Debug
      conditions:
        - realm: github.com/mandelsoft/spiff
  - rule:
      level: Trace
      conditions:
        - attribute:
            name: test
            value:
               value: testvalue  # value is the *value* type, here
```

Rules might provide a deserialization by registering a type object
with `config.RegisterRuleType(name, typ)`. The factory type must implement the
interface `scheme.RuleType` and provide a value object
deserializable by yaml.

In a similar way it is possible to register a deserialization for
`Condition`s. The standard condition rule supports a condition deserialization
based on those registrations.

The standard names for rules are:
 - `rule`: condition rule

The standard names for conditions are:
- `and`: AND expression for a list of sub sequent conditions
- `or`: OR expression for a list of sub sequent conditions
- `not`: negate given expression
- `realm`: name for a realm condition
- `realmprefix`: name for a realm prefix condition
- `attribute`: attribute condition given by a map with `name` and `value`.
  
The config package also offers a value deserialization using
`config.RegisterValueType`. The default value type is `value`. 
It supports an `interface{}` deserialization.

For all deserialization types flat names are reserved for
the global usage by this library. Own types should use a reverse
DNS name to avoid conflicts by different users of this logging
API.

To provide own deserialization context, an own object of type
`config.Registry` can be created using `config.NewRegistry`.
The standard registry can be obtained by `config.DefaultRegistry()`

## Nesting Contexts

Logging contents can inherit from base contexts. This way the rule set,
logger and default level settings can be reused for a sub-level context.
In contrast to [attribution contexts](#attribution-context) such a context
then provides a new scope to define additional rules
and settings only valid for this nested context. Settings done here are not
visible to log requests evaluated against the base context.

If a nested context defines an own base logger, the rules inherited from the base
context are evaluated against this logger if evaluated for a message
context passed to the nested context (extended-self principle).

A logging context reusing the settings provided by the default logging
context can be obtained by:

```go
  ctx := logging.NewWithBase(logging.DefaultContext())
```

or just with 

```go
ctx := logging.DefaultContext().WithContext(<additional message context>)

to directly add a sub sequent message context.
```

Using nested logging contexts it more expensive than just using nested
attribution contexts based on a logging context, because of the inheritance
of the rule environment.
If only a subsequent settings for created loggers are required (message context,
logger names and key/value pairs) an attribution context should be preferred.

## Preconfigured Rules, Message Contexts and Conditions

### Rules

The base library provides the following basic rule implementations.
It is possible to define own more complex rules by implementing
the `logging.Rule` interface.

- `NewRule(level, conditions...)` a simple rule setting a log level
for a message context matching all given conditions.

### Message Contexts and Conditions

The message context is a set of objects describing the context of a
log message. It can be used
- to enrich the log message
- ro enrich the logger (logr.Logger features a name to represent
  the call hierarchy when passing loggers to functions)
- to control the effective log condition based of configuration rules.
  (for example to enable all Info logs for log requests with a dedicated attribute)
 
The base library already provides some ready to use conditions
and message contexts:

- `Name`(*string*)  is attached as additional name part to the logr.Logger. 
  It cannot be used to control the log state.,

- `Tag`(*string*) Just some tag for a log request.
  Used as message context, the tag name is not added to the logger name for
  the log request.

- `Realm`(*string*) the location context of a logging request. This could
  be some kind of denotation for a functional area or Go package. To obtain the
  package realm for some coding the function `logging.Package()` can be used. Used as message context, the realm name is added as additional attribute (`realm`) to log message. As condition realms only match the last realm in a message context.

- `RealmPrefix`(*string*) (only as condition) matches against a complete 
  realm tree specified by a base realm. It matches the last realm in a message
  context, only.

- `Attribute`(*string,interface{}*) the name of an arbitrary attribute with some
  value. Used as message context, the key/value pair is added to the log message.

Meaning of predefined objects in a message context:

| Element       | Rule Condition | Message Context | Logger  | LogMessage Attribute |
|---------------|:--------------:|:---------------:|:-------:|:--------------------:|
| Name          |    &check;     |     &check;     | &check; |       &cross;        |
| Tag           |    &check;     |     &check;     | &cross; |       &cross;        |
| Realm         |    &check;     |     &check;     | &cross; |  &check;  (`realm`)  |
| Attribute     |    &check;     |     &check;     | &cross; |       &check;        |
| RealmPrefix   |    &check;     |     &cross;     | &cross; |       &cross;        |
| UnboundLogger |    &cross;     |     &check;     | &check; |  &check; (partial)   |
| Context       |    &cross;     |     &check;     | &check; |  &check; (partial)   |

(* partial means, that only flattened elements matching the appropriate interface will be used)

It is possible to create own objects using the interfaces:
- `Attacher`: attach information to a logger
- `Condition`: to be usable as condition in a rule.
- `MessageContextProvider`: to be usable as provider for multiple message context.

Only objects implementing at least one of those interfaces can
usefully be passed.

An `[]MessageContext` can also be used as message context, like a `MessageContextProvider`
it wil be expanded to flat list of effective message contexts.

## Bound and Unbound Loggers

By default, logging contexts provide *bound* loggers. The activation of
such a logger is bound to the settings of the rule matching at the time
of its creation. If it does not match any rule, always context's default
level is used.

This behaviour is fine, als long such a logger is used temporarily, for example
it is created at the beginning of a dedicated call hierarchy, and passed down
the call tree. But it does not show the expected behaviour when stored in and
reused from a long-living variable. If the rule settings are changed
during its lifetime, the activation state is NOT adapted.

Nevertheless, it might be useful store and reuse a configured logger.
Configured means, that is instantiated for a dedicated long living message
context, or with a dedicated name. Such a behaviour can be achieved
by not using a logger but a logging context. Because the context does
not provide logging methods a temporary logger has to be created
on-the-fly for issuing log entries.

Another possibility is to use *unbound* loggers created with a message context
for a logging context using the `DynamicLogger` function. It provides
a logger, which keeps track of the actual settings of the context it has been
created for. Whenever the configuration changes, the next logging call will
adapt the effectively used logger on-the-fly. Such loggers keep track of the
context settings as well as the configured message context and logger values
or names (provided by the methods `WithValues` and `WithName`).

They can be used, for example for permanent worker Go routines, to
statically define the log name or standard values used for all subsequent log
requests according to the identity of the worker.

## Condition specific Loggers

Loggers are always enabled according to their effective message context
by evaluating the rules configured for the message context.
If a message context includes a tag, those loggers are enabled
if there is a rule matching this tag. But they are enabled, also, if
there are rules matching other elements in the effective message context
of the context used to retrieve the logger.

Using the `LoggerFor` methods a logger can be retrieved for a dedicated
message context without using the inherited settings from the context.
This way, the retrieved logger is enabled by rules for the
given message context, only.


## Support for special logging systems

The general *logr* logging framework acts as a wrapper for
any other logging framework to provide a uniform frontend,
which can be based on any supported base.

To support this, an adapter must be provided, for example,
the adapter for *github.com/sirupsen.logrus* is provided
by *github.com/bombsimon/logrusr*.

Because this logging framework is based on *logr* 
it can be based on any such supported logging framework.

This library contains some additional special mappings of *logr*, also.

### `logrus`

The support includes three new logrus entry formatters in
package `logrusfmt`, able to be configurable to best match
the features of this library.

- `TextFormatter` an extended logrus.TextFormatter with
  extended capabilities to render an entry.
  This is used by the adapter to generate more human-readable
  logging output supporting the special fields provided by
  this logging system.

- `TextFmtFormatter` an extended `TextFormatter` able
  to render more human-readable log messages by 
  composing a log entry's log message incorporating selected 
  log fields into a readable log message.

- `JSONFormatter` an extended logrus.JSONFormatter with
  extended capabilities to render an entry.
  This is used by the adapter to generate more readable
  logging output with a dedicated ordering of the special fields 
  provided by this logging system.

The package `logrusl` provides configuration methods to 
achieve a `logging.Context` based on *logrus* with special 
preconfigured configurations.