package managed

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// An VaultSecretPublisher publishes ConnectionDetails to Vault
type VaultSecretPublisher struct {
	secret resource.Applicator
	typer  runtime.ObjectTyper
}

// NewVaultSecretPublisher returns a new VaultSecretPublisher.
func NewVaultSecretPublisher(c client.Client, ot runtime.ObjectTyper) *VaultSecretPublisher {
	return &VaultSecretPublisher{
		secret: resource.NewApplicatorWithRetry(resource.NewAPIPatchingApplicator(c),
			resource.IsAPIErrorWrapped, nil),
		typer: ot,
	}
}

// PublishConnection publishes the supplied ConnectionDetails to a Secret in the
// same namespace as the supplied Managed resource. It is a no-op if the secret
// already exists with the supplied ConnectionDetails.
func (v *VaultSecretPublisher) PublishConnection(ctx context.Context, mg resource.Managed, c ConnectionDetails) error {
	// This resource does not want to expose a connection secret.
	if mg.GetWriteConnectionSecretToReference() == nil {
		return nil
	}

	s := resource.ConnectionSecretFor(mg, resource.MustGetKind(mg, v.typer))
	s.Data = c
	return errors.Wrap(v.secret.Apply(ctx, s, resource.ConnectionSecretMustBeControllableBy(mg.GetUID())), errCreateOrUpdateSecret)
}

// UnpublishConnection is no-op since PublishConnection only creates resources
// that will be garbage collected by Kubernetes when the managed resource is
// deleted.
func (v *VaultSecretPublisher) UnpublishConnection(ctx context.Context, mg resource.Managed, c ConnectionDetails) error {
	return nil
}
