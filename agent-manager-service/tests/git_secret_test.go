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

	"github.com/wso2/agent-manager/agent-manager-service/clients/clientmocks"
	"github.com/wso2/agent-manager/agent-manager-service/clients/openchoreosvc/client"
	"github.com/wso2/agent-manager/agent-manager-service/middleware/jwtassertion"
	"github.com/wso2/agent-manager/agent-manager-service/spec"
	"github.com/wso2/agent-manager/agent-manager-service/tests/apitestutils"
	"github.com/wso2/agent-manager/agent-manager-service/utils"
	"github.com/wso2/agent-manager/agent-manager-service/wiring"
)

var (
	gitSecretTestOrgName = fmt.Sprintf("git-secret-test-org-%s", uuid.New().String()[:5])
)

func TestCreateGitSecret(t *testing.T) {
	authMiddleware := jwtassertion.NewMockMiddleware(t)

	t.Run("Creating a git secret with valid basic-auth should return 201", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
			return &client.GitSecretInfo{
				Name: req.Name,
			}, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		// Create the request body
		reqBody := new(bytes.Buffer)
		err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
			"name": "my-git-secret",
			"type": "basic-auth",
			"credentials": map[string]interface{}{
				"username": "git-user",
				"password": "ghp_test-token-12345",
			},
		})
		require.NoError(t, err)

		// Send the request
		url := fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName)
		req := httptest.NewRequest(http.MethodPost, url, reqBody)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		// Assert response
		require.Equal(t, http.StatusCreated, rr.Code)

		// Read and validate response body
		b, err := io.ReadAll(rr.Body)
		require.NoError(t, err)
		t.Logf("response body: %s", string(b))

		var payload spec.GitSecretResponse
		require.NoError(t, json.Unmarshal(b, &payload))

		// Validate response fields
		require.Equal(t, "my-git-secret", payload.Name)

		// Validate service calls
		require.Len(t, openChoreoClient.CreateGitSecretCalls(), 1)

		// Validate call parameters
		createCall := openChoreoClient.CreateGitSecretCalls()[0]
		require.Equal(t, gitSecretTestOrgName, createCall.NamespaceName)
		require.Equal(t, "my-git-secret", createCall.Req.Name)
		require.Equal(t, "git-user", createCall.Req.Username)
		require.Equal(t, "ghp_test-token-12345", createCall.Req.Token)
	})

	validationTests := []struct {
		name           string
		authMiddleware jwtassertion.Middleware
		payload        map[string]interface{}
		wantStatus     int
		wantErrMsg     string
		url            string
		setupMock      func() *clientmocks.OpenChoreoClientMock
	}{
		{
			name:           "return 400 on missing secret name",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"username": "git-user",
					"password": "ghp_token",
				},
			},
			wantStatus: 400,
			wantErrMsg: "Git secret name cannot be empty",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return &client.GitSecretInfo{Name: req.Name}, nil
				}
				return mock
			},
		},
		{
			name:           "return 400 on invalid secret name with special characters",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name": "Invalid_Secret!",
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"username": "git-user",
					"password": "ghp_token",
				},
			},
			wantStatus: 400,
			wantErrMsg: "Git secret name must contain only lowercase alphanumeric characters or '-'",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return &client.GitSecretInfo{Name: req.Name}, nil
				}
				return mock
			},
		},
		{
			name:           "return 400 on invalid secret type",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name": "my-secret",
				"type": "ssh-key",
				"credentials": map[string]interface{}{
					"username": "git-user",
					"password": "ghp_token",
				},
			},
			wantStatus: 400,
			wantErrMsg: "git secret type must be 'basic-auth'",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return &client.GitSecretInfo{Name: req.Name}, nil
				}
				return mock
			},
		},
		{
			name:           "return 400 on missing username",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name": "my-secret",
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"password": "ghp_token",
				},
			},
			wantStatus: 400,
			wantErrMsg: "username is required for basic-auth type",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return &client.GitSecretInfo{Name: req.Name}, nil
				}
				return mock
			},
		},
		{
			name:           "return 400 on missing password",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name": "my-secret",
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"username": "git-user",
				},
			},
			wantStatus: 400,
			wantErrMsg: "password is required for basic-auth type",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return &client.GitSecretInfo{Name: req.Name}, nil
				}
				return mock
			},
		},
		{
			name:           "return 409 on git secret already exists",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name": "existing-secret",
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"username": "git-user",
					"password": "ghp_token",
				},
			},
			wantStatus: 409,
			wantErrMsg: "Git secret already exists",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return nil, utils.ErrConflict
				}
				return mock
			},
		},
		{
			name:           "return 500 on service error",
			authMiddleware: authMiddleware,
			payload: map[string]interface{}{
				"name": "my-secret",
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"username": "git-user",
					"password": "ghp_token",
				},
			},
			wantStatus: 500,
			wantErrMsg: "Failed to create git secret",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return nil, fmt.Errorf("internal service error")
				}
				return mock
			},
		},
		{
			name: "return 401 on missing authentication",
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing header: Authorization")
				})
			},
			payload: map[string]interface{}{
				"name": "my-secret",
				"type": "basic-auth",
				"credentials": map[string]interface{}{
					"username": "git-user",
					"password": "ghp_token",
				},
			},
			wantStatus: 401,
			wantErrMsg: "missing header: Authorization",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.CreateGitSecretFunc = func(ctx context.Context, namespaceName string, req client.CreateGitSecretRequest) (*client.GitSecretInfo, error) {
					return &client.GitSecretInfo{Name: req.Name}, nil
				}
				return mock
			},
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			openChoreoClient := tt.setupMock()
			testClients := wiring.TestClients{
				OpenChoreoClient: openChoreoClient,
			}

			app := apitestutils.MakeAppClientWithDeps(t, testClients, tt.authMiddleware)

			reqBody := new(bytes.Buffer)
			err := json.NewEncoder(reqBody).Encode(tt.payload)
			require.NoError(t, err)

			// Send the request
			req := httptest.NewRequest(http.MethodPost, tt.url, reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			app.ServeHTTP(rr, req)

			// Assert response
			require.Equal(t, tt.wantStatus, rr.Code)

			// Read response body and check error message
			body, err := io.ReadAll(rr.Body)
			require.NoError(t, err)

			if tt.wantStatus >= 400 {
				bodyStr := string(body)
				require.Contains(t, bodyStr, tt.wantErrMsg)
			}
		})
	}
}

func TestListGitSecrets(t *testing.T) {
	authMiddleware := jwtassertion.NewMockMiddleware(t)

	t.Run("Listing git secrets should return 200 with secrets", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
			return []*client.GitSecretInfo{
				{Name: "secret-one"},
				{Name: "secret-two"},
				{Name: "secret-three"},
			}, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		// Send the request
		url := fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName)
		req := httptest.NewRequest(http.MethodGet, url, nil)

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		// Assert response
		require.Equal(t, http.StatusOK, rr.Code)

		// Read and validate response body
		b, err := io.ReadAll(rr.Body)
		require.NoError(t, err)
		t.Logf("response body: %s", string(b))

		var payload spec.GitSecretListResponse
		require.NoError(t, json.Unmarshal(b, &payload))

		// Validate response fields
		require.Len(t, payload.Secrets, 3)
		require.Equal(t, "secret-one", payload.Secrets[0].Name)
		require.Equal(t, "secret-two", payload.Secrets[1].Name)
		require.Equal(t, "secret-three", payload.Secrets[2].Name)
		require.Equal(t, int32(3), payload.Total)

		// Validate service calls
		require.Len(t, openChoreoClient.ListGitSecretsCalls(), 1)
		require.Equal(t, gitSecretTestOrgName, openChoreoClient.ListGitSecretsCalls()[0].NamespaceName)
	})

	t.Run("Listing git secrets with no secrets should return empty list", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
			return []*client.GitSecretInfo{}, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		url := fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName)
		req := httptest.NewRequest(http.MethodGet, url, nil)

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		b, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var payload spec.GitSecretListResponse
		require.NoError(t, json.Unmarshal(b, &payload))

		require.Len(t, payload.Secrets, 0)
		require.Equal(t, int32(0), payload.Total)
	})

	t.Run("Listing git secrets with pagination should return correct page", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
			secrets := make([]*client.GitSecretInfo, 5)
			for i := 0; i < 5; i++ {
				secrets[i] = &client.GitSecretInfo{Name: fmt.Sprintf("secret-%d", i)}
			}
			return secrets, nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		// Request with limit=2 and offset=1
		url := fmt.Sprintf("/api/v1/orgs/%s/git-secrets?limit=2&offset=1", gitSecretTestOrgName)
		req := httptest.NewRequest(http.MethodGet, url, nil)

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		b, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var payload spec.GitSecretListResponse
		require.NoError(t, json.Unmarshal(b, &payload))

		require.Len(t, payload.Secrets, 2)
		require.Equal(t, "secret-1", payload.Secrets[0].Name)
		require.Equal(t, "secret-2", payload.Secrets[1].Name)
		require.Equal(t, int32(5), payload.Total)
		require.Equal(t, int32(2), payload.Limit)
		require.Equal(t, int32(1), payload.Offset)
	})

	validationTests := []struct {
		name           string
		authMiddleware jwtassertion.Middleware
		wantStatus     int
		wantErrMsg     string
		url            string
		setupMock      func() *clientmocks.OpenChoreoClientMock
	}{
		{
			name:           "return 500 on service error",
			authMiddleware: authMiddleware,
			wantStatus:     500,
			wantErrMsg:     "Failed to list git secrets",
			url:            fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
					return nil, fmt.Errorf("internal service error")
				}
				return mock
			},
		},
		{
			name: "return 401 on missing authentication",
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing header: Authorization")
				})
			},
			wantStatus: 401,
			wantErrMsg: "missing header: Authorization",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.ListGitSecretsFunc = func(ctx context.Context, namespaceName string) ([]*client.GitSecretInfo, error) {
					return []*client.GitSecretInfo{}, nil
				}
				return mock
			},
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			openChoreoClient := tt.setupMock()
			testClients := wiring.TestClients{
				OpenChoreoClient: openChoreoClient,
			}

			app := apitestutils.MakeAppClientWithDeps(t, testClients, tt.authMiddleware)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
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

func TestDeleteGitSecret(t *testing.T) {
	authMiddleware := jwtassertion.NewMockMiddleware(t)

	t.Run("Deleting an existing git secret should return 204", func(t *testing.T) {
		openChoreoClient := apitestutils.CreateMockOpenChoreoClient()
		openChoreoClient.DeleteGitSecretFunc = func(ctx context.Context, namespaceName string, secretName string) error {
			return nil
		}

		testClients := wiring.TestClients{
			OpenChoreoClient: openChoreoClient,
		}

		app := apitestutils.MakeAppClientWithDeps(t, testClients, authMiddleware)

		url := fmt.Sprintf("/api/v1/orgs/%s/git-secrets/my-git-secret", gitSecretTestOrgName)
		req := httptest.NewRequest(http.MethodDelete, url, nil)

		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)

		// Validate service calls
		require.Len(t, openChoreoClient.DeleteGitSecretCalls(), 1)
		deleteCall := openChoreoClient.DeleteGitSecretCalls()[0]
		require.Equal(t, gitSecretTestOrgName, deleteCall.NamespaceName)
		require.Equal(t, "my-git-secret", deleteCall.SecretName)
	})

	validationTests := []struct {
		name           string
		authMiddleware jwtassertion.Middleware
		wantStatus     int
		wantErrMsg     string
		url            string
		setupMock      func() *clientmocks.OpenChoreoClientMock
	}{
		{
			name:           "return 404 on git secret not found",
			authMiddleware: authMiddleware,
			wantStatus:     404,
			wantErrMsg:     "Git secret not found",
			url:            fmt.Sprintf("/api/v1/orgs/%s/git-secrets/nonexistent-secret", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.DeleteGitSecretFunc = func(ctx context.Context, namespaceName string, secretName string) error {
					return utils.ErrNotFound
				}
				return mock
			},
		},
		{
			name:           "return 500 on service error",
			authMiddleware: authMiddleware,
			wantStatus:     500,
			wantErrMsg:     "Failed to delete git secret",
			url:            fmt.Sprintf("/api/v1/orgs/%s/git-secrets/my-secret", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.DeleteGitSecretFunc = func(ctx context.Context, namespaceName string, secretName string) error {
					return fmt.Errorf("internal service error")
				}
				return mock
			},
		},
		{
			name: "return 401 on missing authentication",
			authMiddleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing header: Authorization")
				})
			},
			wantStatus: 401,
			wantErrMsg: "missing header: Authorization",
			url:        fmt.Sprintf("/api/v1/orgs/%s/git-secrets/my-secret", gitSecretTestOrgName),
			setupMock: func() *clientmocks.OpenChoreoClientMock {
				mock := apitestutils.CreateMockOpenChoreoClient()
				mock.DeleteGitSecretFunc = func(ctx context.Context, namespaceName string, secretName string) error {
					return nil
				}
				return mock
			},
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			openChoreoClient := tt.setupMock()
			testClients := wiring.TestClients{
				OpenChoreoClient: openChoreoClient,
			}

			app := apitestutils.MakeAppClientWithDeps(t, testClients, tt.authMiddleware)

			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
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
