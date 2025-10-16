package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	credcpi "ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/extensions/pubsub/types/redis/identity"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "redis"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	pubsub.RegisterType(pubsub.NewPubSubType[*Spec](Type,
		pubsub.WithDesciption("a redis pubsub system.")))
	pubsub.RegisterType(pubsub.NewPubSubType[*Spec](TypeV1,
		pubsub.WithFormatSpec(`It is describe by the following field:

- **<code>serverAddr</code>**  *Address of redis server*
- **<code>channel</code>**  *pubsub channel*
- **<code>database</code>**  *database number*

  Publishing using the redis pubsub API. For every change a string message
  with the format <component>:<version> is published. If multiple repositories
  should be used, each repository should be configured with a different
  channel.
`)))
}

// Spec provides a pub sub adapter registering events at its provider.
type Spec struct {
	runtime.ObjectVersionedType
	ServerAddr string `json:"serverAddr"`
	Channel    string `json:"channel"`
	Database   int    `json:"database"`
}

var _ pubsub.PubSubSpec = (*Spec)(nil)

func New(serverurl, channel string, db int) (*Spec, error) {
	return &Spec{
		runtime.NewVersionedObjectType(Type),
		serverurl, channel, db,
	}, nil
}

func (s *Spec) PubSubMethod(repo cpi.Repository) (pubsub.PubSubMethod, error) {
	_, _, err := identity.ParseAddress(s.ServerAddr)
	if err != nil {
		return nil, err
	}

	creds, err := identity.GetCredentials(repo.GetContext(), s.ServerAddr, s.Channel, s.Database)
	if err != nil {
		return nil, err
	}
	return &Method{s, creds}, nil
}

func (s *Spec) Describe(_ cpi.Context) string {
	return fmt.Sprintf("redis pubsub system %s channel %s, database %d", s.ServerAddr, s.Channel, s.Database)
}

// Method finally publishes events.
type Method struct {
	spec  *Spec
	creds credcpi.Credentials
}

var _ pubsub.PubSubMethod = (*Method)(nil)

func (m *Method) NotifyComponentVersion(version common.NameVersion) error {
	// TODO: update to credential provider interface
	opts := &redis.Options{
		Addr: m.spec.ServerAddr,
		DB:   m.spec.Database,
	}
	if m.creds != nil {
		opts.Username = m.creds.GetProperty(identity.ATTR_USERNAME)
		opts.Password = m.creds.GetProperty(identity.ATTR_PASSWORD)
	}

	rdb := redis.NewClient(opts)
	defer rdb.Close()
	return rdb.Publish(context.Background(), m.spec.Channel, version.String()).Err()
}
