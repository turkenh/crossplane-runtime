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

package kube

import (
	"context"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/apis/common/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/kube/gke"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

const (
	errCreateRestConfig   = "cannot create new REST config"
	errExtractCredentials = "cannot extract credentials"

	errExtractGoogleCredentials = "cannot extract Google credentials"
	errInjectGoogleCredentials  = "cannot inject Google credentials"
)

type CredentialsProvider interface {
	GetCredentials() v1alpha1.ProviderCredentials
}

type IdentityProvider interface {
	GetIdentity() *v1alpha1.KubeIdentity
}

type ProviderConfig interface {
	resource.ProviderConfig

	CredentialsProvider
	IdentityProvider
}

func GetRestConfig(ctx context.Context, kube client.Client, pc ProviderConfig) (*rest.Config, error) {
	if pc == nil {
		return nil, errors.New("nil provider config")
	}

	var rc *rest.Config
	var err error

	switch cd := pc.GetCredentials(); cd.Source { //nolint:exhaustive
	case xpv1.CredentialsSourceInjectedIdentity:
		rc, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, errCreateRestConfig)
		}
	default:
		kc, err := resource.CommonCredentialExtractor(ctx, cd.Source, kube, cd.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, errExtractCredentials)
		}

		if rc, err = RestConfigForKubeconfig(kc); err != nil {
			return nil, errors.Wrap(err, errCreateRestConfig)
		}
	}

	// NOTE(negz): We don't currently check the identity type because at the
	// time of writing there's only one valid value (Google App Creds), and
	// that value is required.
	if id := pc.GetIdentity(); id != nil {
		creds, err := resource.CommonCredentialExtractor(ctx, id.Source, kube, id.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, errExtractGoogleCredentials)
		}

		if err := gke.WrapRESTConfig(ctx, rc, creds, gke.DefaultScopes...); err != nil {
			return nil, errors.Wrap(err, errInjectGoogleCredentials)
		}
	}

	return rc, nil
}
