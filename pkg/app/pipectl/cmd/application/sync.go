// Copyright 2020 The PipeCD Authors.
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

package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pipe-cd/pipe/pkg/app/pipectl/client"
	"github.com/pipe-cd/pipe/pkg/cli"
	"github.com/pipe-cd/pipe/pkg/model"
)

type sync struct {
	root *command

	appID         string
	status        []string
	checkInterval time.Duration
	timeout       time.Duration
}

func newSyncCommand(root *command) *cobra.Command {
	c := &sync{
		root:          root,
		checkInterval: 15 * time.Second,
		timeout:       5 * time.Minute,
	}
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync an application.",
		RunE:  cli.WithContext(c.run),
	}

	cmd.Flags().StringVar(&c.appID, "app-id", c.appID, "The application ID.")
	cmd.Flags().StringSliceVar(&c.status, "wait-status", c.status, fmt.Sprintf("The list of waiting statuses. Empty means returning immediately after triggered. (%s)", strings.Join(availableStatuses(), "|")))
	cmd.Flags().DurationVar(&c.checkInterval, "check-interval", c.checkInterval, "The interval of checking the requested command.")
	cmd.Flags().DurationVar(&c.timeout, "timeout", c.timeout, "Maximum execution time.")

	cmd.MarkFlagRequired("app-id")

	return cmd
}

func (c *sync) run(ctx context.Context, t cli.Telemetry) error {
	statuses, err := makeStatuses(c.status)
	if err != nil {
		return fmt.Errorf("invalid deployment status: %w", err)
	}

	cli, err := c.root.clientOptions.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}
	defer cli.Close()

	deploymentID, err := client.SyncApplication(ctx, cli, c.appID, c.checkInterval, c.timeout, t.Logger)
	if err != nil {
		return err
	}

	t.Logger.Info(fmt.Sprintf("Successfully triggered deployment %s", deploymentID))
	if len(statuses) == 0 {
		return nil
	}

	t.Logger.Info("Waiting until the deployment reaches one of the specified statuses")

	return client.WaitDeploymentStatuses(
		ctx,
		cli,
		deploymentID,
		statuses,
		c.checkInterval,
		c.timeout,
		t.Logger,
	)
}

func makeStatuses(statuses []string) ([]model.DeploymentStatus, error) {
	out := make([]model.DeploymentStatus, 0, len(statuses))
	for _, s := range statuses {
		status, ok := model.DeploymentStatus_value["DEPLOYMENT_"+s]
		if !ok {
			return nil, fmt.Errorf("bad status %s", s)
		}
		out = append(out, model.DeploymentStatus(status))
	}
	return out, nil
}

func availableStatuses() []string {
	out := make([]string, 0, len(model.DeploymentStatus_value))
	for s := range model.DeploymentStatus_value {
		out = append(out, strings.TrimPrefix(s, "DEPLOYMENT_"))
	}
	return out
}