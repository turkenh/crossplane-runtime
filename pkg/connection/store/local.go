package store

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretTypeConnection is the type of Crossplane connection secrets.
const SecretTypeConnection v1.SecretType = "connection.crossplane.io/v1alpha1"

type LocalKubernetes struct {
	secret resource.ClientApplicator
	typer  runtime.ObjectTyper
}

// NewLocalKubernetes returns a new LocalKubernetes.
func NewLocalKubernetes(c client.Client, ot runtime.ObjectTyper) LocalKubernetes {
	// NOTE(negz): We transparently inject an APIPatchingApplicator in order to maintain
	// backward compatibility with the original API of this function.
	return LocalKubernetes{
		secret: resource.ClientApplicator{
			Client: c,
			Applicator: resource.NewApplicatorWithRetry(resource.NewAPIPatchingApplicator(c),
				resource.IsAPIErrorWrapped, nil),
		},
		typer: ot,
	}
}

// PublishConnection publishes the supplied ConnectionDetails to a Secret in the
// same namespace as the supplied Managed resource. It is a no-op if the secret
// already exists with the supplied ConnectionDetails.
func (l LocalKubernetes) PublishConnection(ctx context.Context, p resource.PublishConnectionConfig, c managed.ConnectionDetails) error {
	// This resource does not want to publish a connection secret to local
	// kubernetes.
	if p.GetPublishConnectionSecretTo() == nil || p.GetPublishConnectionSecretTo().Kubernetes == nil {
		return nil
	}

	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       p.GetPublishConnectionSecretTo().Kubernetes.Namespace,
			Name:            p.GetPublishConnectionSecretTo().Name,
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(p, resource.MustGetKind(p, l.typer)))},
		},
		Type: SecretTypeConnection,
		Data: make(map[string][]byte),
	}
	s.Data = c
	return errors.Wrap(l.secret.Apply(ctx, s, resource.ConnectionSecretMustBeControllableBy(p.GetUID())), "cannot update secret")
}

// UnpublishConnection is no-op since PublishConnection only creates resources
// that will be garbage collected by Kubernetes when the managed resource is
// deleted.
func (l LocalKubernetes) UnpublishConnection(ctx context.Context, p resource.PublishConnectionConfig, c managed.ConnectionDetails) error {
	return nil
}

func (l LocalKubernetes) FetchConnectionDetails(ctx context.Context, p resource.PublishConnectionConfig) (managed.ConnectionDetails, error) {
	s := &v1.Secret{}

	cfg := p.GetPublishConnectionSecretTo()
	if cfg.Kubernetes == nil {
		return nil, nil // no kubernetes secret for this resource
	}
	if err := l.secret.Get(ctx, types.NamespacedName{Name: cfg.Name, Namespace: cfg.Kubernetes.Namespace}, s); client.IgnoreNotFound(err) != nil {
		return nil, errors.Wrap(err, "cannot get secret")
	}

	return s.Data, nil
}
