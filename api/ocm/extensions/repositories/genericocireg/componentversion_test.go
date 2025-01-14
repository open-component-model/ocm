package genericocireg

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils/accessobj"
)

func TestComponentVersionContainer_Check(t *testing.T) {
	// Setup
	state, err := accessobj.NewBlobStateForBlob(accessobj.ACC_READONLY, nil, NewStateHandler("mock.test.com/state_handler", "1.0.0"))
	assert.NoError(t, err)
	repo := &RepositoryImpl{ctx: internal.DefaultContext}
	comp := &componentAccessImpl{repo: repo}
	cvc := &ComponentVersionContainer{state: state, comp: comp}

	// Test cases
	tests := []struct {
		name      string
		setup     func()
		expectErr bool
	}{
		{
			name: "valid version and name",
			setup: func() {
				cvc.version = "1.0.0"
				cvc.GetDescriptor().Version = "1.0.0"
				cvc.comp.name = "test-component"
				cvc.GetDescriptor().Name = "test-component"
			},
			expectErr: false,
		},
		{
			name: "half valid version - containing META_SEPARATOR = " + META_SEPARATOR,
			setup: func() {
				cvc.version = "0.0.1-20250108132333.build-af79499"
				cvc.GetDescriptor().Version = "0.0.1-20250108132333+af79499"
				cvc.comp.name = "test-component"
				cvc.GetDescriptor().Name = "test-component"
			},
			expectErr: false,
		},
		{
			name: "valid version - containing '+'",
			setup: func() {
				cvc.version = "0.0.1-20250108132333+af79499"
				cvc.GetDescriptor().Version = "0.0.1-20250108132333+af79499"
				cvc.comp.name = "test-component"
				cvc.GetDescriptor().Name = "test-component"
			},
			expectErr: false,
		},
		{
			name: "invalid version",
			setup: func() {
				cvc.version = "1.0.0"
				cvc.GetDescriptor().Version = "2.0.0"
			},
			expectErr: true,
		},
		{
			name: "invalid name",
			setup: func() {
				cvc.comp.name = "test-component"
				cvc.GetDescriptor().Name = "invalid-component"
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			err := cvc.Check()
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
