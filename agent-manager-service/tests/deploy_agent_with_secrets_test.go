// Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wso2/agent-manager/agent-manager-service/middleware/jwtassertion"
	"github.com/wso2/agent-manager/agent-manager-service/spec"
	"github.com/wso2/agent-manager/agent-manager-service/tests/apitestutils"
	"github.com/wso2/agent-manager/agent-manager-service/wiring"
)

var (
	deploySecretTestOrgName   = fmt.Sprintf("deploy-secret-test-org-%s", uuid.New().String()[:5])
	deploySecretTestProjName  = fmt.Sprintf("deploy-secret-test-proj-%s", uuid.New().String()[:5])
	deploySecretTestAgentName = fmt.Sprintf("deploy-secret-test-agent-%s", uuid.New().String()[:5])
)

func TestDeployAgentWithSecrets(t *testing.T) {
	authMiddleware := jwtassertion.NewMockMiddleware(t)

	t.Run("Deploying agent with sensitive env vars should return 202", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.ComponentExistsFunc = func(ctx context.Context, orgName string, projName string, agentName string, verifyProject bool) (bool, error) {
			return true, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		// Create the request body with sensitive environment variables
		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
			"imageId": "registry.example.com/myapp:v2.0.0",
			"env": []map[string]interface{}{
				{
					"key":   "DATABASE_URL",
					"value": "postgresql://localhost:5432/mydb",
				},
				{
					"key":   "SECRET_API_KEY",
					"value": "sk-secret-key-for-deploy",
				},
				{
					"key":   "LOG_LEVEL",
					"value": "DEBUG",
				},
			},
		})
		require.NoError(t, err)

		// Send the request
		url := fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents/%s/deployments",
			deploySecretTestOrgName, deploySecretTestProjName, deploySecretTestAgentName)
		req := httptest.NewRequest(http.MethodPost, url, reqBody)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		// Assert response
		require.Equal(t, http.StatusAccepted, rr.Code)

		// Read and validate response body
		b, err := io.ReadAll(rr.Body)
		require.NoError(t, err)
		t.Logf("response body: %s", string(b))

		var response spec.DeploymentResponse
		require.NoError(t, json.Unmarshal(b, &response))

		// Validate response fields
		require.Equal(t, deploySecretTestAgentName, response.AgentName)
		require.Equal(t, deploySecretTestProjName, response.ProjectName)
		require.Equal(t, "registry.example.com/myapp:v2.0.0", response.ImageId)
		require.Equal(t, "Development", response.Environment)

		// Validate service calls
		require.Len(t, openChoreoClient.DeployCalls(), 1)

		// Validate call parameters including env vars
		deployCall := openChoreoClient.DeployCalls()[0]
		require.Equal(t, deploySecretTestOrgName, deployCall.NamespaceName)
		require.Equal(t, deploySecretTestProjName, deployCall.ProjectName)
		require.Equal(t, deploySecretTestAgentName, deployCall.ComponentName)
		require.Equal(t, "registry.example.com/myapp:v2.0.0", deployCall.Req.ImageID)

		// Validate environment variables were passed
		require.Len(t, deployCall.Req.Env, 3)
		require.Equal(t, "DATABASE_URL", deployCall.Req.Env[0].Key)
		require.Equal(t, "postgresql://localhost:5432/mydb", deployCall.Req.Env[0].Value)
		require.Equal(t, "SECRET_API_KEY", deployCall.Req.Env[1].Key)
		require.Equal(t, "sk-secret-key-for-deploy", deployCall.Req.Env[1].Value)
		require.Equal(t, "LOG_LEVEL", deployCall.Req.Env[2].Key)
		require.Equal(t, "DEBUG", deployCall.Req.Env[2].Value)
	})

	t.Run("Deploying agent with empty env list should return 202", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.ComponentExistsFunc = func(ctx context.Context, orgName string, projName string, agentName string, verifyProject bool) (bool, error) {
			return true, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
			"imageId": "registry.example.com/myapp:v3.0.0",
			"env":     []map[string]interface{}{},
		})
		require.NoError(t, err)

		url := fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents/%s/deployments",
			deploySecretTestOrgName, deploySecretTestProjName, deploySecretTestAgentName)
		req := httptest.NewRequest(http.MethodPost, url, reqBody)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)

		// Validate that deploy was called with empty env
		require.Len(t, openChoreoClient.DeployCalls(), 1)
		deployCall := openChoreoClient.DeployCalls()[0]
		require.Empty(t, deployCall.Req.Env)
	})

	t.Run("Deploying agent that does not exist should return 404", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.ComponentExistsFunc = func(ctx context.Context, orgName string, projName string, agentName string, verifyProject bool) (bool, error) {
			if agentName == "nonexistent-agent-xyz" {
				return false, nil
			}
			return true, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
			"imageId": "registry.example.com/myapp:v1.0.0",
			"env": []map[string]interface{}{
				{
					"key":   "SECRET_KEY",
					"value": "should-not-reach-deployment",
				},
			},
		})
		require.NoError(t, err)

		url := fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents/nonexistent-agent-xyz/deployments",
			deploySecretTestOrgName, deploySecretTestProjName)
		req := httptest.NewRequest(http.MethodPost, url, reqBody)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		// Agent not found should return 404
		require.Equal(t, http.StatusNotFound, rr.Code)
	})
}
