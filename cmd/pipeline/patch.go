// Copyright (c) 2018, Google, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/spf13/cobra"
	"github.com/spinnaker/spin/cmd/gateclient"
	"github.com/spinnaker/spin/util"
)

type PatchOptions struct {
	*pipelineOptions
	application string
	name        string
	patch       string
}

var (
	patchPipelineShort = "Patches the specified pipeline definition"
	patchPipelineLong  = "Patches the specified pipeline definition"
)

func NewPatchCmd(pipelineOptions pipelineOptions) *cobra.Command {
	options := PatchOptions{
		pipelineOptions: &pipelineOptions,
	}
	cmd := &cobra.Command{
		Use:   "patch",
		Short: patchPipelineShort,
		Long:  patchPipelineLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			return patchPipeline(cmd, options)
		},
	}

	cmd.PersistentFlags().StringVarP(&options.application, "application", "a", "", "Spinnaker application the pipeline belongs to")
	cmd.PersistentFlags().StringVarP(&options.name, "name", "n", "", "name of the pipeline")
	cmd.PersistentFlags().StringVarP(&options.patch, "patch", "p", "", "patch value in json")

	return cmd
}

func patchPipeline(cmd *cobra.Command, options PatchOptions) error {
	gateClient, err := gateclient.NewGateClient(cmd.InheritedFlags())
	if err != nil {
		return err
	}

	if options.application == "" || options.name == "" {
		return errors.New("one of required parameters 'application' or 'name' not set")
	}

	if options.patch == "" {
		return errors.New("patch value must be set")
	}

	// Get pipeline
	pipeline, err := loadPipelineJSON(gateClient, options.application, options.name)
	if err != nil {
		return nil
	}

	patchedPipelineBytes, err := jsonpatch.MergePatch(pipeline, []byte(options.patch))
	if err != nil {
		return err
	}

	patchedPipeline := make(map[string]interface{})
	json.Unmarshal(patchedPipelineBytes, &patchedPipeline)

	util.UI.JsonOutput(patchedPipeline, util.UI.OutputFormat)
	return nil
}

func loadPipelineJSON(gateClient *gateclient.GatewayClient, app string, name string) ([]byte, error) {
	successPayload, resp, err := gateClient.ApplicationControllerApi.GetPipelineConfigUsingGET(gateClient.Context,
		app,
		name)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Encountered an error getting pipeline in pipeline %s with name %s, status code: %d\n",
			app,
			name,
			resp.StatusCode)
	}

	pipelineJSON, err := json.Marshal(successPayload)
	if err != nil {
		return nil, err
	}

	return pipelineJSON, nil
}
