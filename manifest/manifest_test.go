package manifest_test

import (
	"testing"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/manifest"
	"github.com/stretchr/testify/assert"
)

func TestManifestLoad(t *testing.T) {
	m, err := testdataManifest("full", manifest.Environment{"FOO": "bar", "SECRET": "shh"})
	if !assert.NoError(t, err) {
		return
	}

	n := &manifest.Manifest{
		Environment: manifest.Environment{
			"DEVELOPMENT": "false",
			"SECRET":      "shh",
		},
		Keys: manifest.Keys{
			manifest.Key{
				Name: "master",
			},
		},
		Queues: manifest.Queues{
			manifest.Queue{
				Name: "traffic",
			},
		},
		Resources: manifest.Resources{
			manifest.Resource{
				Name: "database",
				Type: "postgres",
			},
		},
		Services: manifest.Services{
			manifest.Service{
				Name: "api",
				Build: manifest.ServiceBuild{
					Path: "api",
				},
				Certificate: "foo.example.org",
				Command:     "",
				Environment: []string{
					"DEVELOPMENT=false",
					"SECRET",
				},
				Health: manifest.ServiceHealth{
					Path:     "/",
					Interval: 10,
					Timeout:  9,
				},
				Port:      manifest.ServicePort{Port: 1000, Scheme: "http"},
				Resources: []string{"database"},
				Scale: manifest.ServiceScale{
					Count:  &manifest.ServiceScaleCount{Min: 3, Max: 10},
					Memory: 256,
				},
				Test: "make  test",
			},
			manifest.Service{
				Name:    "proxy",
				Command: "bash",
				Health: manifest.ServiceHealth{
					Path:     "/auth",
					Interval: 5,
					Timeout:  4,
				},
				Image: "ubuntu:16.04",
				Environment: []string{
					"SECRET",
				},
				Port: manifest.ServicePort{Port: 2000, Scheme: "https"},
				Scale: manifest.ServiceScale{
					Count:  &manifest.ServiceScaleCount{Min: 1, Max: 1},
					Memory: 512,
				},
			},
			manifest.Service{
				Name: "foo",
				Build: manifest.ServiceBuild{
					Path: ".",
				},
				Command: "foo",
				Health: manifest.ServiceHealth{
					Interval: 5,
					Path:     "/",
					Timeout:  3,
				},
				Port: manifest.ServicePort{Port: 3000, Scheme: "https"},
				Scale: manifest.ServiceScale{
					Count:  &manifest.ServiceScaleCount{Min: 0, Max: 0},
					Memory: 256,
				},
			},
			manifest.Service{
				Name: "bar",
				Build: manifest.ServiceBuild{
					Path: ".",
				},
				Command: "",
				Health: manifest.ServiceHealth{
					Interval: 5,
					Path:     "/",
					Timeout:  4,
				},
				Scale: manifest.ServiceScale{
					Count:  &manifest.ServiceScaleCount{Min: 1, Max: 1},
					Memory: 256,
				},
			},
		},
		Tables: manifest.Tables{
			manifest.Table{
				Name: "proxies",
				Indexes: []string{
					"password",
				},
			},
			manifest.Table{
				Name: "traffic",
				Indexes: []string{
					"proxy:started",
				},
			},
		},
		Workflows: manifest.Workflows{
			{
				Type:    "change",
				Trigger: "close",
				Steps: manifest.WorkflowSteps{
					{Type: "delete", Target: "staging/praxis-$branch"},
				},
			},
			{
				Type:    "change",
				Trigger: "create",
				Steps: manifest.WorkflowSteps{
					{Type: "test"},
					{Type: "create", Target: "staging/praxis-$branch"},
					{Type: "deploy", Target: "staging/praxis-$branch"},
				},
			},
			{
				Type:    "change",
				Trigger: "update",
				Steps: manifest.WorkflowSteps{
					{Type: "test"},
					{Type: "deploy", Target: "staging/praxis-$branch"},
				},
			},
			{
				Type:    "merge",
				Trigger: "demo",
				Steps: manifest.WorkflowSteps{
					{Type: "deploy", Target: "demo/praxis-demo"},
				},
			},
			{
				Type:    "merge",
				Trigger: "master",
				Steps: manifest.WorkflowSteps{
					{Type: "test"},
					{Type: "deploy", Target: "staging/praxis-staging"},
					{Type: "copy", Target: "production/praxis-production"},
				},
			},
		},
	}

	assert.Equal(t, n, m)
}

func TestManifestLoadInvalid(t *testing.T) {
	m, err := testdataManifest("invalid.1", manifest.Environment{})
	assert.Nil(t, m)
	assert.Error(t, err, "yaml: line 2: did not find expected comment or line break")

	m, err = testdataManifest("invalid.2", manifest.Environment{})
	assert.Nil(t, m)
	assert.Error(t, err, "yaml: line 3: mapping values are not allowed in this context")
}

func testdataManifest(name string, env manifest.Environment) (*manifest.Manifest, error) {
	data, err := helpers.Testdata(name)
	if err != nil {
		return nil, err
	}

	return manifest.Load(data, env)
}
