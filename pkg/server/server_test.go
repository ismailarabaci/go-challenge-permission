package server

import (
	"context"
	"testing"
)

// Test helper to create a test server
func setupTestServer(t *testing.T) *Server {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	return s
}

// Stage 1 Tests - User Operations

func Test_Stage1_CreateUser(t *testing.T) {
	tests := []struct {
		name     string
		userName string
	}{
		{name: "create user Alice", userName: "Alice"},
		{name: "create user Bob", userName: "Bob"},
		{name: "create user with spaces", userName: "John Doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestServer(t)
			defer s.Close()
			ctx := context.Background()

			userID, err := s.CreateUser(ctx, tt.userName)
			if err != nil {
				t.Fatalf("CreateUser failed: %v", err)
			}
			if userID <= 0 {
				t.Errorf("Expected positive user ID, got %d", userID)
			}
		})
	}
}

func Test_Stage1_GetUserName(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
	}{
		{name: "get user Alice", userName: "Alice"},
		{name: "get user Bob", userName: "Bob"},
		{name: "get user Charlie", userName: "Charlie"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create user
			userID, err := s.CreateUser(ctx, tt.userName)
			if err != nil {
				t.Fatalf("CreateUser failed: %v", err)
			}

			// Get user name
			name, err := s.GetUserName(ctx, userID)
			if err != nil {
				t.Fatalf("GetUserName failed: %v", err)
			}
			if name != tt.userName {
				t.Errorf("Expected name %q, got %q", tt.userName, name)
			}
		})
	}
}

func Test_Stage1_GetUserName_NonExistent(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name   string
		userID int
	}{
		{name: "large non-existent ID", userID: 999999},
		{name: "negative ID", userID: -1},
		{name: "zero ID", userID: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.GetUserName(ctx, tt.userID)
			if err == nil {
				t.Error("Expected error for non-existent user, got nil")
			}
		})
	}
}

// Stage 2 Tests - User Groups

func Test_Stage2_CreateUserGroup(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
	}{
		{name: "create group Developers", groupName: "Developers"},
		{name: "create group Admins", groupName: "Admins"},
		{name: "create group with spaces", groupName: "Engineering Team"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestServer(t)
			defer s.Close()
			ctx := context.Background()

			groupID, err := s.CreateUserGroup(ctx, tt.groupName)
			if err != nil {
				t.Fatalf("CreateUserGroup failed: %v", err)
			}
			if groupID <= 0 {
				t.Errorf("Expected positive group ID, got %d", groupID)
			}
		})
	}
}

func Test_Stage2_GetUserGroupName(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name      string
		groupName string
	}{
		{name: "get group Developers", groupName: "Developers"},
		{name: "get group Admins", groupName: "Admins"},
		{name: "get group Users", groupName: "Users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupID, err := s.CreateUserGroup(ctx, tt.groupName)
			if err != nil {
				t.Fatalf("CreateUserGroup failed: %v", err)
			}

			name, err := s.GetUserGroupName(ctx, groupID)
			if err != nil {
				t.Fatalf("GetUserGroupName failed: %v", err)
			}
			if name != tt.groupName {
				t.Errorf("Expected group name %q, got %q", tt.groupName, name)
			}
		})
	}
}

func Test_Stage2_AddUserToGroup(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name      string
		groupName string
		userNames []string
		wantCount int
	}{
		{
			name:      "add single user",
			groupName: "Group1",
			userNames: []string{"Alice"},
			wantCount: 1,
		},
		{
			name:      "add multiple users",
			groupName: "Group2",
			userNames: []string{"Bob", "Charlie", "Diana"},
			wantCount: 3,
		},
		{
			name:      "add two users",
			userNames: []string{"Eve", "Frank"},
			groupName: "Group3",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create group
			groupID, err := s.CreateUserGroup(ctx, tt.groupName)
			if err != nil {
				t.Fatalf("CreateUserGroup failed: %v", err)
			}

			// Create and add users
			for _, userName := range tt.userNames {
				uid, createErr := s.CreateUser(ctx, userName)
				if createErr != nil {
					t.Fatalf("CreateUser failed: %v", createErr)
				}

				addErr := s.AddUserToGroup(ctx, uid, groupID)
				if addErr != nil {
					t.Fatalf("AddUserToGroup failed: %v", addErr)
				}
			}

			// Verify count
			users, err := s.GetUsersInGroup(ctx, groupID)
			if err != nil {
				t.Fatalf("GetUsersInGroup failed: %v", err)
			}
			if len(users) != tt.wantCount {
				t.Errorf("Expected %d users in group, got %d", tt.wantCount, len(users))
			}
		})
	}
}

func Test_Stage2_GetUsersInGroup_Empty(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	groupID, err := s.CreateUserGroup(ctx, "EmptyGroup")
	if err != nil {
		t.Fatalf("CreateUserGroup failed: %v", err)
	}

	users, err := s.GetUsersInGroup(ctx, groupID)
	if err != nil {
		t.Fatalf("GetUsersInGroup failed: %v", err)
	}

	if users == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users in empty group, got %d", len(users))
	}
}

func Test_Stage2_AddUserToGroup_Duplicate(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	userID, _ := s.CreateUser(ctx, "Alice")
	groupID, _ := s.CreateUserGroup(ctx, "TestGroup")

	// Add user first time
	err := s.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		t.Fatalf("First AddUserToGroup failed: %v", err)
	}

	// Add same user again (should not error)
	err = s.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		t.Errorf("AddUserToGroup duplicate should not error: %v", err)
	}

	// Verify user is still in group and count is still 1
	users, _ := s.GetUsersInGroup(ctx, groupID)
	if len(users) != 1 {
		t.Errorf("Expected 1 user after duplicate add, got %d", len(users))
	}
}

func Test_Stage2_GetUsersInGroup_Membership(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	// Create users
	user1, _ := s.CreateUser(ctx, "Alice")
	user2, _ := s.CreateUser(ctx, "Bob")
	user3, _ := s.CreateUser(ctx, "Charlie")

	// Create group and add only user1 and user2
	groupID, _ := s.CreateUserGroup(ctx, "TestGroup")
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

	// Create map for easy lookup
	userMap := make(map[int]bool)
	for _, uid := range users {
		userMap[uid] = true
	}

	// Verify user1 and user2 are in group
	if !userMap[user1] {
		t.Error("user1 should be in group")
	}
	if !userMap[user2] {
		t.Error("user2 should be in group")
	}

	// Verify user3 is NOT in group
	if userMap[user3] {
		t.Error("user3 should not be in group")
	}
}

// Stage 3 Tests - Hierarchical Groups

func Test_Stage3_AddGroupToGroup(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name      string
		parent    string
		children  []string
		wantCount int
	}{
		{
			name:      "single child group",
			parent:    "Engineering",
			children:  []string{"Backend"},
			wantCount: 1,
		},
		{
			name:      "multiple child groups",
			parent:    "Company",
			children:  []string{"Engineering", "Sales", "Marketing"},
			wantCount: 3,
		},
		{
			name:      "two child groups",
			parent:    "Development",
			children:  []string{"Frontend", "Backend"},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parent group
			parentID, err := s.CreateUserGroup(ctx, tt.parent)
			if err != nil {
				t.Fatalf("CreateUserGroup failed: %v", err)
			}

			// Create and add child groups
			for _, childName := range tt.children {
				cid, createErr := s.CreateUserGroup(ctx, childName)
				if createErr != nil {
					t.Fatalf("CreateUserGroup failed: %v", createErr)
				}

				addErr := s.AddUserGroupToGroup(ctx, cid, parentID)
				if addErr != nil {
					t.Fatalf("AddUserGroupToGroup failed: %v", addErr)
				}
			}

			// Verify count
			subgroups, err := s.GetUserGroupsInGroup(ctx, parentID)
			if err != nil {
				t.Fatalf("GetUserGroupsInGroup failed: %v", err)
			}
			if len(subgroups) != tt.wantCount {
				t.Errorf("Expected %d subgroups, got %d", tt.wantCount, len(subgroups))
			}
		})
	}
}

func Test_Stage3_GetUserGroupsInGroup_Empty(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	emptyGroup, err := s.CreateUserGroup(ctx, "EmptyGroup")
	if err != nil {
		t.Fatalf("CreateUserGroup failed: %v", err)
	}

	subgroups, err := s.GetUserGroupsInGroup(ctx, emptyGroup)
	if err != nil {
		t.Fatalf("GetUserGroupsInGroup failed: %v", err)
	}

	if subgroups == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(subgroups) != 0 {
		t.Errorf("Expected 0 subgroups, got %d", len(subgroups))
	}
}

func Test_Stage3_CycleDetection_SelfCycle(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	groupID, _ := s.CreateUserGroup(ctx, "TestGroup")

	// Try to add group to itself
	err := s.AddUserGroupToGroup(ctx, groupID, groupID)
	if err == nil {
		t.Error("Expected error for self-cycle, got nil")
	}
}

func Test_Stage3_CycleDetection_IndirectCycle(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() (int, int) // Returns childID, parentID that would create cycle
		description string
	}{
		{
			name: "two-level cycle",
			setup: func() (int, int) {
				a, _ := s.CreateUserGroup(ctx, "GroupA")
				b, _ := s.CreateUserGroup(ctx, "GroupB")
				if err := s.AddUserGroupToGroup(ctx, b, a); err != nil {
					t.Fatalf("AddUserGroupToGroup failed: %v", err)
				}
				return a, b // Try to add A to B (creates cycle)
			},
			description: "A->B, then B->A",
		},
		{
			name: "three-level cycle",
			setup: func() (int, int) {
				a, _ := s.CreateUserGroup(ctx, "GroupX")
				b, _ := s.CreateUserGroup(ctx, "GroupY")
				c, _ := s.CreateUserGroup(ctx, "GroupZ")
				if err := s.AddUserGroupToGroup(ctx, b, a); err != nil {
					t.Fatalf("AddUserGroupToGroup failed: %v", err)
				}
				if err := s.AddUserGroupToGroup(ctx, c, b); err != nil {
					t.Fatalf("AddUserGroupToGroup failed: %v", err)
				}
				return a, c // Try to add A to C (creates cycle)
			},
			description: "A->B->C, then C->A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			childID, parentID := tt.setup()

			err := s.AddUserGroupToGroup(ctx, childID, parentID)
			if err == nil {
				t.Errorf("Expected error for indirect cycle (%s), got nil", tt.description)
			}
		})
	}
}

// Stage 4 Tests - Transitive Membership

func Test_Stage4_TransitiveMembership_ThreeLevelHierarchy(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	// Create users
	alice, _ := s.CreateUser(ctx, "Alice")
	bob, _ := s.CreateUser(ctx, "Bob")
	charlie, _ := s.CreateUser(ctx, "Charlie")
	dave, _ := s.CreateUser(ctx, "Dave")

	// Create hierarchical groups: company -> engineering -> backend
	company, _ := s.CreateUserGroup(ctx, "Company")
	engineering, _ := s.CreateUserGroup(ctx, "Engineering")
	backend, _ := s.CreateUserGroup(ctx, "Backend")

	if err := s.AddUserGroupToGroup(ctx, engineering, company); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}
	if err := s.AddUserGroupToGroup(ctx, backend, engineering); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}

	// Add users at different levels
	if err := s.AddUserToGroup(ctx, alice, company); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserToGroup(ctx, bob, engineering); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserToGroup(ctx, charlie, backend); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	// dave is not in any group

	tests := []struct {
		name         string
		description  string
		wantUsers    []int
		notWantUsers []int
		groupID      int
		wantCount    int
	}{
		{
			name:         "company includes all nested users",
			description:  "Top level group should include all users from nested groups",
			wantUsers:    []int{alice, bob, charlie},
			notWantUsers: []int{dave},
			groupID:      company,
			wantCount:    3,
		},
		{
			name:         "engineering includes middle and bottom users",
			description:  "Middle level group should include its users and nested group users",
			wantUsers:    []int{bob, charlie},
			notWantUsers: []int{alice, dave},
			groupID:      engineering,
			wantCount:    2,
		},
		{
			name:         "backend includes only direct user",
			description:  "Bottom level group should include only its direct users",
			wantUsers:    []int{charlie},
			notWantUsers: []int{alice, bob, dave},
			groupID:      backend,
			wantCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := s.GetUsersInGroupTransitive(ctx, tt.groupID)
			if err != nil {
				t.Fatalf("GetUsersInGroupTransitive failed: %v", err)
			}

			if len(users) != tt.wantCount {
				t.Errorf("Expected %d users, got %d", tt.wantCount, len(users))
			}

			// Create map for easy lookup
			userMap := make(map[int]bool)
			for _, uid := range users {
				userMap[uid] = true
			}

			// Verify expected users are present
			for _, wantUser := range tt.wantUsers {
				if !userMap[wantUser] {
					t.Errorf("Expected user %d to be in group", wantUser)
				}
			}

			// Verify unwanted users are not present
			for _, notWantUser := range tt.notWantUsers {
				if userMap[notWantUser] {
					t.Errorf("User %d should not be in group", notWantUser)
				}
			}
		})
	}
}

func Test_Stage4_TransitiveMembership_MultipleHierarchies(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	tests := []struct {
		name        string
		setupFunc   func() (int, []int) // Returns groupID and expected userIDs
		description string
		wantCount   int
	}{
		{
			name: "single level with multiple users",
			setupFunc: func() (int, []int) {
				u1, _ := s.CreateUser(ctx, "User1")
				u2, _ := s.CreateUser(ctx, "User2")
				u3, _ := s.CreateUser(ctx, "User3")
				g, _ := s.CreateUserGroup(ctx, "SingleLevel")
				_ = s.AddUserToGroup(ctx, u1, g)
				_ = s.AddUserToGroup(ctx, u2, g)
				_ = s.AddUserToGroup(ctx, u3, g)
				return g, []int{u1, u2, u3}
			},
			description: "Single level group with multiple direct users",
			wantCount:   3,
		},
		{
			name: "two level hierarchy",
			setupFunc: func() (int, []int) {
				u1, _ := s.CreateUser(ctx, "UserA")
				u2, _ := s.CreateUser(ctx, "UserB")
				parent, _ := s.CreateUserGroup(ctx, "Parent")
				child, _ := s.CreateUserGroup(ctx, "Child")
				_ = s.AddUserGroupToGroup(ctx, child, parent)
				_ = s.AddUserToGroup(ctx, u1, parent)
				_ = s.AddUserToGroup(ctx, u2, child)
				return parent, []int{u1, u2}
			},
			description: "Two level hierarchy with users at both levels",
			wantCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupID, expectedUsers := tt.setupFunc()

			users, err := s.GetUsersInGroupTransitive(ctx, groupID)
			if err != nil {
				t.Fatalf("GetUsersInGroupTransitive failed: %v", err)
			}

			if len(users) != tt.wantCount {
				t.Errorf("Expected %d users, got %d", tt.wantCount, len(users))
			}

			// Verify all expected users are present
			userMap := make(map[int]bool)
			for _, uid := range users {
				userMap[uid] = true
			}

			for _, expectedUser := range expectedUsers {
				if !userMap[expectedUser] {
					t.Errorf("Expected user %d to be in transitive membership", expectedUser)
				}
			}
		})
	}
}

func Test_Stage4_TransitiveMembership_EmptyGroup(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	emptyGroup, _ := s.CreateUserGroup(ctx, "EmptyGroup")

	users, err := s.GetUsersInGroupTransitive(ctx, emptyGroup)
	if err != nil {
		t.Fatalf("GetUsersInGroupTransitive failed: %v", err)
	}

	if users == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

// Stage 5 Tests - Permissions

// Scenario 1: Direct user-to-user permission
func Test_Stage5_DirectUserToUserPermission(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	alice, _ := s.CreateUser(ctx, "Alice")
	bob, _ := s.CreateUser(ctx, "Bob")

	// Grant alice permission to access bob
	err := s.AddUserToUserPermission(ctx, alice, bob)
	if err != nil {
		t.Fatalf("AddUserToUserPermission failed: %v", err)
	}

	// Alice should be able to access bob
	name, err := s.GetUserNameWithPermissionCheck(ctx, alice, bob)
	if err != nil {
		t.Errorf("Expected alice to access bob, got error: %v", err)
	}
	if name != "Bob" {
		t.Errorf("Expected name 'Bob', got %q", name)
	}
}

func Test_Stage5_DirectPermission_NotBidirectional(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	alice, _ := s.CreateUser(ctx, "Alice")
	bob, _ := s.CreateUser(ctx, "Bob")

	// Grant alice -> bob permission (one way)
	if err := s.AddUserToUserPermission(ctx, alice, bob); err != nil {
		t.Fatalf("AddUserToUserPermission failed: %v", err)
	}

	// Bob should NOT be able to access alice
	_, err := s.GetUserNameWithPermissionCheck(ctx, bob, alice)
	if err == nil {
		t.Error("Permission should not be bidirectional: bob should not access alice")
	}
}

// Scenario 2: Source user in group -> target user
func Test_Stage5_UserInGroupToUserPermission(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	alice, _ := s.CreateUser(ctx, "Alice")
	charlie, _ := s.CreateUser(ctx, "Charlie")
	admins, _ := s.CreateUserGroup(ctx, "Admins")

	// Add alice to admins group
	if err := s.AddUserToGroup(ctx, alice, admins); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}

	// Grant admins group permission to access charlie
	err := s.AddUserGroupToUserPermission(ctx, admins, charlie)
	if err != nil {
		t.Fatalf("AddUserGroupToUserPermission failed: %v", err)
	}

	// Alice (in admins) should be able to access charlie
	name, err := s.GetUserNameWithPermissionCheck(ctx, alice, charlie)
	if err != nil {
		t.Errorf("Expected alice (in admins) to access charlie, got error: %v", err)
	}
	if name != "Charlie" {
		t.Errorf("Expected name 'Charlie', got %q", name)
	}
}

// Scenario 3: Source user -> target user in group
func Test_Stage5_UserToUserInGroupPermission(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	bob, _ := s.CreateUser(ctx, "Bob")
	charlie, _ := s.CreateUser(ctx, "Charlie")
	dave, _ := s.CreateUser(ctx, "Dave")
	users, _ := s.CreateUserGroup(ctx, "Users")

	// Add bob and charlie to users group
	if err := s.AddUserToGroup(ctx, bob, users); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserToGroup(ctx, charlie, users); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}

	// Grant dave permission to access users group
	err := s.AddUserToUserGroupPermission(ctx, dave, users)
	if err != nil {
		t.Fatalf("AddUserToUserGroupPermission failed: %v", err)
	}

	// Dave should be able to access bob (in users)
	name, err := s.GetUserNameWithPermissionCheck(ctx, dave, bob)
	if err != nil {
		t.Errorf("Expected dave to access bob (in users), got error: %v", err)
	}
	if name != "Bob" {
		t.Errorf("Expected name 'Bob', got %q", name)
	}

	// Dave should also be able to access charlie (in users)
	name, err = s.GetUserNameWithPermissionCheck(ctx, dave, charlie)
	if err != nil {
		t.Errorf("Expected dave to access charlie (in users), got error: %v", err)
	}
	if name != "Charlie" {
		t.Errorf("Expected name 'Charlie', got %q", name)
	}
}

// Scenario 4: Source user in group -> target user in group
func Test_Stage5_UserInGroupToUserInGroupPermission(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	bob, _ := s.CreateUser(ctx, "Bob")
	eve, _ := s.CreateUser(ctx, "Eve")
	managers, _ := s.CreateUserGroup(ctx, "Managers")
	users, _ := s.CreateUserGroup(ctx, "Users")

	// Add users to their respective groups
	if err := s.AddUserToGroup(ctx, eve, managers); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserToGroup(ctx, bob, users); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}

	// Grant managers group permission to access users group
	err := s.AddUserGroupToUserGroupPermission(ctx, managers, users)
	if err != nil {
		t.Fatalf("AddUserGroupToUserGroupPermission failed: %v", err)
	}

	// Eve (in managers) should be able to access bob (in users)
	name, err := s.GetUserNameWithPermissionCheck(ctx, eve, bob)
	if err != nil {
		t.Errorf("Expected eve (in managers) to access bob (in users), got error: %v", err)
	}
	if name != "Bob" {
		t.Errorf("Expected name 'Bob', got %q", name)
	}
}

func Test_Stage5_NoPermission(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	alice, _ := s.CreateUser(ctx, "Alice")
	bob, _ := s.CreateUser(ctx, "Bob")
	charlie, _ := s.CreateUser(ctx, "Charlie")

	// Grant alice -> bob permission only
	if err := s.AddUserToUserPermission(ctx, alice, bob); err != nil {
		t.Fatalf("AddUserToUserPermission failed: %v", err)
	}

	// Charlie should NOT be able to access bob (no permission)
	_, err := s.GetUserNameWithPermissionCheck(ctx, charlie, bob)
	if err == nil {
		t.Error("Expected error for charlie accessing bob without permission")
	}

	// Charlie should NOT be able to access alice (no permission)
	_, err = s.GetUserNameWithPermissionCheck(ctx, charlie, alice)
	if err == nil {
		t.Error("Expected error for charlie accessing alice without permission")
	}
}

func Test_Stage5_GetUserGroupNameWithPermissionCheck(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	alice, _ := s.CreateUser(ctx, "Alice")
	frank, _ := s.CreateUser(ctx, "Frank")
	admins, _ := s.CreateUserGroup(ctx, "Admins")

	// Grant alice permission to access admins group
	err := s.AddUserToUserGroupPermission(ctx, alice, admins)
	if err != nil {
		t.Fatalf("AddUserToUserGroupPermission failed: %v", err)
	}

	// Alice should be able to access admins group
	groupName, err := s.GetUserGroupNameWithPermissionCheck(ctx, alice, admins)
	if err != nil {
		t.Errorf("Expected alice to access admins group, got error: %v", err)
	}
	if groupName != "Admins" {
		t.Errorf("Expected group name 'Admins', got %q", groupName)
	}

	// Frank should NOT be able to access admins group (no permission)
	_, err = s.GetUserGroupNameWithPermissionCheck(ctx, frank, admins)
	if err == nil {
		t.Error("Expected error for frank accessing admins group without permission")
	}
}

func Test_Stage5_TransitivePermissions(t *testing.T) {
	s := setupTestServer(t)
	defer s.Close()
	ctx := context.Background()

	// Create users
	admin, _ := s.CreateUser(ctx, "Admin")
	member, _ := s.CreateUser(ctx, "Member")

	// Create hierarchical groups: organization -> department -> team
	organization, _ := s.CreateUserGroup(ctx, "Organization")
	department, _ := s.CreateUserGroup(ctx, "Department")
	team, _ := s.CreateUserGroup(ctx, "Team")

	if err := s.AddUserGroupToGroup(ctx, department, organization); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}
	if err := s.AddUserGroupToGroup(ctx, team, department); err != nil {
		t.Fatalf("AddUserGroupToGroup failed: %v", err)
	}

	// Add users to nested groups
	if err := s.AddUserToGroup(ctx, admin, organization); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserToGroup(ctx, member, team); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}

	// Admin group has permission on organization (which includes team transitively)
	adminGroup, _ := s.CreateUserGroup(ctx, "AdminGroup")
	if err := s.AddUserToGroup(ctx, admin, adminGroup); err != nil {
		t.Fatalf("AddUserToGroup failed: %v", err)
	}
	if err := s.AddUserGroupToUserGroupPermission(ctx, adminGroup, organization); err != nil {
		t.Fatalf("AddUserGroupToUserGroupPermission failed: %v", err)
	}

	// Admin should be able to access member (transitive through groups)
	name, err := s.GetUserNameWithPermissionCheck(ctx, admin, member)
	if err != nil {
		t.Errorf("Expected admin to access member transitively, got error: %v", err)
	}
	if name != "Member" {
		t.Errorf("Expected name 'Member', got %q", name)
	}
}
