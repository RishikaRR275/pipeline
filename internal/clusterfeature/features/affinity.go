// Copyright © 2019 Banzai Cloud
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

package features

import (
	v1 "k8s.io/api/core/v1"

	pkgCommon "github.com/banzaicloud/pipeline/pkg/common"
)

type AffinityService struct {
	headNodePoolName string
	cluster          interface {
		NodePoolExists(nodePoolName string) bool
	}
}

func NewAffinityService(headNodePoolName string, cluster interface {
	NodePoolExists(nodePoolName string) bool
}) AffinityService {
	return AffinityService{
		headNodePoolName: headNodePoolName,
		cluster:          cluster,
	}
}

func (s AffinityService) GetHeadNodeAffinity() v1.Affinity {
	if s.headNodePoolName == "" {
		return v1.Affinity{}
	}
	if !s.cluster.NodePoolExists(s.headNodePoolName) {
		return v1.Affinity{}
	}
	return v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      pkgCommon.LabelKey,
								Operator: v1.NodeSelectorOpIn,
								Values: []string{
									s.headNodePoolName,
								},
							},
						},
					},
				},
			},
		},
	}
}
