// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"fmt"
	"net/url"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmmetav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	ocmreg "ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/resolvers"
	"ocm.software/ocm/api/ocm/tools/signing"
	ocmsigning "ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

const (
	SignAlgo = rsa.Algorithm
)

type Resource struct {
	Name          string
	Version       string
	Data          string
	Type          string
	ExtraIdentity map[string]string
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
					Name:          resource.Name,
					Version:       resource.Version,
					ExtraIdentity: resource.ExtraIdentity,
				},
				Type:     resource.Type,
				Relation: ocmmetav1.LocalRelation,
			},
			blobaccess.ForString(mime.MIME_TEXT, resource.Data),
			"", nil, ocm.ModifyElement(true),
		)
	}
}

// ImageRefResource creates an image reference type resource.
func ImageRefResource(ref string, resource Resource) ComponentModification {
	return func(compvers ocm.ComponentVersionAccess) error {
		return compvers.SetResource(&compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:          resource.Name,
				Version:       resource.Version,
				ExtraIdentity: resource.ExtraIdentity,
			},
			Type:     resource.Type,
			Relation: ocmmetav1.ExternalRelation,
		}, ociartifact.New(ref), ocm.ModifyElement(true))
	}
}

// ComponentVersionRef creates a component version reference for the given component version.
func ComponentVersionRef(ref ComponentRef) ComponentModification {
	return func(compvers ocm.ComponentVersionAccess) error {
		return compvers.SetReference(&compdesc.Reference{
			ElementMeta: compdesc.ElementMeta{
				Name:    ref.Name,
				Version: ref.Version,
			},
			ComponentName: ref.ComponentName,
		}, ocm.ModifyElement(true))
	}
}

// ComponentModification defines functions that can modify the generated component version.
type ComponentModification func(compvers ocm.ComponentVersionAccess) error

// AddComponentVersionToRepository takes a component description and optional resources. Then pushes that component
// into the locally forwarded registry.
func AddComponentVersionToRepository(component Component, scheme string, componentModifications ...ComponentModification) error {
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

	if err := comp.AddVersion(compvers, true); err != nil {
		return fmt.Errorf("failed to add Version: %w", err)
	}

	if component.Sign != nil {
		resolver := resolvers.NewCompoundResolver(target)
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
