package mvn

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/mvn/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const BLOB_HANDLER_NAME = "ocm/" + resourcetypes.MVN_ARTIFACT

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
		return nil, fmt.Errorf("MVN repository url not provided")
	}

	// setup logger
	log := log.WithValues("repository", b.spec.Url)
	// identify artifact
	artifact := mvn.DeSerialize(hint)
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
		log.Debug("uploading", "file", file)
		artifact = artifact.ClassifierExtensionFrom(file)

		readHash, err := tempFs.Open(file)
		if err != nil {
			return nil, err
		}
		defer readHash.Close()
		// MD5 + SHA1 are still the most used ones in the mvn context
		hr := iotools.NewHashReader(readHash, crypto.SHA256, crypto.SHA1, crypto.MD5)
		_, _ = hr.CalcHashes()

		reader, err := tempFs.Open(file)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		err = deploy(artifact, b.spec.Url, reader, ctx.GetContext(), hr)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("done", "artifact", artifact)
	return mvn.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version, mvn.WithClassifier(artifact.Classifier), mvn.WithExtension(artifact.Extension)), nil
}

// deploy an artifact to the specified destination. See https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact
func deploy(artifact *mvn.Artifact, url string, reader io.ReadCloser, ctx accspeccpi.Context, hashes *iotools.HashReader) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, artifact.Url(url), reader)
	if err != nil {
		return err
	}
	identity.BasicAuth(req, ctx, url, artifact.GroupPath())
	// give the remote server a chance to decide based upon the checksum policy
	for k, v := range hashes.HttpHeader() {
		req.Header.Set(k, v)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("http (%d) - failed to upload artifact: %s", resp.StatusCode, string(all))
	}
	log.Debug("uploaded", "artifact", artifact, "extension", artifact.Extension, "classifier", artifact.Classifier)

	// Validate the response - especially the hash values with the ones we've tried to send
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var artifactBody Body
	err = json.Unmarshal(respBody, &artifactBody)
	if err != nil {
		return err
	}

	// let's check only SHA256 for now
	digest := hashes.GetString(crypto.SHA256)
	remoteDigest := artifactBody.Checksums[strings.ReplaceAll(strings.ToLower(crypto.SHA256.String()), "-", "")]
	if remoteDigest == "" {
		log.Warn("no checksum found for algorithm, we can't guarantee that the artifact has been uploaded correctly", "algorithm", crypto.SHA256)
	} else if remoteDigest != digest {
		return fmt.Errorf("failed to upload artifact: checksums do not match")
	}
	log.Debug("digests are ok", "remoteDigest", remoteDigest, "digest", digest)
	return nil
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
