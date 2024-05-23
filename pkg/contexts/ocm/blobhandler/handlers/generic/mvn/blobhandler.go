package mvn

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/maven"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/optionutils"
	"io"
	"net/http"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/maven/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const BlobHandlerName = "ocm/" + resourcetypes.MVN_ARTIFACT

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

var log = logging.DynamicLogger(identity.REALM)

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, resourceType string, hint string, _ cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	// check conditions
	if b.spec == nil {
		return nil, nil
	}
	mimeType := blob.MimeType()
	if resourcetypes.MVN_ARTIFACT != resourceType {
		log.Debug("not a MVN artifact", "resourceType", resourceType)
		return nil, nil
	}
	if mime.MIME_TGZ != mimeType {
		log.Debug("not a tarball, can't be a complete mvn GAV", "mimeType", mimeType)
		return nil, nil
	}
	if b.spec.Url == "" {
		return nil, errors.New("MVN repository url not provided")
	}

	// setup logger
	log := log.WithValues("repository", b.spec.Url)
	// identify artifact
	artifact, err := maven.Parse(hint)
	if err != nil {
		return nil, err
	}
	log = log.WithValues("groupId", artifact.GroupId, "artifactId", artifact.ArtifactId, "version", artifact.Version)
	log.Debug("identified")

	blobReader, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer blobReader.Close()
	tempFs, err := tarutils.ExtractTgzToTempFs(blobReader)
	if err != nil {
		return nil, err
	}
	defer vfs.Cleanup(tempFs)
	files, err := tarutils.ListSortedFilesInDir(tempFs, "", false)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		e := func() (err error) {
			log.Debug("uploading", "file", file)
			err = artifact.SetClassifierExtensionBy(file)
			if err != nil {
				return
			}
			readHash, err := tempFs.Open(file)
			if err != nil {
				return
			}
			defer readHash.Close()
			// MD5 + SHA1 are still the most used ones in the mvn context
			hr := iotools.NewHashReader(readHash, crypto.SHA256, crypto.SHA1, crypto.MD5)
			_, err = hr.CalcHashes()
			if err != nil {
				return
			}
			reader, err := tempFs.Open(file)
			if err != nil {
				return
			}
			defer reader.Close()
			err = deploy(artifact, b.spec.Url, reader, ctx.GetContext(), hr)
			return
		}()
		if e != nil {
			return nil, e
		}
	}

	log.Debug("done", "artifact", artifact)
	return access.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version, maven.WithClassifier(optionutils.AsValue(artifact.Classifier)), maven.WithExtension(optionutils.AsValue(artifact.Extension))), nil
}

// deploy an artifact to the specified destination. See https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact
func deploy(artifact *maven.Coordinates, url string, reader io.ReadCloser, ctx accspeccpi.Context, hashes *iotools.HashReader) (err error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, artifact.Url(url), reader)
	if err != nil {
		return
	}
	err = identity.BasicAuth(req, ctx, url, artifact.GroupPath())
	if err != nil {
		return
	}
	// give the remote server a chance to decide based upon the checksum policy
	for k, v := range hashes.HttpHeader() {
		req.Header.Set(k, v)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		all, e := io.ReadAll(resp.Body)
		if e != nil {
			return e
		}
		return fmt.Errorf("http (%d) - failed to upload artifact: %s", resp.StatusCode, string(all))
	}
	log.Debug("uploaded", "artifact", artifact, "extension", artifact.Extension, "classifier", artifact.Classifier)

	// Validate the response - especially the hash values with the ones we've tried to send
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var artifactBody Body
	err = json.Unmarshal(respBody, &artifactBody)
	if err != nil {
		return
	}

	// let's check only SHA256 for now
	digest := hashes.GetString(crypto.SHA256)
	remoteDigest := artifactBody.Checksums[strings.ReplaceAll(strings.ToLower(crypto.SHA256.String()), "-", "")]
	if remoteDigest == "" {
		log.Warn("no checksum found for algorithm, we can't guarantee that the artifact has been uploaded correctly", "algorithm", crypto.SHA256)
	} else if remoteDigest != digest {
		return errors.New("failed to upload artifact: checksums do not match")
	}
	log.Debug("digests are ok", "remoteDigest", remoteDigest, "digest", digest)
	return
}

// Body is the response struct of a deployment from the MVN repository (JFrog Artifactory).
type Body struct {
	Repo        string            `json:"repo"`
	Path        string            `json:"path"`
	DownloadUri string            `json:"downloadUri"`
	Uri         string            `json:"uri"`
	MimeType    string            `json:"mimeType"`
	Size        string            `json:"size"`
	Checksums   map[string]string `json:"checksums"`
}
