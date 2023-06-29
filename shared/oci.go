// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"fmt"
	"net/url"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmmetav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	ocmreg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/mime"
	ocmsigning "github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const (
	SignAlgo = rsa.Algorithm
)

type Resource struct {
	Name    string
	Version string
	Data    string
	Type    string
}

type ComponentRef struct {
	Name          string
	Version       string
	ComponentName string
}

// CreateOptions presents a simple layout for a resource that AddComponentVersionToRepository will use.
type CreateOptions struct {
	Resource     *Resource
	ComponentRef *ComponentRef
}

// Sign defines the two needed values to perform a component signing.
type Sign struct {
	Name string
	Key  []byte
}

// Component presents a simple layout for a component. If `Sign` is not empty, it's used to
// sign the component. It should be the byte representation of a private key.
type Component struct {
	Name    string
	Version string
	Sign    *Sign
}

// BlobResource creates a blob type resource for local access.
func BlobResource(resource Resource) ComponentModification {
	return func(compvers ocm.ComponentVersionAccess) error {
		return compvers.SetResourceBlob(
			&compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    resource.Name,
					Version: resource.Version,
				},
				Type:     resource.Type,
				Relation: ocmmetav1.LocalRelation,
			},
			accessio.BlobAccessForString(mime.MIME_TEXT, resource.Data),
			"", nil,
		)
	}
}

// ImageRefResource creates an image reference type resource.
func ImageRefResource(ref string, resource Resource) ComponentModification {
	return func(compvers ocm.ComponentVersionAccess) error {
		return compvers.SetResource(&compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:    resource.Name,
				Version: resource.Version,
			},
			Type:     resource.Type,
			Relation: ocmmetav1.ExternalRelation,
		}, ociartifact.New(ref))
	}
}

// ComponentVersionRef creates a component version reference for the given component version.
func ComponentVersionRef(ref ComponentRef) ComponentModification {
	return func(compvers ocm.ComponentVersionAccess) error {
		return compvers.SetReference(&compdesc.ComponentReference{
			ElementMeta: compdesc.ElementMeta{
				Name:    ref.Name,
				Version: ref.Version,
			},
			ComponentName: ref.ComponentName,
		})
	}
}

// ComponentModification defines functions that can modify the generated component version.
type ComponentModification func(compvers ocm.ComponentVersionAccess) error

// AddComponentVersionToRepository takes a component description and optional resources. Then pushes that component
// into the locally forwarded registry.
func AddComponentVersionToRepository(component Component, repository, scheme string, componentModifications ...ComponentModification) error {
	u, err := url.Parse("https://127.0.0.1:5000")
	if err != nil {
		return fmt.Errorf("failed to parse base url: %w", err)
	}
	u.Scheme = scheme

	// Re-parsing after scheme was set.
	u, err = url.Parse(u.String())
	if err != nil {
		return fmt.Errorf("failed to reparse base url: %w", err)
	}

	// join up with the repository.
	u = u.JoinPath(repository)

	baseURL := u.String()
	octx := ocm.FromContext(context.Background())

	target, err := octx.RepositoryForSpec(ocmreg.NewRepositorySpec(baseURL, nil))
	if err != nil {
		return fmt.Errorf("failed to create repository for spec: %w", err)
	}

	defer target.Close()

	comp, err := target.LookupComponent(component.Name)
	if err != nil {
		return fmt.Errorf("failed to look up component: %w", err)
	}

	compvers, err := comp.NewVersion(component.Version, true)
	if err != nil {
		return fmt.Errorf("failed to create new Version '%s': %w", component.Version, err)
	}

	defer compvers.Close()

	for _, modify := range componentModifications {
		if err := modify(compvers); err != nil {
			return fmt.Errorf("failed to modify component version: %w", err)
		}
	}

	if err := comp.AddVersion(compvers); err != nil {
		return fmt.Errorf("failed to add Version: %w", err)
	}

	if component.Sign != nil {
		resolver := ocm.NewCompoundResolver(target)
		opts := signing.NewOptions(
			signing.Sign(ocmsigning.DefaultHandlerRegistry().GetSigner(SignAlgo), component.Sign.Name),
			signing.Resolver(resolver),
			signing.PrivateKey(component.Sign.Name, component.Sign.Key),
			signing.Update(), signing.VerifyDigests(),
		)

		if err := opts.Complete(signingattr.Get(octx)); err != nil {
			return fmt.Errorf("failed to complete signing: %w", err)
		}

		if _, err := signing.Apply(nil, nil, compvers, opts); err != nil {
			return fmt.Errorf("failed to apply signing: %w", err)
		}
	}

	return nil
}
