/*
Copyright 2021 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRecognizedKustomizationFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "kustomization.yaml",
			path:     "/path/to/kustomization.yaml",
			expected: true,
		},
		{
			name:     "kustomization.yml",
			path:     "/path/to/kustomization.yml",
			expected: true,
		},
		{
			name:     "Kustomization",
			path:     "/path/to/Kustomization",
			expected: true,
		},
		{
			name:     "regular yaml file",
			path:     "/path/to/deployment.yaml",
			expected: false,
		},
		{
			name:     "no extension",
			path:     "/path/to/somefile",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRecognizedKustomizationFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadObjects(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(t *testing.T) (root, manifestPath string)
		expectError bool
	}{
		{
			name: "valid yaml file",
			setupFile: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				manifestPath := filepath.Join(dir, "test.yaml")
				content := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value`
				require.NoError(t, os.WriteFile(manifestPath, []byte(content), 0644))
				return dir, manifestPath
			},
			expectError: false,
		},
		{
			name: "directory instead of file",
			setupFile: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				subDir := filepath.Join(dir, "subdir")
				require.NoError(t, os.Mkdir(subDir, 0755))
				return dir, subDir
			},
			expectError: true,
		},
		{
			name: "non-existent file",
			setupFile: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				return dir, filepath.Join(dir, "missing.yaml")
			},
			expectError: true,
		},
		{
			name: "kustomization file",
			setupFile: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				kustomizationPath := filepath.Join(dir, "kustomization.yaml")
				content := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml`
				require.NoError(t, os.WriteFile(kustomizationPath, []byte(content), 0644))

				deploymentPath := filepath.Join(dir, "deployment.yaml")
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment`
				require.NoError(t, os.WriteFile(deploymentPath, []byte(deploymentContent), 0644))
				return dir, kustomizationPath
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, manifestPath := tt.setupFile(t)

			objects, err := readObjects(root, manifestPath)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, objects)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, objects)
			}
		})
	}
}
