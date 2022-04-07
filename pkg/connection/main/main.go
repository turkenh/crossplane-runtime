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

package main

import (
	"context"
	"fmt"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection/store"
	"github.com/crossplane/crossplane-runtime/pkg/connection/store/external"
)

func main() {
	e := v1.SecretStoreExternal
	ss, err := external.NewSecretStore(context.Background(), nil, v1.SecretStoreConfig{
		Type:         &e,
		DefaultScope: "crossplane-system",
		External: &v1.ExternalSecretStoreConfig{
			PluginConfigRef: v1.TypedReference{
				APIVersion: "secrets.crossplane.io/v1alpha1",
				Kind:       "VaultStoreConfig",
				Name:       "in-cluster",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	s := &store.Secret{}
	err = ss.ReadKeyValues(context.Background(), store.ScopedName{Name: "hasan", Scope: "crossplane-system"}, s)
	if err != nil {
		panic(err)
	}

	fmt.Printf("success with data: %v\n", s.Data)
}
