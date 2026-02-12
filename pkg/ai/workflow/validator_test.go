package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		workflow   Workflow
		wantErr    bool
		errContain string
	}{
		{
			name: "valid workflow",
			workflow: Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     "1.0.0",
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing metadata id",
			workflow: Workflow{
				Metadata: Metadata{
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     "1.0.0",
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing metadata name",
			workflow: Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Description: "A test workflow",
					Version:     "1.0.0",
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing metadata description",
			workflow: Workflow{
				Metadata: Metadata{
					ID:      "test-workflow",
					Name:    "Test Workflow",
					Version: "1.0.0",
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid semver version",
			workflow: Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     "invalid-version",
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing steps",
			workflow: Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     "1.0.0",
				},
				Steps: []Step{},
			},
			wantErr: true,
		},
		{
			name: "step missing id",
			workflow: Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     "1.0.0",
				},
				Steps: []Step{
					{
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "step missing instructions",
			workflow: Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     "1.0.0",
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.workflow)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSemver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid semver 1.0.0",
			version: "1.0.0",
			wantErr: false,
		},
		{
			name:    "valid semver 0.1.0",
			version: "0.1.0",
			wantErr: false,
		},
		{
			name:    "valid semver 10.20.30",
			version: "10.20.30",
			wantErr: false,
		},
		{
			name:    "valid semver with prerelease 1.0.0-alpha",
			version: "1.0.0-alpha",
			wantErr: false,
		},
		{
			name:    "valid semver with build metadata 1.0.0+20130313144700",
			version: "1.0.0+20130313144700",
			wantErr: false,
		},
		{
			name:    "valid semver with prerelease and build 1.0.0-beta+exp.sha.5114f85",
			version: "1.0.0-beta+exp.sha.5114f85",
			wantErr: false,
		},
		{
			name:    "invalid semver - missing patch",
			version: "1.0",
			wantErr: true,
		},
		{
			name:    "invalid semver - not a version",
			version: "invalid",
			wantErr: true,
		},
		{
			name:    "invalid semver - leading zeros",
			version: "01.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := Workflow{
				Metadata: Metadata{
					ID:          WorkflowId("test-workflow"),
					Name:        "Test Workflow",
					Description: "A test workflow",
					Version:     tt.version,
				},
				Steps: []Step{
					{
						ID:           "step1",
						Name:         "Step 1",
						Description:  "First step",
						Instructions: []string{"Do something"},
					},
				},
			}

			err := Validate(workflow)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
