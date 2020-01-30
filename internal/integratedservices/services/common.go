// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"strings"

	"emperror.dev/errors"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"

	"github.com/banzaicloud/pipeline/internal/integratedservices"
)

type ValuesConfig string

func NewValuesConfig(mapIn map[string]interface{}) (ValuesConfig, error) {
	out, err := yaml.Marshal(mapIn)
	if err != nil {
		return "", errors.WrapIf(err, "failed to create values config")
	}
	return ValuesConfig(out), nil
}

func (v ValuesConfig) ToMap() (map[string]interface{}, error) {
	var out = make(map[string]interface{})
	var trimmedStr = strings.TrimSpace(string(v))
	err := yaml.Unmarshal([]byte(trimmedStr), &out)
	if err != nil {
		return nil, errors.WrapIf(err, "error during converting to map")
	}

	return out, nil
}

// BindIntegratedServiceSpec binds an incoming integrated service specific raw spec (json) into the appropriate struct
func BindIntegratedServiceSpec(inputSpec integratedservices.IntegratedServiceSpec, boundSpec interface{}) error {
	if err := mapstructure.Decode(inputSpec, &boundSpec); err != nil {
		return errors.WrapIf(err, "failed to decode integrated service specification")
	}
	return nil
}
