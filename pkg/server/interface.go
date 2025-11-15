package server

import "context"

// enforce interface compliance
var _ Stage5 = (*Server)(nil)

/*
	The challenge is divided into stages, one for each of the interfaces below.
	Each stage has a test suite that checks that the server implements the requirements of the stage.
*/

// The server must be able to create users and retrieve their names
type Stage1 interface {
	CreateUser(ctx context.Context, name string) (int, error)
	GetUserName(ctx context.Context, userID int) (string, error)
}

// The server must be able to group users into user groups
type Stage2 interface {
	Stage1
	CreateUserGroup(ctx context.Context, name string) (int, error)
	GetUserGroupName(ctx context.Context, userGroupID int) (string, error)
	AddUserToGroup(ctx context.Context, userID, userGroupID int) error
	GetUsersInGroup(ctx context.Context, userGroupID int) ([]int, error)
}

// The server must be also able to group user groups instead of just users (Cycles are not allowed)
type Stage3 interface {
	Stage2
	AddUserGroupToGroup(ctx context.Context, childUserGroupID, parentUserGroupID int) error
	GetUserGroupsInGroup(ctx context.Context, userGroupID int) ([]int, error)
}

// Group membership is now defined as a transitive relation
type Stage4 interface {
	Stage3
	/*
		GetUsersInGroupTransitive returns the set of users that are:
		1) contained in the selected user group
		2) contained in any user group that is transitively contained in the selected user group
	*/
	GetUsersInGroupTransitive(ctx context.Context, userGroupID int) ([]int, error)
}

// In order to access a user or user group, one must have permission on it.
// A permission has a source and a target, both of which can be either a user or a user group.
type Stage5 interface {
	Stage4
	AddUserToUserPermission(ctx context.Context, sourceUserID, targetUserID int) error
	AddUserToUserGroupPermission(ctx context.Context, sourceUserID, targetUserGroupID int) error
	AddUserGroupToUserPermission(ctx context.Context, sourceUserGroupID, targetUserID int) error
	AddUserGroupToUserGroupPermission(ctx context.Context, sourceUserGroupID, targetUserGroupID int) error

	/*
		Getting the name of a user now requires that the user making the request has permission on the targeted user
		 User 0 has permission on user 1 in any of the following cases:
		 	1) User 0 directly has permission on user 1
			2) User 0 is transitively contained in a user group 0 (i.e. GetUsersInGroupTransitive for user group 0 contains user 0) and user group 0 directly has permission on user 1
		 	3) User 1 is transitively contained in a user group 0 and user 0 has permission on user group 0
		 	4) User 0 is transitively contained in a user group 0, user 1 is transitively contained in a user group 1 and user group 0 has permission on user group 1
	*/
	GetUserNameWithPermissionCheck(ctx context.Context, contextUserID, targetUserID int) (string, error)
	// If the target is a user group, the permission logic is analogous to the case where the target is a user
	GetUserGroupNameWithPermissionCheck(ctx context.Context, contextUserID, targetUserGroupID int) (string, error)
}
