package connection

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// DetailsFetcher fetches the connection details of the resource.
type DetailsFetcher interface {
	// FetchConnectionDetails for the supplied resource
	FetchConnectionDetails(ctx context.Context, p resource.PublishConnectionConfig) (managed.ConnectionDetails, error)
}

type DetailsPublisher interface {
	// Todo(turkenh): Do we need a ConfigureClient() func or would it be ok to
	//  do that in PublishConnection/UnpublishConnection? It might be required
	//  if Publish/Unpublish called multiple times with the same config.

	// PublishConnection details for the supplied resource. Publishing
	// must be additive; i.e. if details (a, b, c) are published, subsequently
	// publicing details (b, c, d) should update (b, c) but not remove a.
	PublishConnection(ctx context.Context, p resource.PublishConnectionConfig, c managed.ConnectionDetails) error

	// UnpublishConnection details for the supplied resource.
	UnpublishConnection(ctx context.Context, p resource.PublishConnectionConfig, c managed.ConnectionDetails) error
}

type SecretStore interface {
	DetailsPublisher
	DetailsFetcher
}
