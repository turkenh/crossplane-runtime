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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// ProviderCredentials required to authenticate.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret;InjectedIdentity;Environment;Filesystem
	Source xpv1.CredentialsSource `json:"source"`

	xpv1.CommonCredentialSelectors `json:",inline"`
}

// IdentityType used to authenticate to the Kubernetes API.
type IdentityType string

const (
	IdentityTypeGoogleApplicationCredentials = "GoogleApplicationCredentials"
)

// KubeIdentity used to authenticate.
type KubeIdentity struct {
	// Type of identity.
	// +kubebuilder:validation:Enum=GoogleApplicationCredentials
	Type IdentityType `json:"type"`

	ProviderCredentials `json:",inline"`
}

// A KubeProviderConfigSpec defines the desired state of a ProviderConfig.
type KubeProviderConfigSpec struct {
	// Credentials used to connect to the Kubernetes API. Typically, a
	// kubeconfig file. Use InjectedIdentity for in-cluster config.
	Credentials ProviderCredentials `json:"credentials"`
	// Identity used to authenticate to the Kubernetes API. The identity
	// credentials can be used to supplement kubeconfig 'credentials', for
	// example by configuring a bearer token source such as OAuth.
	// +optional
	Identity *KubeIdentity `json:"identity,omitempty"`
}
