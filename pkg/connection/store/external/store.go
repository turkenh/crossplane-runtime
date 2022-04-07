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

package external

import (
	"context"
	"path/filepath"

	"google.golang.org/grpc"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection/store"
	pb "github.com/crossplane/crossplane-runtime/pkg/connection/store/external/v1alpha1"
)

type SecretStore struct {
	pluginConfigRef v1.TypedReference

	defaultScope string

	conn *grpc.ClientConn
}

// NewSecretStore returns a new Kubernetes SecretStore.
func NewSecretStore(_ context.Context, _ client.Client, cfg v1.SecretStoreConfig) (*SecretStore, error) {
	return &SecretStore{
		pluginConfigRef: cfg.External.PluginConfigRef,
		defaultScope:    cfg.DefaultScope,
	}, nil
}

func (ss *SecretStore) ReadKeyValues(ctx context.Context, n store.ScopedName, s *store.Secret) error {
	conn, err := grpc.Dial("localhost:8099")
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	resp, err := pb.NewExternalSecretStoreServiceClient(conn).GetSecret(ctx, &pb.GetSecretRequest{
		Config: &pb.ConfigReference{
			ApiVersion: "secrets.crossplane.io.secrets.crossplane.io",
			Kind:       "",
			Name:       "",
		},
		Secret: &pb.Secret{
			ScopedName: "crossplane-system/ess-conn",
			Metadata:   nil,
			Data:       nil,
		},
	})
	if err != nil {
		return err
	}
	s.Data = resp.Secret.Data
	return nil
}

func (ss *SecretStore) WriteKeyValues(ctx context.Context, s *store.Secret, wo ...store.WriteOption) (changed bool, err error) {
	resp, err := pb.NewExternalSecretStoreServiceClient(ss.conn).ApplySecret(ctx, &pb.ApplySecretRequest{
		Config: &pb.ConfigReference{
			ApiVersion: ss.pluginConfigRef.APIVersion,
			Kind:       ss.pluginConfigRef.Kind,
			Name:       ss.pluginConfigRef.Name,
		},
		Secret: &pb.Secret{
			ScopedName: filepath.Join(s.Scope, s.Name),
			Metadata:   s.Metadata.Labels,
			Data:       s.Data,
		},
	})
	if err != nil {
		return false, err
	}
	return resp.Changed, nil
}

func (ss *SecretStore) DeleteKeyValues(ctx context.Context, s *store.Secret, do ...store.DeleteOption) error {
	panic("implement me")
}

func (ss *SecretStore) path(s store.ScopedName) string {
	if s.Scope != "" {
		return filepath.Join(s.Scope, s.Name)
	}
	return filepath.Join(ss.defaultParentPath, s.Name)
}
