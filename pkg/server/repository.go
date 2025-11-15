package server

import "context"

// Repository defines the interface for data access operations
// This abstraction allows for different storage implementations and easier testing
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, name string) (int, error)
	GetUserByID(ctx context.Context, userID int) (string, error)

	// User group operations
	CreateUserGroup(ctx context.Context, name string) (int, error)
	GetUserGroupByID(ctx context.Context, groupID int) (string, error)

	// Membership operations
	AddUserToGroup(ctx context.Context, userID, groupID int) error
	GetUsersInGroup(ctx context.Context, groupID int) ([]int, error)
	GetUsersInGroupTransitive(ctx context.Context, groupID int) ([]int, error)

	// Hierarchy operations
	AddGroupToGroup(ctx context.Context, childID, parentID int) error
	GetGroupsInGroup(ctx context.Context, groupID int) ([]int, error)
	WouldCreateCycle(ctx context.Context, childID, parentID int) (bool, error)

	// Permission operations
	AddPermission(ctx context.Context, sourceType, targetType string, sourceID, targetID int) error
	HasUserPermissionOnUser(ctx context.Context, sourceUserID, targetUserID int) (bool, error)
	HasUserPermissionOnGroup(ctx context.Context, sourceUserID, targetGroupID int) (bool, error)

	// Close closes the repository and releases any resources
	Close() error
}
