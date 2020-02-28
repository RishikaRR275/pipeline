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

package clustersetup

import (
	"context"
	"time"

	"emperror.dev/errors"
	processClient "github.com/banzaicloud/pipeline/internal/app/pipeline/process/client"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CreatePipelineNamespaceActivityName = "create-pipeline-namespace"

type CreatePipelineNamespaceActivity struct {
	namespace string

	clientFactory ClientFactory

	processLogger *processClient.Client
}

// NewCreatePipelineNamespaceActivity returns a new CreatePipelineNamespaceActivity.
func NewCreatePipelineNamespaceActivity(
	namespace string,
	clientFactory ClientFactory,
	processLogger *processClient.Client,
) CreatePipelineNamespaceActivity {
	return CreatePipelineNamespaceActivity{
		namespace:     namespace,
		clientFactory: clientFactory,
		processLogger: processLogger,
	}
}

type CreatePipelineNamespaceActivityInput struct {
	// Kubernetes cluster config secret ID.
	ConfigSecretID string
}

func (a CreatePipelineNamespaceActivity) Execute(ctx context.Context, input CreatePipelineNamespaceActivityInput) error {
	{
		ainfo := activity.GetInfo(ctx)
		pe := processClient.ProcessEvent{
			ProcessID: ainfo.WorkflowExecution.ID,
			Timestamp: ainfo.StartedTimestamp,
			Name:      ainfo.ActivityType.Name,
			Log:       ainfo.ActivityType.Name + " started",
		}

		err := a.processLogger.LogEvent(context.Background(), pe)
		if err != nil {
			activity.GetLogger(ctx).Warn("failed to write process event", zap.Error(err))
		}

		defer func() {
			pe.Timestamp = time.Now()
			pe.Log = ainfo.ActivityType.Name + " finished"

			err := a.processLogger.LogEvent(context.Background(), pe)
			if err != nil {
				activity.GetLogger(ctx).Warn("failed to write process event", zap.Error(err))
			}
		}()
	}

	client, err := a.clientFactory.FromSecret(ctx, input.ConfigSecretID)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: a.namespace,
			Labels: map[string]string{
				"scan":  "noscan",
				"name":  a.namespace,
				"owner": "pipeline",
			},
		},
	})

	if err != nil && k8sapierrors.IsAlreadyExists(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	return nil
}
