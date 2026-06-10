package s3

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// regionFrom301Error mirrors the logic in Download to extract X-Amz-Bucket-Region from a 301 error.
func regionFrom301Error(err error) (string, bool) {
	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusMovedPermanently {
		if region := respErr.Response.Header.Get("X-Amz-Bucket-Region"); region != "" {
			return region, true
		}
	}
	return "", false
}

func Test_regionFrom301Error(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantRegion string
		wantOK     bool
	}{
		{
			name: "301 with region header",
			err: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{
					StatusCode: http.StatusMovedPermanently,
					Header:     http.Header{"X-Amz-Bucket-Region": []string{"eu-central-1"}},
				}},
				Err: fmt.Errorf("MovedPermanently"),
			},
			wantRegion: "eu-central-1",
			wantOK:     true,
		},
		{
			name: "301 without region header",
			err: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{
					StatusCode: http.StatusMovedPermanently,
					Header:     http.Header{},
				}},
				Err: fmt.Errorf("MovedPermanently"),
			},
			wantRegion: "",
			wantOK:     false,
		},
		{
			name: "403 with region header — not a redirect",
			err: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{
					StatusCode: http.StatusForbidden,
					Header:     http.Header{"X-Amz-Bucket-Region": []string{"eu-central-1"}},
				}},
				Err: fmt.Errorf("Forbidden"),
			},
			wantRegion: "",
			wantOK:     false,
		},
		{
			name:       "non-HTTP error",
			err:        fmt.Errorf("some network error"),
			wantRegion: "",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := regionFrom301Error(tt.err)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.wantRegion {
				t.Errorf("region = %q, want %q", got, tt.wantRegion)
			}
		})
	}
}
