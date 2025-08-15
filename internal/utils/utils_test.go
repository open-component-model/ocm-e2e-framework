/*
Copyright 2020 The Flux authors

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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestSplitKubeConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "single path",
			path:     "/home/user/.kube/config",
			expected: []string{"/home/user/.kube/config"},
		},
		{
			name:     "multiple paths unix",
			path:     "/home/user/.kube/config:/etc/kubernetes/config",
			expected: []string{"/home/user/.kube/config", "/etc/kubernetes/config"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{""},
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name     string
			path     string
			expected []string
		}{
			name:     "multiple paths windows",
			path:     "C:\\Users\\user\\.kube\\config;C:\\etc\\kubernetes\\config",
			expected: []string{"C:\\Users\\user\\.kube\\config", "C:\\etc\\kubernetes\\config"},
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitKubeConfigPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsItemString(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "grape",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "apple",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    []string{"apple", "", "cherry"},
			item:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsItemString(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsEqualFoldItemString(t *testing.T) {
	tests := []struct {
		name         string
		slice        []string
		item         string
		expectedItem string
		expectedBool bool
	}{
		{
			name:         "exact match",
			slice:        []string{"Apple", "Banana", "Cherry"},
			item:         "Banana",
			expectedItem: "Banana",
			expectedBool: true,
		},
		{
			name:         "case insensitive match",
			slice:        []string{"Apple", "Banana", "Cherry"},
			item:         "apple",
			expectedItem: "Apple",
			expectedBool: true,
		},
		{
			name:         "no match",
			slice:        []string{"Apple", "Banana", "Cherry"},
			item:         "grape",
			expectedItem: "",
			expectedBool: false,
		},
		{
			name:         "empty slice",
			slice:        []string{},
			item:         "apple",
			expectedItem: "",
			expectedBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, found := ContainsEqualFoldItemString(tt.slice, tt.item)
			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedBool, found)
		})
	}
}

func TestParseNamespacedName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected types.NamespacedName
	}{
		{
			name:  "with namespace",
			input: "default/my-resource",
			expected: types.NamespacedName{
				Namespace: "default",
				Name:      "my-resource",
			},
		},
		{
			name:  "without namespace",
			input: "my-resource",
			expected: types.NamespacedName{
				Name: "my-resource",
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: types.NamespacedName{
				Name: "",
			},
		},
		{
			name:  "multiple slashes",
			input: "ns/sub/resource",
			expected: types.NamespacedName{
				Name: "ns/sub/resource",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseNamespacedName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseObjectKindName(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedKind string
		expectedName string
	}{
		{
			name:         "with kind",
			input:        "Pod/my-pod",
			expectedKind: "Pod",
			expectedName: "my-pod",
		},
		{
			name:         "without kind",
			input:        "my-pod",
			expectedKind: "",
			expectedName: "my-pod",
		},
		{
			name:         "empty input",
			input:        "",
			expectedKind: "",
			expectedName: "",
		},
		{
			name:         "multiple slashes",
			input:        "Pod/sub/resource",
			expectedKind: "",
			expectedName: "Pod/sub/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, name := ParseObjectKindName(tt.input)
			assert.Equal(t, tt.expectedKind, kind)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

func TestParseObjectKindNameNamespace(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedKind      string
		expectedName      string
		expectedNamespace string
	}{
		{
			name:              "full format",
			input:             "Pod/my-pod.default",
			expectedKind:      "Pod",
			expectedName:      "my-pod",
			expectedNamespace: "default",
		},
		{
			name:              "no namespace",
			input:             "Pod/my-pod",
			expectedKind:      "Pod",
			expectedName:      "my-pod",
			expectedNamespace: "",
		},
		{
			name:              "no kind",
			input:             "my-pod.default",
			expectedKind:      "",
			expectedName:      "my-pod",
			expectedNamespace: "default",
		},
		{
			name:              "multiple dots in name",
			input:             "Pod/my.service.pod.default",
			expectedKind:      "Pod",
			expectedName:      "my.service.pod",
			expectedNamespace: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, name, namespace := ParseObjectKindNameNamespace(tt.input)
			assert.Equal(t, tt.expectedKind, kind)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedNamespace, namespace)
		})
	}
}

func TestCompatibleVersion(t *testing.T) {
	tests := []struct {
		name     string
		binary   string
		target   string
		expected bool
	}{
		{
			name:     "same versions",
			binary:   "v2.1.0",
			target:   "v2.1.0",
			expected: true,
		},
		{
			name:     "compatible minor versions",
			binary:   "v2.1.0",
			target:   "v2.1.5",
			expected: true,
		},
		{
			name:     "incompatible major versions",
			binary:   "v1.5.0",
			target:   "v2.1.0",
			expected: false,
		},
		{
			name:     "incompatible minor versions",
			binary:   "v2.0.0",
			target:   "v2.1.0",
			expected: false,
		},
		{
			name:     "prerelease binary",
			binary:   "v2.1.0-rc.1",
			target:   "v2.0.0",
			expected: true,
		},
		{
			name:     "invalid binary version",
			binary:   "invalid",
			target:   "v2.1.0",
			expected: false,
		},
		{
			name:     "invalid target version",
			binary:   "v2.1.0",
			target:   "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompatibleVersion(tt.binary, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}