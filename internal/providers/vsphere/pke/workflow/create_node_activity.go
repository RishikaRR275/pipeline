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

package workflow

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"emperror.dev/errors"
	"github.com/banzaicloud/pipeline/internal/providers/pke/pkeworkflow/pkeworkflowadapter"
	"github.com/ghodss/yaml"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/cadence/activity"
)

// DeleteNodeActivityName is the default registration name of the activity
const DeleteNodeActivityName = "pke-vsphere-delete-node"

// DeleteNodeActivity represents an activity for creating a vSphere virtual machine
type DeleteNodeActivity struct {
	vmomiClientFactory *VMOMIClientFactory
	tokenGenerator     pkeworkflowadapter.TokenGenerator
}

// MakeDeleteNodeActivity returns a new DeleteNodeActivity
func MakeDeleteNodeActivity(vmomiClientFactory *VMOMIClientFactory) DeleteNodeActivity {
	return DeleteNodeActivity{
		vmomiClientFactory: vmomiClientFactory,
	}
}

// DeleteNodeActivityInput represents the input needed for executing a DeleteNodeActivity
type DeleteNodeActivityInput struct {
	OrganizationID uint
	ClusterID      uint
	SecretID       string
	ClusterName    string
	//HTTPProxy         intPKEWorkflow.HTTPProxy
	Node
}

// Execute performs the activity
func (a DeleteNodeActivity) Execute(ctx context.Context, input DeleteNodeActivityInput) (existed bool, err error) {
	logger := activity.GetLogger(ctx).Sugar().With(
		"organization", input.OrganizationID,
		"cluster", input.ClusterName,
		"secret", input.SecretID,
		"node", input.Name,
	)

	/*keyvals := []interface{}{
		"cluster", input.ClusterName,
		"node", input.Node.Name,
	}*/

	c, err := a.vmomiClientFactory.New(input.OrganizationID, input.SecretID)
	if err = errors.WrapIf(err, "failed to create cloud connection"); err != nil {
		return true, err
	}

	expectedTags := getClusterTags(input.Name, input.NodePoolName)

	finder := find.NewFinder(c.Client)
	folder, err := finder.FolderOrDefault(ctx, input.FolderName)
	if err != nil {
		return vmRef, err
	}
	vms, err := finder.VirtualMachineList(ctx, input.Name)
	if err != nil {
		return false, errors.WrapIff(err, "couldn't find a VM named %q", input.Name)
	}

	for _, vm := range vms {
		config, err := vm.QueryConfigTarget()
	}

	template, err := finder.VirtualMachine(ctx, input.TemplateName)
	if err != nil {
		return vmRef, err
	}
	templateRef := template.Reference()

	pool, err := finder.ResourcePoolOrDefault(ctx, input.ResourcePoolName)
	if err != nil {
		return vmRef, err
	}

	poolRef := pool.Reference()
	cloneSpec.Location.Pool = &poolRef

	ds, err := finder.DatastoreOrDefault(ctx, input.DatastoreName)
	if err == nil {
		dsRef := ds.Reference()
		cloneSpec.Location.Datastore = &dsRef
	} else {
		if _, ok := err.(*find.NotFoundError); !ok {
			return vmRef, err
		}

		logger.Debugf("ds %s not found, fallback to drs", input.DatastoreName)
		storagePod, err := finder.DatastoreCluster(ctx, input.DatastoreName)
		if err != nil {
			if _, ok := err.(*find.NotFoundError); ok {
				return vmRef, fmt.Errorf("neither a datastore nor a datastore cluster named %q found", input.DatastoreName)
			}
			return vmRef, err
		}

		storagePodRef := storagePod.Reference()

		podSelectionSpec := types.StorageDrsPodSelectionSpec{
			StoragePod: &storagePodRef,
		}

		storagePlacementSpec := types.StoragePlacementSpec{
			Folder:           &folderRef,
			Vm:               &templateRef,
			CloneName:        input.Name,
			CloneSpec:        &cloneSpec,
			PodSelectionSpec: podSelectionSpec,
			Type:             string(types.StoragePlacementSpecPlacementTypeClone),
		}

		storageResourceManager := object.NewStorageResourceManager(c.Client)
		result, err := storageResourceManager.RecommendDatastores(ctx, storagePlacementSpec)
		if err != nil {
			return vmRef, err
		}

		if len(result.Recommendations) == 0 {
			return vmRef, fmt.Errorf("no datastore-cluster recommendations")
		}

		cloneSpec.Location.Datastore = &result.Recommendations[0].Action[0].(*types.StoragePlacementAction).Destination
		logger.Infof("deploying to %q datastore based on recommendation", cloneSpec.Location.Datastore)
	}

	task, err := template.Clone(ctx, folder, input.Name, cloneSpec)
	if err != nil {
		return vmRef, err
	}

	logger.Info("cloning template", "task", task.String())

	taskInfo, err := task.WaitForResult(ctx, nil)
	if err != nil {
		return vmRef, err
	}

	logger.Infof("vm deleted: %+v\n", taskInfo)

	if ref, ok := taskInfo.Result.(types.ManagedObjectReference); ok {
		vmRef = ref
	}
	return vmRef, nil
}

func encodeGuestinfo(data string) string {
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	compressor := gzip.NewWriter(encoder)

	compressor.Write([]byte(data))

	compressor.Close()
	encoder.Close()

	return buffer.String()
}

func generateCloudConfig(user, publicKey, script, hostname string) string {

	data := map[string]interface{}{
		"hostname":          hostname,
		"fqdn":              hostname,
		"preserve_hostname": false,
		"runcmd":            []string{script},
	}

	if publicKey != "" {
		if user == "" {
			user = "banzaicloud"
		}
		data["users"] = []map[string]interface{}{
			map[string]interface{}{
				"name":                user,
				"sudo":                "ALL=(ALL) NOPASSWD:ALL",
				"ssh-authorized-keys": []string{publicKey}}}
	}

	out, _ := yaml.Marshal(data)
	return "#cloud-config\n" + string(out)
}
