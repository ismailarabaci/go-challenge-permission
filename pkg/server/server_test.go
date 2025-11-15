package server

import (
	"context"
	"testing"
)

// run and validate the server through tests

// Stage 1 Tests
func TestStage1_CreateAndGetUser(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	
	// Test CreateUser
	userID, err := s.CreateUser(ctx, "Alice")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if userID <= 0 {
		t.Fatalf("Expected positive user ID, got %d", userID)
	}
	
	// Test GetUserName
	name, err := s.GetUserName(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserName failed: %v", err)
	}
	if name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", name)
	}
	
	// Test GetUserName with non-existent user
	_, err = s.GetUserName(ctx, 999999)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

// Stage 2 Tests
func TestStage2_UserGroups(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	
	// Create users
	user1, _ := s.CreateUser(ctx, "Bob")
	user2, _ := s.CreateUser(ctx, "Charlie")
	user3, _ := s.CreateUser(ctx, "Diana")
	
	// Create user group
	groupID, err := s.CreateUserGroup(ctx, "Developers")
	if err != nil {
		t.Fatalf("CreateUserGroup failed: %v", err)
	}
	if groupID <= 0 {
		t.Fatalf("Expected positive group ID, got %d", groupID)
	}
	
	// Test GetUserGroupName
	groupName, err := s.GetUserGroupName(ctx, groupID)
	if err != nil {
		t.Fatalf("GetUserGroupName failed: %v", err)
	}
	if groupName != "Developers" {
		t.Errorf("Expected group name 'Developers', got '%s'", groupName)
	}
	
	// Add users to group
	if err := s.AddUserToGroup(ctx, user1, groupID); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserToGroup(ctx, user2, groupID); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	
	// Get users in group
	users, err := s.GetUsersInGroup(ctx, groupID)
	if err != nil {
		t.Fatalf("GetUsersInGroup failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users in group, got %d", len(users))
	}
	
	// Verify it's an empty slice for empty group
	emptyGroupID, _ := s.CreateUserGroup(ctx, "Empty")
	emptyUsers, err := s.GetUsersInGroup(ctx, emptyGroupID)
	if err != nil {
		t.Fatalf("GetUsersInGroup failed for empty group: %v", err)
	}
	if emptyUsers == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(emptyUsers) != 0 {
		t.Errorf("Expected 0 users in empty group, got %d", len(emptyUsers))
	}
	
	// Test adding duplicate (should not error)
	if err := s.AddUserToGroup(ctx, user1, groupID); err != nil {
		t.Errorf("AddUserToGroup duplicate should not error: %v", err)
	}
	
	// user3 is not in the group
	users, _ = s.GetUsersInGroup(ctx, groupID)
	for _, uid := range users {
		if uid == user3 {
			t.Error("user3 should not be in group")
		}
	}
}

// Stage 3 Tests
func TestStage3_HierarchicalGroups(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	
	// Create groups
	engineering, _ := s.CreateUserGroup(ctx, "Engineering")
	backend, _ := s.CreateUserGroup(ctx, "Backend")
	frontend, _ := s.CreateUserGroup(ctx, "Frontend")
	
	// Add groups to groups
	if err := s.AddUserGroupToGroup(ctx, backend, engineering); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}
	if err := s.AddUserGroupToGroup(ctx, frontend, engineering); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}
	
	// Get groups in group
	subgroups, err := s.GetUserGroupsInGroup(ctx, engineering)
	if err != nil {
		t.Fatalf("GetUserGroupsInGroup failed: %v", err)
	}
	if len(subgroups) != 2 {
		t.Errorf("Expected 2 subgroups, got %d", len(subgroups))
	}
	
	// Test cycle detection - direct cycle
	err = s.AddUserGroupToGroup(ctx, engineering, engineering)
	if err == nil {
		t.Error("Expected error for self-cycle, got nil")
	}
	
	// Test cycle detection - indirect cycle
	ui, _ := s.CreateUserGroup(ctx, "UI")
	s.AddUserGroupToGroup(ctx, ui, frontend)
	err = s.AddUserGroupToGroup(ctx, engineering, ui) // Would create cycle
	if err == nil {
		t.Error("Expected error for indirect cycle, got nil")
	}
	
	// Verify empty groups work
	emptyGroup, _ := s.CreateUserGroup(ctx, "EmptyGroup")
	emptySubgroups, err := s.GetUserGroupsInGroup(ctx, emptyGroup)
	if err != nil {
		t.Fatalf("GetUserGroupsInGroup failed for empty group: %v", err)
	}
	if emptySubgroups == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(emptySubgroups) != 0 {
		t.Errorf("Expected 0 subgroups, got %d", len(emptySubgroups))
	}
}

// Stage 4 Tests
func TestStage4_TransitiveMembership(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	
	// Create users
	alice, _ := s.CreateUser(ctx, "Alice")
	bob, _ := s.CreateUser(ctx, "Bob")
	charlie, _ := s.CreateUser(ctx, "Charlie")
	dave, _ := s.CreateUser(ctx, "Dave")
	
	// Create hierarchical groups
	company, _ := s.CreateUserGroup(ctx, "Company")
	engineering, _ := s.CreateUserGroup(ctx, "Engineering")
	backend, _ := s.CreateUserGroup(ctx, "Backend")
	
	// Build hierarchy: company -> engineering -> backend
	s.AddUserGroupToGroup(ctx, engineering, company)
	s.AddUserGroupToGroup(ctx, backend, engineering)
	
	// Add users at different levels
	s.AddUserToGroup(ctx, alice, company)    // Direct to top level
	s.AddUserToGroup(ctx, bob, engineering)  // Middle level
	s.AddUserToGroup(ctx, charlie, backend)  // Bottom level
	// dave is not in any group
	
	// Get transitive users in company (should include all)
	users, err := s.GetUsersInGroupTransitive(ctx, company)
	if err != nil {
		t.Fatalf("GetUsersInGroupTransitive failed: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users in company (transitive), got %d", len(users))
	}
	
	// Verify all expected users are present
	userMap := make(map[int]bool)
	for _, uid := range users {
		userMap[uid] = true
	}
	if !userMap[alice] || !userMap[bob] || !userMap[charlie] {
		t.Error("Not all expected users found in transitive membership")
	}
	if userMap[dave] {
		t.Error("dave should not be in transitive membership")
	}
	
	// Get transitive users in engineering (should include bob and charlie)
	users, err = s.GetUsersInGroupTransitive(ctx, engineering)
	if err != nil {
		t.Fatalf("GetUsersInGroupTransitive failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users in engineering (transitive), got %d", len(users))
	}
	
	// Get transitive users in backend (should include only charlie)
	users, err = s.GetUsersInGroupTransitive(ctx, backend)
	if err != nil {
		t.Fatalf("GetUsersInGroupTransitive failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user in backend (transitive), got %d", len(users))
	}
	if len(users) > 0 && users[0] != charlie {
		t.Errorf("Expected charlie in backend, got user %d", users[0])
	}
	
	// Test empty group
	emptyGroup, _ := s.CreateUserGroup(ctx, "Empty")
	emptyUsers, err := s.GetUsersInGroupTransitive(ctx, emptyGroup)
	if err != nil {
		t.Fatalf("GetUsersInGroupTransitive failed for empty group: %v", err)
	}
	if emptyUsers == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(emptyUsers) != 0 {
		t.Errorf("Expected 0 users, got %d", len(emptyUsers))
	}
}

// Stage 5 Tests
func TestStage5_Permissions(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	
	// Create users
	alice, _ := s.CreateUser(ctx, "Alice")
	bob, _ := s.CreateUser(ctx, "Bob")
	charlie, _ := s.CreateUser(ctx, "Charlie")
	dave, _ := s.CreateUser(ctx, "Dave")
	eve, _ := s.CreateUser(ctx, "Eve")
	frank, _ := s.CreateUser(ctx, "Frank")
	
	// Create groups
	admins, _ := s.CreateUserGroup(ctx, "Admins")
	users, _ := s.CreateUserGroup(ctx, "Users")
	
	// Setup group memberships
	s.AddUserToGroup(ctx, alice, admins)
	s.AddUserToGroup(ctx, bob, users)
	s.AddUserToGroup(ctx, charlie, users)
	
	// Test Scenario 1: Direct user-to-user permission
	s.AddUserToUserPermission(ctx, alice, bob)
	name, err := s.GetUserNameWithPermissionCheck(ctx, alice, bob)
	if err != nil {
		t.Errorf("Scenario 1: Expected alice to access bob, got error: %v", err)
	}
	if name != "Bob" {
		t.Errorf("Scenario 1: Expected name 'Bob', got '%s'", name)
	}
	
	// Test that permission is not bidirectional
	_, err = s.GetUserNameWithPermissionCheck(ctx, bob, alice)
	if err == nil {
		t.Error("Scenario 1: Permission should not be bidirectional")
	}
	
	// Test Scenario 2: Source user in group -> target user
	s.AddUserGroupToUserPermission(ctx, admins, charlie)
	name, err = s.GetUserNameWithPermissionCheck(ctx, alice, charlie)
	if err != nil {
		t.Errorf("Scenario 2: Expected alice (in admins) to access charlie, got error: %v", err)
	}
	if name != "Charlie" {
		t.Errorf("Scenario 2: Expected name 'Charlie', got '%s'", name)
	}
	
	// Test Scenario 3: Source user -> target user in group
	s.AddUserToUserGroupPermission(ctx, dave, users)
	name, err = s.GetUserNameWithPermissionCheck(ctx, dave, bob)
	if err != nil {
		t.Errorf("Scenario 3: Expected dave to access bob (in users), got error: %v", err)
	}
	if name != "Bob" {
		t.Errorf("Scenario 3: Expected name 'Bob', got '%s'", name)
	}
	
	name, err = s.GetUserNameWithPermissionCheck(ctx, dave, charlie)
	if err != nil {
		t.Errorf("Scenario 3: Expected dave to access charlie (in users), got error: %v", err)
	}
	if name != "Charlie" {
		t.Errorf("Scenario 3: Expected name 'Charlie', got '%s'", name)
	}
	
	// Test Scenario 4: Source user in group -> target user in group
	managers, _ := s.CreateUserGroup(ctx, "Managers")
	s.AddUserToGroup(ctx, eve, managers)
	s.AddUserGroupToUserGroupPermission(ctx, managers, users)
	
	name, err = s.GetUserNameWithPermissionCheck(ctx, eve, bob)
	if err != nil {
		t.Errorf("Scenario 4: Expected eve (in managers) to access bob (in users), got error: %v", err)
	}
	if name != "Bob" {
		t.Errorf("Scenario 4: Expected name 'Bob', got '%s'", name)
	}
	
	// Test no permission
	_, err = s.GetUserNameWithPermissionCheck(ctx, frank, bob)
	if err == nil {
		t.Error("Expected error for frank accessing bob without permission")
	}
	
	// Test GetUserGroupNameWithPermissionCheck
	// Direct user to group permission
	s.AddUserToUserGroupPermission(ctx, alice, admins)
	groupName, err := s.GetUserGroupNameWithPermissionCheck(ctx, alice, admins)
	if err != nil {
		t.Errorf("Expected alice to access admins group, got error: %v", err)
	}
	if groupName != "Admins" {
		t.Errorf("Expected group name 'Admins', got '%s'", groupName)
	}
	
	// No permission on group
	_, err = s.GetUserGroupNameWithPermissionCheck(ctx, frank, admins)
	if err == nil {
		t.Error("Expected error for frank accessing admins group without permission")
	}
}

// Test transitive permissions with hierarchical groups
func TestStage5_TransitivePermissions(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	
	// Create users
	admin, _ := s.CreateUser(ctx, "Admin")
	member, _ := s.CreateUser(ctx, "Member")
	
	// Create hierarchical groups
	organization, _ := s.CreateUserGroup(ctx, "Organization")
	department, _ := s.CreateUserGroup(ctx, "Department")
	team, _ := s.CreateUserGroup(ctx, "Team")
	
	// Build hierarchy: organization -> department -> team
	s.AddUserGroupToGroup(ctx, department, organization)
	s.AddUserGroupToGroup(ctx, team, department)
	
	// Add users to nested groups
	s.AddUserToGroup(ctx, admin, organization)  // Top level
	s.AddUserToGroup(ctx, member, team)         // Bottom level
	
	// Admin group has permission on team members
	adminGroup, _ := s.CreateUserGroup(ctx, "AdminGroup")
	s.AddUserToGroup(ctx, admin, adminGroup)
	s.AddUserGroupToUserGroupPermission(ctx, adminGroup, organization)
	
	// Admin should be able to access member (transitive through groups)
	name, err := s.GetUserNameWithPermissionCheck(ctx, admin, member)
	if err != nil {
		t.Errorf("Expected admin to access member transitively, got error: %v", err)
	}
	if name != "Member" {
		t.Errorf("Expected name 'Member', got '%s'", name)
	}
}
