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

package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pipe-cd/pipe/pkg/config"
	"github.com/pipe-cd/pipe/pkg/model"
)

func TestMatch(t *testing.T) {
	testcases := []struct {
		name      string
		config    config.NotificationRoute
		matchings map[model.Event]bool
	}{
		{
			name:   "empty config",
			config: config.NotificationRoute{},
			matchings: map[model.Event]bool{
				{}: true,
				{Type: model.EventType_EVENT_DEPLOYMENT_TRIGGERED}: true,
			},
		},
		{
			name: "filter by event",
			config: config.NotificationRoute{
				Events: []string{
					"DEPLOYMENT_TRIGGERED",
				},
				IgnoreEvents: []string{
					"DEPLOYMENT_ROLLING_BACK",
				},
			},
			matchings: map[model.Event]bool{
				{
					Type: model.EventType_EVENT_DEPLOYMENT_TRIGGERED,
				}: true,
				{
					Type: model.EventType_EVENT_DEPLOYMENT_ROLLING_BACK,
				}: false,
			},
		},
		{
			name: "filter by group",
			config: config.NotificationRoute{
				Groups: []string{
					"DEPLOYMENT",
				},
				IgnoreGroups: []string{
					"APPLICATION",
				},
			},
			matchings: map[model.Event]bool{
				{
					Type: model.EventType_EVENT_DEPLOYMENT_TRIGGERED,
				}: true,
				{
					Type: model.EventType_EVENT_APPLICATION_SYNCED,
				}: false,
			},
		},
		{
			name: "filter by app",
			config: config.NotificationRoute{
				Apps: []string{
					"canary",
				},
				IgnoreApps: []string{
					"bluegreen",
				},
			},
			matchings: map[model.Event]bool{
				{
					Type: model.EventType_EVENT_DEPLOYMENT_TRIGGERED,
					Metadata: &model.EventDeploymentTriggered{
						Deployment: &model.Deployment{
							ApplicationId: "canary",
						},
					},
				}: true,
				{
					Type: model.EventType_EVENT_DEPLOYMENT_PLANNED,
					Metadata: &model.EventDeploymentTriggered{
						Deployment: &model.Deployment{
							ApplicationId: "bluegreen",
						},
					},
				}: false,
				{
					Type: model.EventType_EVENT_DEPLOYMENT_SUCCEEDED,
					Metadata: &model.EventDeploymentTriggered{
						Deployment: &model.Deployment{
							ApplicationId: "not-specified",
						},
					},
				}: false,
				{
					Type:     model.EventType_EVENT_PIPED_STARTED,
					Metadata: &model.EventPipedStarted{},
				}: true,
			},
		},
		{
			name: "filter by env",
			config: config.NotificationRoute{
				Envs: []string{
					"prod",
				},
				IgnoreEnvs: []string{
					"dev",
				},
			},
			matchings: map[model.Event]bool{
				{
					Type: model.EventType_EVENT_DEPLOYMENT_TRIGGERED,
					Metadata: &model.EventDeploymentTriggered{
						Deployment: &model.Deployment{
							EnvId: "prod",
						},
					},
				}: true,
				{
					Type: model.EventType_EVENT_DEPLOYMENT_PLANNED,
					Metadata: &model.EventDeploymentTriggered{
						Deployment: &model.Deployment{
							EnvId: "dev",
						},
					},
				}: false,
				{
					Type: model.EventType_EVENT_DEPLOYMENT_SUCCEEDED,
					Metadata: &model.EventDeploymentTriggered{
						Deployment: &model.Deployment{
							EnvId: "not-specified",
						},
					},
				}: false,
				{
					Type:     model.EventType_EVENT_PIPED_STARTED,
					Metadata: &model.EventPipedStarted{},
				}: true,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			matcher := newMatcher(tc.config)
			for event, expected := range tc.matchings {
				got := matcher.Match(event)
				assert.Equal(t, expected, got, event.Type.String())
			}
		})
	}
}
