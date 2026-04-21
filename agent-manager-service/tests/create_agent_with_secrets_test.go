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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wso2/agent-manager/agent-manager-service/clients/clientmocks"
	"github.com/wso2/agent-manager/agent-manager-service/clients/openchoreosvc/client"
	"github.com/wso2/agent-manager/agent-manager-service/middleware/jwtassertion"
	"github.com/wso2/agent-manager/agent-manager-service/models"
	"github.com/wso2/agent-manager/agent-manager-service/spec"
	"github.com/wso2/agent-manager/agent-manager-service/tests/apitestutils"
	"github.com/wso2/agent-manager/agent-manager-service/wiring"
)

var (
	secretAgentTestOrgName  = fmt.Sprintf("secret-agent-test-org-%s", uuid.New().String()[:5])
	secretAgentTestProjName = fmt.Sprintf("secret-agent-test-proj-%s", uuid.New().String()[:5])
)

func TestCreateAgentWithSecretRef(t *testing.T) {
	authMiddleware := jwtassertion.NewMockMiddleware(t)

	t.Run("Creating an agent with private repo secretRef should return 202", func(t *testing.T) {
		testAgentName := fmt.Sprintf("private-repo-agent-%s", uuid.New().String()[:5])
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()

		// Override GetComponentFunc to return valid component for token generation
		openChoreoClient.GetComponentFunc = func(ctx context.Context, namespaceName, projectName, componentName string) (*models.AgentResponse, error) {
			return &models.AgentResponse{
				UUID:        uuid.New().String(),
				Name:        componentName,
				ProjectName: projectName,
				Provisioning: models.Provisioning{
					Type: "internal",
				},
				CreatedAt: time.Now(),
			}, nil
		}

		// Mock ListGitSecrets to return the secret that matches the secretRef
		openChoreoClient.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
			return []*client.GitSecretInfo{
				{Name: "my-private-repo-secret"},
				{Name: "another-secret"},
			}, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		// Create the request body with secretRef for private repository
		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
			"name":        testAgentName,
			"displayName": "Private Repo Agent",
			"description": "Agent from a private repository",
			"provisioning": map[string]interface{}{
				"type": "internal",
				"repository": map[string]interface{}{
					"url":       "https://github.com/private-org/private-repo",
					"branch":    "main",
					"appPath":   "/agent-sample",
					"secretRef": "my-private-repo-secret",
				},
			},
			"build": map[string]interface{}{
				"type": "buildpack",
				"buildpack": map[string]interface{}{
					"language":        "python",
					"languageVersion": "3.11",
					"runCommand":      "uvicorn app:app --host 0.0.0.0 --port 8000",
				},
			},
			"agentType": map[string]interface{}{
				"type":    "agent-api",
				"subType": "chat-api",
			},
			"inputInterface": map[string]interface{}{
				"type": "HTTP",
			},
		})
		require.NoError(t, err)

		// Send the request
		url := fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents", secretAgentTestOrgName, secretAgentTestProjName)
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

		var payload spec.AgentResponse
		require.NoError(t, json.Unmarshal(b, &payload))

		// Validate response fields
		require.Equal(t, testAgentName, payload.Name)
		require.Equal(t, "Agent from a private repository", payload.Description)
		require.Equal(t, secretAgentTestProjName, payload.ProjectName)
		require.NotZero(t, payload.CreatedAt)

		// Validate that ListGitSecrets was called to validate the secretRef
		require.Len(t, openChoreoClient.ListGitSecretsCalls(), 1)
		require.Equal(t, secretAgentTestOrgName, openChoreoClient.ListGitSecretsCalls()[0].NamespaceName)

		// Validate service calls
		require.Len(t, openChoreoClient.CreateComponentCalls(), 1)

		// Validate that the secretRef was passed to CreateComponent
		createCall := openChoreoClient.CreateComponentCalls()[0]
		require.Equal(t, secretAgentTestOrgName, createCall.NamespaceName)
		require.Equal(t, secretAgentTestProjName, createCall.ProjectName)
		require.Equal(t, testAgentName, createCall.Req.Name)
		require.NotNil(t, createCall.Req.Repository)
		require.Equal(t, "my-private-repo-secret", createCall.Req.Repository.SecretRef)
	})

	t.Run("Creating an agent with sensitive env vars should return 202", func(t *testing.T) {
		testAgentName := fmt.Sprintf("sensitive-env-agent-%s", uuid.New().String()[:5])
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		secretMgmtClient := apitestutils.CreateMockSecretManagementClient()

		openChoreoClient.GetComponentFunc = func(ctx context.Context, namespaceName, projectName, componentName string) (*models.AgentResponse, error) {
			return &models.AgentResponse{
				UUID:        uuid.New().String(),
				Name:        componentName,
				ProjectName: projectName,
				Provisioning: models.Provisioning{
					Type: "internal",
				},
				CreatedAt: time.Now(),
			}, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
			SecretMgmtClient: secretMgmtClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		// Create the request body with sensitive environment variables
		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
			"name":        testAgentName,
			"displayName": "Agent With Secrets",
			"description": "Agent with sensitive environment variables",
			"provisioning": map[string]interface{}{
				"type": "internal",
				"repository": map[string]interface{}{
					"url":     "https://github.com/test/test-repo",
					"branch":  "main",
					"appPath": "/agent-sample",
				},
			},
			"build": map[string]interface{}{
				"type": "buildpack",
				"buildpack": map[string]interface{}{
					"language":        "python",
					"languageVersion": "3.11",
					"runCommand":      "uvicorn app:app --host 0.0.0.0 --port 8000",
				},
			},
			"configurations": map[string]interface{}{
				"env": []map[string]interface{}{
					{
						"key":         "DB_HOST",
						"value":       "localhost",
						"isSensitive": false,
					},
					{
						"key":         "DB_PASSWORD",
						"value":       "super-secret-password",
						"isSensitive": true,
					},
					{
						"key":         "API_SECRET_KEY",
						"value":       "sk-secret-key-12345",
						"isSensitive": true,
					},
				},
			},
			"agentType": map[string]interface{}{
				"type":    "agent-api",
				"subType": "chat-api",
			},
			"inputInterface": map[string]interface{}{
				"type": "HTTP",
			},
		})
		require.NoError(t, err)

		// Send the request
		url := fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents", secretAgentTestOrgName, secretAgentTestProjName)
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

		var payload spec.AgentResponse
		require.NoError(t, json.Unmarshal(b, &payload))

		require.Equal(t, testAgentName, payload.Name)
		require.NotZero(t, payload.CreatedAt)

		// Validate service calls
		require.Len(t, openChoreoClient.CreateComponentCalls(), 1)

		// Validate that configurations were passed correctly
		createCall := openChoreoClient.CreateComponentCalls()[0]
		require.NotNil(t, createCall.Req.Configurations)
		require.NotEmpty(t, createCall.Req.Configurations.Env)

		// Sensitive vars should use secretKeyRef, non-sensitive should have direct value
		foundPlain := false
		foundSecretRef := false
		for _, env := range createCall.Req.Configurations.Env {
			if env.Key == "DB_HOST" {
				foundPlain = true
				require.Equal(t, "localhost", env.Value)
				require.Nil(t, env.ValueFrom)
			}
			if env.Key == "DB_PASSWORD" || env.Key == "API_SECRET_KEY" {
				foundSecretRef = true
				require.NotNil(t, env.ValueFrom)
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				require.Equal(t, env.Key, env.ValueFrom.SecretKeyRef.Key)
			}
		}
		require.True(t, foundPlain, "Should have found plain env var DB_HOST")
		require.True(t, foundSecretRef, "Should have found secret ref env vars")

		// Validate that CreateSecret was called to store sensitive values
		require.NotEmpty(t, secretMgmtClient.CreateSecretCalls())
	})

	validationTests := []struct {
		name           string
		authMiddleware jwtassertion.Middleware
		payload        map[string]interface{}
		wantStatus     int
		wantErrMsg     string
		url            string
		setupMock      func() (*clientmocks.OpenChoreoClientMock, *clientmocks.SecretManagementClientMock)
	}{
		{
			name:           "return 404 when secretRef does not exist",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name":        fmt.Sprintf("test-agent-%s", uuid.New().String()[:5]),
				"displayName": "Test Agent",
				"description": "Test description",
				"provisioning": map[string]interface{}{
					"type": "internal",
					"repository": map[string]interface{}{
						"url":       "https://github.com/private-org/private-repo",
						"branch":    "main",
						"appPath":   "/agent-sample",
						"secretRef": "nonexistent-secret",
					},
				},
				"build": map[string]interface{}{
					"type": "buildpack",
					"buildpack": map[string]interface{}{
						"language":        "python",
						"languageVersion": "3.11",
						"runCommand":      "uvicorn app:app --host 0.0.0.0 --port 8000",
					},
				},
				"agentType": map[string]interface{}{
					"type":    "agent-api",
					"subType": "chat-api",
				},
				"inputInterface": map[string]interface{}{
					"type": "HTTP",
				},
			},
			wantStatus: 404,
			wantErrMsg: "Git secret not found",
			url:        fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents", secretAgentTestOrgName, secretAgentTestProjName),
			setupMock: func() (*clientmocks.OpenChoreoClientMock, *clientmocks.SecretManagementClientMock) {
				mock := apitestutils.CreateMockOpenChoreoClient()
				// Return secrets that don't include the requested secretRef
				mock.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
					return []*client.GitSecretInfo{
						{Name: "some-other-secret"},
					}, nil
				}
				return mock, apitestutils.CreateMockSecretManagementClient()
			},
		},
		{
			name:           "return 500 when git secret validation fails with service error",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name":        fmt.Sprintf("test-agent-%s", uuid.New().String()[:5]),
				"displayName": "Test Agent",
				"description": "Test description",
				"provisioning": map[string]interface{}{
					"type": "internal",
					"repository": map[string]interface{}{
						"url":       "https://github.com/private-org/private-repo",
						"branch":    "main",
						"appPath":   "/agent-sample",
						"secretRef": "my-secret",
					},
				},
				"build": map[string]interface{}{
					"type": "buildpack",
					"buildpack": map[string]interface{}{
						"language":        "python",
						"languageVersion": "3.11",
						"runCommand":      "uvicorn app:app --host 0.0.0.0 --port 8000",
					},
				},
				"agentType": map[string]interface{}{
					"type":    "agent-api",
					"subType": "chat-api",
				},
				"inputInterface": map[string]interface{}{
					"type": "HTTP",
				},
			},
			wantStatus: 500,
			wantErrMsg: "Failed to create agent",
			url:        fmt.Sprintf("/api/v1/orgs/%s/projects/%s/agents", secretAgentTestOrgName, secretAgentTestProjName),
			setupMock: func() (*clientmocks.OpenChoreoClientMock, *clientmocks.SecretManagementClientMock) {
				mock := apitestutils.CreateMockOpenChoreoClient()
				// ListGitSecrets fails with an error
				mock.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
					return nil, fmt.Errorf("failed to connect to OpenChoreo")
				}
				return mock, apitestutils.CreateMockSecretManagementClient()
			},
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			openChoreoClient, secretMgmtClient := tt.setupMock()
			testClients := wiring.TestClients{
				OpenChoreoClient: openChoreoClient,
				SecretMgmtClient: secretMgmtClient,
			}

			app := apitestutils.MakeAppClientWithDeps(t, testClients, tt.authMiddleware)

			reqBody := new(bytes.Buffer)
			err := json.NewEncoder(reqBody).Encode(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, tt.url, reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			app.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)

			body, err := io.ReadAll(rr.Body)
			require.NoError(t, err)

			if tt.wantStatus >= 400 {
				bodyStr := string(body)
				require.Contains(t, bodyStr, tt.wantErrMsg)
			}
		})
	}
}
