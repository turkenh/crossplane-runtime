/*
Copyright 2022 The Crossplane Authors.

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

package vault

import (
	"context"
	"path/filepath"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection/secret/store"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
)

// Error strings.
const (
	errNoConfig     = "no Vault config provided"
	errNewClient    = "cannot create new client"
	errExtractToken = "cannot extract token"

	errRead         = "cannot read secret"
	errReadToAppend = "cannot read secret to append keys"
	errWrite        = "cannot write secret"
	errDelete       = "cannot delete secret"
)

type Client interface {
	Read(path string) (*api.Secret, error)
	Write(path string, data map[string]interface{}) (*api.Secret, error)
	Delete(path string) (*api.Secret, error)
}

type SecretStore struct {
	client Client

	pathPrefix        string
	defaultParentPath string
}

func NewSecretStore(ctx context.Context, kube client.Client, cfg v1.SecretStoreConfig) (store.Store, error) {
	if cfg.Vault == nil {
		return nil, errors.New(errNoConfig)
	}
	if cfg.Vault.Auth.Method != v1.VaultAuthToken {
		return nil, errors.Errorf("%q auth not supported yet, please use Token auth", cfg.Vault.Auth.Method)
	}
	if cfg.Vault.Auth.Token == nil {
		return nil, errors.New("token auth configured but no token provided")
	}
	vCfg := api.DefaultConfig()
	vCfg.Address = cfg.Vault.Server

	c, err := api.NewClient(vCfg)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	t, err := resource.CommonCredentialExtractor(ctx, cfg.Vault.Auth.Token.Source, kube, cfg.Vault.Auth.Token.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errExtractToken)
	}

	c.SetToken(string(t))

	return &SecretStore{
		client: c.Logical(),

		pathPrefix:        cfg.Vault.PathPrefix,
		defaultParentPath: cfg.DefaultScope,
	}, nil
}

func (ss *SecretStore) ReadKeyValues(_ context.Context, i store.SecretInstance) (store.KeyValues, error) {
	// TODO(turkenh): Handle not found
	s, err := ss.client.Read(ss.pathForSecretInstance(i))
	if err != nil {
		return nil, errors.Wrap(err, errRead)
	}
	// TODO(turkenh): debug log s.Warnings ?
	kv := make(map[string][]byte, len(s.Data))
	for k, v := range s.Data {
		kv[k] = []byte(v.(string))
	}
	return kv, nil
}

func (ss *SecretStore) WriteKeyValues(_ context.Context, i store.SecretInstance, kv store.KeyValues) error {
	if len(kv) == 0 {
		// Nothing to write
		return nil
	}
	s, err := ss.client.Read(ss.pathForSecretInstance(i))
	if err != nil {
		return errors.Wrap(err, errReadToAppend)
	}

	var existing map[string]interface{}
	if s != nil {
		existing = s.Data
	}
	data := make(map[string]interface{}, len(kv)+len(existing))
	for k, v := range kv {
		// Note(turkenh): value here is or type []byte, it is stored as base64
		// encoded in Vault if we don't cast it to string. This could be a
		// configuration option if needed.
		data[k] = string(v)
	}
	for k, v := range existing {
		data[k] = v
	}

	_, err = ss.client.Write(ss.pathForSecretInstance(i), data)
	// TODO(turkenh): debug log s.Warnings ?
	return errors.Wrap(err, errWrite)
}

func (ss *SecretStore) DeleteKeyValues(_ context.Context, i store.SecretInstance, kv store.KeyValues) error {
	_, err := ss.client.Delete(ss.pathForSecretInstance(i))
	return errors.Wrap(err, errDelete)
}

func (ss *SecretStore) pathForSecretInstance(i store.SecretInstance) string {
	if i.Scope != "" {
		return filepath.Clean(filepath.Join(ss.pathPrefix, i.Scope, i.Name))
	}
	return filepath.Clean(filepath.Join(ss.pathPrefix, ss.defaultParentPath, i.Name))
}
