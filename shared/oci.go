package shared

import (
	"context"
	"fmt"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmmetav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	ocmreg "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/mime"
	ocmsigning "github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const (
	SignAlgo = rsa.Algorithm
)

// Resource presents a simple layout for a resource that AddComponentVersionToRepository will use.
type Resource struct {
	Name    string
	Version string
	Data    string
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

// AddComponentVersionToRepository takes a component description and optional resources. Then pushes that component
// into the locally forwarded registry.
func AddComponentVersionToRepository(component Component, repository string, resources ...Resource) error {
	baseUrl := "http://127.0.0.1:5000/" + repository
	octx := ocm.ForContext(context.Background())
	target, err := octx.RepositoryForSpec(ocmreg.NewRepositorySpec(baseUrl, nil))

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

	for _, resource := range resources {
		err = compvers.SetResourceBlob(
			&compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    resource.Name,
					Version: resource.Version,
				},
				Type:     resourcetypes.BLOB,
				Relation: ocmmetav1.LocalRelation,
			},
			accessio.BlobAccessForString(mime.MIME_TEXT, resource.Data),
			"", nil,
		)
		if err != nil {
			return fmt.Errorf("failed to set resource blob: %w", err)
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
