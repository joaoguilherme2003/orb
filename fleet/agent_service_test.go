// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Adapted for Orb project, modifications licensed under MPL v. 2.0:
/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package fleet_test

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ns1labs/orb/fleet"
	flmocks "github.com/ns1labs/orb/fleet/mocks"
	"github.com/ns1labs/orb/pkg/errors"
	"github.com/ns1labs/orb/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	agent = fleet.Agent{
		Name:          types.Identifier{},
		MFOwnerID:     "",
		MFThingID:     "",
		MFKeyID:       "",
		MFChannelID:   "",
		Created:       time.Time{},
		OrbTags:       nil,
		AgentTags:     nil,
		AgentMetadata: nil,
		State:         0,
		LastHBData:    nil,
		LastHB:        time.Time{},
	}
)

func TestUpdateAgent(t *testing.T) {
	users := flmocks.NewAuthService(map[string]string{token: email})

	thingsServer := newThingsServer(newThingsService(users))
	fleetService := newService(users, thingsServer.URL)

	ag, err := createAgent(t, "my-agent1", fleetService)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	wrongAgentGroup := fleet.Agent{MFThingID: wrongID}
	cases := map[string]struct {
		group fleet.Agent
		token string
		err   error
	}{
		"update existing agent": {
			group: ag,
			token: token,
			err:   nil,
		},
		"update group with wrong credentials": {
			group: ag,
			token: invalidToken,
			err:   fleet.ErrUnauthorizedAccess,
		},
		"update a non-existing group": {
			group: wrongAgentGroup,
			token: token,
			err:   fleet.ErrNotFound,
		},
	}

	for desc, tc := range cases {
		_, err := fleetService.EditAgent(context.Background(), tc.token, tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %d got %d", desc, tc.err, err))
	}
}


func TestCreateAgent(t *testing.T) {
	users := flmocks.NewAuthService(map[string]string{token: email})
	thingsServer := newThingsServer(newThingsService(users))
	fleetService := newService(users, thingsServer.URL)

	ownerID, err := uuid.NewV4()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	nameID, err := types.NewIdentifier("eu-agents")
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	validAgent := fleet.Agent{
		MFOwnerID:   ownerID.String(),
		Name:        nameID,
		OrbTags:     make(map[string]string),
		Created:     time.Time{},
	}
	validAgent.OrbTags = map[string]string{
		"region":    "eu",
		"node_type": "dns",
	}
	cases := map[string]struct {
		agent fleet.Agent
		token string
		err   error
	}{
		"add a valid agent": {
			agent: validAgent,
			token: token,
			err:   nil,
		},
		"add a valid agent with an invalid token": {
			agent: validAgent,
			token: invalidToken,
			err:   fleet.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		_, err := fleetService.CreateAgent(context.Background(), tc.token, tc.agent)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s", desc, tc.err, err))
	}
}

func TestValidateAgent(t *testing.T) {
	users := flmocks.NewAuthService(map[string]string{token: email})
	thingsServer := newThingsServer(newThingsService(users))
	fleetService := newService(users, thingsServer.URL)

	ownerID, err := uuid.NewV4()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	nameID, err := types.NewIdentifier("eu-agents")
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	validAgent := fleet.Agent{
		MFOwnerID:   ownerID.String(),
		Name:        nameID,
		OrbTags:     make(map[string]string),
	}
	validAgent.OrbTags = map[string]string{
		"region":    "eu",
		"node_type": "dns",
	}
	cases := map[string]struct {
		agent fleet.Agent
		token string
		err   error
	}{
		"validate a valid agent": {
			agent: validAgent,
			token: token,
			err:   nil,
		},
		"validate a valid agent with an invalid token": {
			agent: validAgent,
			token: invalidToken,
			err:   fleet.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		_, err := fleetService.ValidateAgent(context.Background(), tc.token, tc.agent)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s", desc, tc.err, err))
	}
}

func createAgent(t *testing.T, name string, svc fleet.Service) (fleet.Agent, error) {
	t.Helper()
	aCopy := agent
	validName, err := types.NewIdentifier(name)
	if err != nil {
		return fleet.Agent{}, err
	}
	aCopy.Name = validName
	ag, err := svc.CreateAgent(context.Background(), token, agent)
	if err != nil {
		return fleet.Agent{}, err
	}
	return ag, nil
}
