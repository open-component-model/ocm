package regclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/scheme/reg"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	regref "github.com/regclient/regclient/types/ref"
)

type ClientOptions struct {
	Host    *config.Host
	Version string
}

type Client struct {
	rc  *regclient.RegClient
	ref regref.Ref
}

type pushRequest struct {
	rc   *regclient.RegClient
	desc descriptor.Descriptor
	ref  regref.Ref
}

// Commit and Status are actually not really used. Commit is a second stage operation and Status is never called in
// the library.
func (p *pushRequest) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	return p.rc.Close(ctx, p.ref)
}

func (p *pushRequest) Status() (content.Status, error) {
	return content.Status{
		Ref:   p.ref.Reference,
		Total: p.desc.Size,
	}, nil
}

var _ PushRequest = &pushRequest{}

var (
	_ Resolver = &Client{}
	_ Fetcher  = &Client{}
	_ Pusher   = &Client{}
	_ Lister   = &Client{}
)

func New(opts ClientOptions) *Client {
	rc := regclient.New(
		regclient.WithConfigHost(*opts.Host),
		regclient.WithDockerCerts(),
		regclient.WithDockerCreds(),
		regclient.WithUserAgent("containerd/"+opts.Version),
		regclient.WithRegOpts(
			// reg.WithCertDirs([]string{"."}),
			reg.WithDelay(2*time.Second, 15*time.Second),
			reg.WithRetryLimit(5),
			reg.WithCache(5*time.Minute, 500), // built in cache!! Nice!
		),
	)

	return &Client{rc: rc}
}

// Close must be called at the end of the operation.
func (c *Client) Close(ctx context.Context, ref regref.Ref) error {
	return c.rc.Close(ctx, ref)
}

func (c *Client) convertDescriptorToOCI(desc descriptor.Descriptor) ociv1.Descriptor {
	var p *ociv1.Platform
	if desc.Platform != nil {
		p = &ociv1.Platform{
			Architecture: desc.Platform.Architecture,
			OS:           desc.Platform.OS,
			OSVersion:    desc.Platform.OSVersion,
			OSFeatures:   desc.Platform.OSFeatures,
			Variant:      desc.Platform.Variant,
		}
	}

	return ociv1.Descriptor{
		MediaType:    desc.MediaType,
		Size:         desc.Size,
		Digest:       desc.Digest,
		Platform:     p,
		URLs:         desc.URLs,
		Annotations:  desc.Annotations,
		Data:         desc.Data,
		ArtifactType: desc.ArtifactType,
	}
}

func (c *Client) convertDescriptorToRegClient(desc ociv1.Descriptor) descriptor.Descriptor {
	var p *platform.Platform
	if desc.Platform != nil {
		p = &platform.Platform{
			Architecture: desc.Platform.Architecture,
			OS:           desc.Platform.OS,
			OSVersion:    desc.Platform.OSVersion,
			OSFeatures:   desc.Platform.OSFeatures,
			Variant:      desc.Platform.Variant,
		}
	}

	return descriptor.Descriptor{
		MediaType:    desc.MediaType,
		Size:         desc.Size,
		Digest:       desc.Digest,
		Platform:     p,
		URLs:         desc.URLs,
		Annotations:  desc.Annotations,
		Data:         desc.Data,
		ArtifactType: desc.ArtifactType,
	}
}

func (c *Client) Resolve(ctx context.Context, ref string) (string, ociv1.Descriptor, error) {
	// TODO: figure out what to do about closing c.rc.
	r, err := regref.New(ref)
	if err != nil {
		return "", ociv1.Descriptor{}, err
	}

	if r.Digest != "" {
		blob, err := c.rc.BlobHead(ctx, r, descriptor.Descriptor{
			Digest: digest.Digest(r.Digest),
		})
		defer blob.Close() // we can safely close it as this is not when we read it.

		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return "", ociv1.Descriptor{}, errdefs.ErrNotFound
			}

			return "", ociv1.Descriptor{}, fmt.Errorf("failed to resolve blob head: %w", err)
		}

		c.ref = r

		return ref, c.convertDescriptorToOCI(blob.GetDescriptor()), nil
	}

	// if digest is set it will use that.
	fmt.Println("we are in manifest")
	m, err := c.rc.ManifestHead(ctx, r)
	if err != nil {
		fmt.Println("we are in manifest error: ", err)
		if strings.Contains(err.Error(), "not found") {
			return "", ociv1.Descriptor{}, errdefs.ErrNotFound
		}

		return "", ociv1.Descriptor{}, fmt.Errorf("failed to get manifest: %w", err)
	}

	// update the Ref of the client to the resolved reference.
	c.ref = r

	fmt.Println("we returned")
	return ref, c.convertDescriptorToOCI(m.GetDescriptor()), nil
}

func (c *Client) Fetcher(ctx context.Context, ref string) (Fetcher, error) {
	var err error
	c.ref, err = regref.New(ref)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Pusher(ctx context.Context, ref string) (Pusher, error) {
	var err error
	c.ref, err = regref.New(ref)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Lister(ctx context.Context, ref string) (Lister, error) {
	var err error
	c.ref, err = regref.New(ref)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Push(ctx context.Context, d ociv1.Descriptor, src Source) (PushRequest, error) {
	reader, err := src.Reader()
	if err != nil {
		return nil, err
	}

	if c.isManifest(d) {
		manifestContent, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest: %w", err)
		}

		m, err := manifest.New(manifest.WithDesc(c.convertDescriptorToRegClient(d)), manifest.WithRef(c.ref), manifest.WithRaw(manifestContent))
		if err != nil {
			return nil, fmt.Errorf("failed to create a manifest: %w", err)
		}

		if err := c.rc.ManifestPut(ctx, c.ref, m); err != nil {
			return nil, err
		}

		// pushRequest closes the RC on `Commit`.
		return &pushRequest{
			desc: c.convertDescriptorToRegClient(d),
			rc:   c.rc,
			ref:  c.ref,
		}, nil
	}

	desc, err := c.rc.BlobPut(ctx, c.ref, c.convertDescriptorToRegClient(d), reader)
	if err != nil {
		return nil, err
	}

	return &pushRequest{
		desc: desc,
		rc:   c.rc,
		ref:  c.ref,
	}, nil
}

func (c *Client) Fetch(ctx context.Context, desc ociv1.Descriptor) (_ io.ReadCloser, err error) {
	defer func() {
		if cerr := c.rc.Close(ctx, c.ref); cerr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close the client after fetch: %w", cerr))
		}
	}()

	// -1 is not a thing in regclient.
	if desc.Size < 0 {
		desc.Size = 0
	}

	fmt.Println("in fetch: ", desc, c.ref)

	if c.isManifest(desc) {
		fmt.Println("in manifest: ", desc, c.ref)
		manifestContent, err := c.rc.ManifestGet(ctx, c.ref)
		if err != nil {
			return nil, err
		}

		body, err := manifestContent.RawBody()
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(body)), nil
	}

	reader, err := c.rc.BlobGet(ctx, c.ref, c.convertDescriptorToRegClient(desc))
	if err != nil {
		return nil, fmt.Errorf("failed to get the blob reader: %w", err)
	}

	return reader, nil
}

func (c *Client) List(ctx context.Context) (_ []string, err error) {
	defer func() {
		if cerr := c.rc.Close(ctx, c.ref); cerr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close the client after list: %w", cerr))
		}
	}()

	tags, err := c.rc.TagList(ctx, c.ref)
	if err != nil {
		return nil, err
	}

	return tags.Tags, nil
}

func (c *Client) isManifest(desc ociv1.Descriptor) bool {
	switch desc.MediaType {
	case images.MediaTypeDockerSchema2Manifest, images.MediaTypeDockerSchema2ManifestList,
		ociv1.MediaTypeImageManifest, ociv1.MediaTypeImageIndex:
		return true
	}

	return false
}
