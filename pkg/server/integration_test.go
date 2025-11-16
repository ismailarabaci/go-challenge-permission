package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// HTTP request/response types for integration tests

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const contextUserIDKey contextKey = "contextUserID"

type CreateUserRequest struct {
	Name string `json:"name"`
}

type CreateUserResponse struct {
	ID int `json:"id"`
}

type GetUserResponse struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type CreateGroupResponse struct {
	ID int `json:"id"`
}

type GetGroupResponse struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type AddUserToGroupRequest struct {
	UserID int `json:"user_id"`
}

type AddPermissionRequest struct {
	SourceType string `json:"source_type"` // "user" or "group"
	TargetType string `json:"target_type"` // "user" or "group"
	SourceID   int    `json:"source_id"`
	TargetID   int    `json:"target_id"`
}

// HTTP Handler implementation for integration tests

type HTTPHandler struct {
	server *Server
}

func NewHTTPHandler(server *Server) *HTTPHandler {
	return &HTTPHandler{server: server}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.enrichContext(r)
	h.route(w, r.WithContext(ctx))
}

// enrichContext adds authentication context from headers
func (h *HTTPHandler) enrichContext(r *http.Request) context.Context {
	ctx := r.Context()
	if contextUserIDStr := r.Header.Get("X-Context-User-ID"); contextUserIDStr != "" {
		if contextUserID, err := strconv.Atoi(contextUserIDStr); err == nil {
			ctx = context.WithValue(ctx, contextUserIDKey, contextUserID)
		}
	}
	return ctx
}

// route dispatches requests to appropriate handlers
func (h *HTTPHandler) route(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	switch {
	case method == "POST" && path == "/users":
		h.handleCreateUser(w, r)
	case method == "GET" && strings.HasPrefix(path, "/users/"):
		h.handleGetUser(w, r)
	case method == "POST" && path == "/groups":
		h.handleCreateGroup(w, r)
	case method == "GET" && strings.HasPrefix(path, "/groups/"):
		h.handleGetGroup(w, r)
	case method == "POST" && strings.Contains(path, "/groups/") && strings.HasSuffix(path, "/users"):
		h.handleAddUserToGroup(w, r)
	case method == "POST" && path == "/permissions":
		h.handleAddPermission(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *HTTPHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.server.CreateUser(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateUserResponse{ID: id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path /users/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if context user ID is present (permission check)
	contextUserID, hasContext := r.Context().Value(contextUserIDKey).(int)

	var name string
	if hasContext {
		// Use permission check
		name, err = h.server.GetUserNameWithPermissionCheck(r.Context(), contextUserID, userID)
	} else {
		// No permission check
		name, err = h.server.GetUserName(r.Context(), userID)
	}

	if err != nil {
		// Check if it's a permission denied error
		if _, ok := err.(*PermissionDeniedError); ok {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := GetUserResponse{ID: userID, Name: name}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.server.CreateUserGroup(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateGroupResponse{ID: id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) handleGetGroup(w http.ResponseWriter, r *http.Request) {
	// Extract group ID from path /groups/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/groups/")
	groupID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if context user ID is present (permission check)
	contextUserID, hasContext := r.Context().Value(contextUserIDKey).(int)

	var name string
	if hasContext {
		// Use permission check
		name, err = h.server.GetUserGroupNameWithPermissionCheck(r.Context(), contextUserID, groupID)
	} else {
		// No permission check
		name, err = h.server.GetUserGroupName(r.Context(), groupID)
	}

	if err != nil {
		// Check if it's a permission denied error
		if _, ok := err.(*PermissionDeniedError); ok {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := GetGroupResponse{ID: groupID, Name: name}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) handleAddUserToGroup(w http.ResponseWriter, r *http.Request) {
	// Extract group ID from path /groups/{id}/users
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/users")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid group ID", http.StatusBadRequest)
		return
	}

	var req AddUserToGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.server.AddUserToGroup(r.Context(), req.UserID, groupID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) handleAddPermission(w http.ResponseWriter, r *http.Request) {
	var req AddPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	switch {
	case req.SourceType == "user" && req.TargetType == "user":
		err = h.server.AddUserToUserPermission(r.Context(), req.SourceID, req.TargetID)
	case req.SourceType == "user" && req.TargetType == "group":
		err = h.server.AddUserToUserGroupPermission(r.Context(), req.SourceID, req.TargetID)
	case req.SourceType == "group" && req.TargetType == "user":
		err = h.server.AddUserGroupToUserPermission(r.Context(), req.SourceID, req.TargetID)
	case req.SourceType == "group" && req.TargetType == "group":
		err = h.server.AddUserGroupToUserGroupPermission(r.Context(), req.SourceID, req.TargetID)
	default:
		http.Error(w, "invalid permission type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Test helpers

func setupHTTPTestServer(t *testing.T) (httpServer *httptest.Server, server *Server) {
	t.Helper()

	server = setupTestServer(t)
	handler := NewHTTPHandler(server)
	httpServer = httptest.NewServer(handler)

	return httpServer, server
}

func makeRequest(t *testing.T, method, url string, body interface{}, contextUserID *int) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if contextUserID != nil {
		req.Header.Set("X-Context-User-ID", strconv.Itoa(*contextUserID))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

func createUserViaHTTP(t *testing.T, baseURL, name string) int {
	t.Helper()

	reqBody := CreateUserRequest{Name: name}
	resp := makeRequest(t, "POST", baseURL+"/users", reqBody, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var respBody CreateUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return respBody.ID
}

func createGroupViaHTTP(t *testing.T, baseURL, name string) int {
	t.Helper()

	reqBody := CreateGroupRequest{Name: name}
	resp := makeRequest(t, "POST", baseURL+"/groups", reqBody, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var respBody CreateGroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return respBody.ID
}

func addUserToGroupViaHTTP(t *testing.T, baseURL string, userID, groupID int) {
	t.Helper()

	reqBody := AddUserToGroupRequest{UserID: userID}
	url := fmt.Sprintf("%s/groups/%d/users", baseURL, groupID)
	resp := makeRequest(t, "POST", url, reqBody, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", resp.StatusCode)
	}
}

func addPermissionViaHTTP(t *testing.T, baseURL, sourceType string, sourceID int, targetType string, targetID int) {
	t.Helper()

	reqBody := AddPermissionRequest{
		SourceType: sourceType,
		SourceID:   sourceID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	resp := makeRequest(t, "POST", baseURL+"/permissions", reqBody, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", resp.StatusCode)
	}
}

func getUserViaHTTP(t *testing.T, baseURL string, userID int, contextUserID *int) (name string, statusCode int) {
	t.Helper()

	url := fmt.Sprintf("%s/users/%d", baseURL, userID)
	resp := makeRequest(t, "GET", url, nil, contextUserID)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusForbidden {
		return "", resp.StatusCode
	}

	var respBody GetUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return respBody.Name, resp.StatusCode
}

func getGroupViaHTTP(t *testing.T, baseURL string, groupID int, contextUserID *int) (name string, statusCode int) {
	t.Helper()

	url := fmt.Sprintf("%s/groups/%d", baseURL, groupID)
	resp := makeRequest(t, "GET", url, nil, contextUserID)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusForbidden {
		return "", resp.StatusCode
	}

	var respBody GetGroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return respBody.Name, resp.StatusCode
}

// Integration Tests

// Test_Integration_ComplexPermissionScenario tests a complex permission scenario
// with hierarchical groups and group-to-group permissions via HTTP
func Test_Integration_ComplexPermissionScenario(t *testing.T) {
	httpServer, server := setupHTTPTestServer(t)
	defer httpServer.Close()
	defer server.Close()
	baseURL := httpServer.URL
	ctx := context.Background()

	// Create multiple users
	alice := createUserViaHTTP(t, baseURL, "Alice")
	bob := createUserViaHTTP(t, baseURL, "Bob")
	charlie := createUserViaHTTP(t, baseURL, "Charlie")
	dave := createUserViaHTTP(t, baseURL, "Dave")

	// Create hierarchical groups
	company := createGroupViaHTTP(t, baseURL, "Company")
	engineering := createGroupViaHTTP(t, baseURL, "Engineering")
	backend := createGroupViaHTTP(t, baseURL, "Backend")
	admins := createGroupViaHTTP(t, baseURL, "Admins")

	// Build hierarchy: company -> engineering -> backend
	if err := server.AddUserGroupToGroup(ctx, engineering, company); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}
	if err := server.AddUserGroupToGroup(ctx, backend, engineering); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}

	// Add users to nested groups
	addUserToGroupViaHTTP(t, baseURL, alice, admins)    // Alice is admin
	addUserToGroupViaHTTP(t, baseURL, bob, company)     // Bob at top level
	addUserToGroupViaHTTP(t, baseURL, charlie, backend) // Charlie at bottom level
	// Dave is not in any group

	// Set up group-to-group permissions: admins can access company (and nested groups)
	addPermissionViaHTTP(t, baseURL, "group", admins, "group", company)

	// Test 1: Alice (in admins) should access Bob (in company) via HTTP
	t.Run("admin can access company member", func(t *testing.T) {
		name, status := getUserViaHTTP(t, baseURL, bob, &alice)
		if status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", status)
		}
		if name != "Bob" {
			t.Errorf("Expected name 'Bob', got %q", name)
		}
	})

	// Test 2: Alice (in admins) should access Charlie (in backend, nested under company) via HTTP
	t.Run("admin can access nested group member", func(t *testing.T) {
		name, status := getUserViaHTTP(t, baseURL, charlie, &alice)
		if status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", status)
		}
		if name != "Charlie" {
			t.Errorf("Expected name 'Charlie', got %q", name)
		}
	})

	// Test 3: Dave should NOT access Bob (no permission)
	t.Run("user without permission denied", func(t *testing.T) {
		_, status := getUserViaHTTP(t, baseURL, bob, &dave)
		if status != http.StatusForbidden {
			t.Errorf("Expected status 403 (Forbidden), got %d", status)
		}
	})

	// Test 4: Bob should NOT access Alice (no reverse permission)
	t.Run("permission is not bidirectional", func(t *testing.T) {
		_, status := getUserViaHTTP(t, baseURL, alice, &bob)
		if status != http.StatusForbidden {
			t.Errorf("Expected status 403 (Forbidden), got %d", status)
		}
	})

	// Test 5: Add direct user-to-user permission and verify via HTTP
	t.Run("direct user permission works", func(t *testing.T) {
		addPermissionViaHTTP(t, baseURL, "user", dave, "user", bob)

		name, status := getUserViaHTTP(t, baseURL, bob, &dave)
		if status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", status)
		}
		if name != "Bob" {
			t.Errorf("Expected name 'Bob', got %q", name)
		}
	})
}

// Test_Integration_TransitiveGroupMembershipWithPermissions tests transitive
// group membership with 3-level hierarchy and permissions via HTTP
func Test_Integration_TransitiveGroupMembershipWithPermissions(t *testing.T) {
	httpServer, server := setupHTTPTestServer(t)
	defer httpServer.Close()
	defer server.Close()
	baseURL := httpServer.URL
	ctx := context.Background()

	// Create users
	admin := createUserViaHTTP(t, baseURL, "Admin")
	manager := createUserViaHTTP(t, baseURL, "Manager")
	developer := createUserViaHTTP(t, baseURL, "Developer")
	outsider := createUserViaHTTP(t, baseURL, "Outsider")

	// Create 3-level group hierarchy: organization -> department -> team
	organization := createGroupViaHTTP(t, baseURL, "Organization")
	department := createGroupViaHTTP(t, baseURL, "Department")
	team := createGroupViaHTTP(t, baseURL, "Team")

	// Create admin group
	adminGroup := createGroupViaHTTP(t, baseURL, "AdminGroup")

	// Build hierarchy
	if err := server.AddUserGroupToGroup(ctx, department, organization); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}
	if err := server.AddUserGroupToGroup(ctx, team, department); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}

	// Add users at different levels
	addUserToGroupViaHTTP(t, baseURL, admin, adminGroup)     // Admin in admin group
	addUserToGroupViaHTTP(t, baseURL, manager, organization) // Manager at top level
	addUserToGroupViaHTTP(t, baseURL, developer, team)       // Developer at bottom level
	// Outsider is not in any group

	// Set permissions at top level: admin group can access organization
	addPermissionViaHTTP(t, baseURL, "group", adminGroup, "group", organization)

	// Test 1: Admin should access manager (direct member of organization)
	t.Run("admin accesses direct organization member", func(t *testing.T) {
		name, status := getUserViaHTTP(t, baseURL, manager, &admin)
		if status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", status)
		}
		if name != "Manager" {
			t.Errorf("Expected name 'Manager', got %q", name)
		}
	})

	// Test 2: Admin should access developer (transitive member through department -> team)
	t.Run("admin accesses transitive member in nested team", func(t *testing.T) {
		name, status := getUserViaHTTP(t, baseURL, developer, &admin)
		if status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", status)
		}
		if name != "Developer" {
			t.Errorf("Expected name 'Developer', got %q", name)
		}
	})

	// Test 3: Admin can access organization group itself
	t.Run("admin accesses organization group", func(t *testing.T) {
		// Add permission to access group
		addPermissionViaHTTP(t, baseURL, "user", admin, "group", organization)

		name, status := getGroupViaHTTP(t, baseURL, organization, &admin)
		if status != http.StatusOK {
			t.Errorf("Expected status 200, got %d", status)
		}
		if name != "Organization" {
			t.Errorf("Expected group name 'Organization', got %q", name)
		}
	})

	// Test 4: Outsider should NOT access developer (no permission)
	t.Run("outsider denied access to developer", func(t *testing.T) {
		_, status := getUserViaHTTP(t, baseURL, developer, &outsider)
		if status != http.StatusForbidden {
			t.Errorf("Expected status 403 (Forbidden), got %d", status)
		}
	})

	// Test 5: Outsider should NOT access organization group (no permission)
	t.Run("outsider denied access to organization group", func(t *testing.T) {
		_, status := getGroupViaHTTP(t, baseURL, organization, &outsider)
		if status != http.StatusForbidden {
			t.Errorf("Expected status 403 (Forbidden), got %d", status)
		}
	})

	// Test 6: Manager should NOT access admin (no reverse permission)
	t.Run("no reverse permission from organization to admin", func(t *testing.T) {
		_, status := getUserViaHTTP(t, baseURL, admin, &manager)
		if status != http.StatusForbidden {
			t.Errorf("Expected status 403 (Forbidden), got %d", status)
		}
	})
}
