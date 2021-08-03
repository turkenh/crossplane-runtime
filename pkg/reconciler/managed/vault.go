package managed

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// An VaultSecretPublisher publishes ConnectionDetails to Vault
type VaultSecretPublisher struct {
	client *vaultapi.Client
	typer  runtime.ObjectTyper
}

// NewVaultSecretPublisher returns a new VaultSecretPublisher.
func NewVaultSecretPublisher(ot runtime.ObjectTyper) *VaultSecretPublisher {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = os.Getenv("VAULT_ADDR")

	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil
	}
	token, err := ioutil.ReadFile(filepath.Clean(os.Getenv("VAULT_TOKEN_FILE")))
	if err != nil {
		return nil
	}

	client.SetToken(strings.TrimSpace(string(token)))

	return &VaultSecretPublisher{
		client: client,
		typer:  ot,
	}
}

// PublishConnection publishes the supplied ConnectionDetails to a Vault Secret
func (v *VaultSecretPublisher) PublishConnection(ctx context.Context, mg resource.Managed, c ConnectionDetails) error {
	vs := mg.GetPublishConnectionDetailsToVaultSink()
	// This resource does not want to publish a connection secret to Vault.
	if vs == nil {
		return nil
	}

	path := vs.Path
	d := make(map[string]interface{}, len(c))
	for k, val := range c {
		d[k] = val
	}
	_, err := v.client.Logical().Write(path, d)
	return errors.Wrap(err, "failed to write secret to Vault")
}

// UnpublishConnection deletes connection details from Vault
func (v *VaultSecretPublisher) UnpublishConnection(ctx context.Context, mg resource.Managed, c ConnectionDetails) error {
	vs := mg.GetPublishConnectionDetailsToVaultSink()
	// This resource does not want to publish a connection secret to Vault.
	if vs == nil {
		return nil
	}
	_, err := v.client.Logical().Delete(vs.Path)
	return errors.Wrap(err, "failed to delete secret from Vault")
}
