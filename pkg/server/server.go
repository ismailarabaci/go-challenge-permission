package server

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Server implements the Stage5 interface using a repository for data access
type Server struct {
	repo Repository
}

// New creates a new Server with the given repository.
// This constructor enforces Dependency Injection - all dependencies must be
// explicitly provided by the caller. This design choice ensures:
//   - Full control over dependencies (database, repository implementation)
//   - Easy testing with mock repositories
//   - Clear dependency graph
//   - No hidden side effects or internal resource creation
//
// Example usage:
//
//	config := server.DefaultConfig()
//	db, err := server.OpenDatabase(config)
//	if err != nil {
//	    return err
//	}
//	repo := server.NewMySQLRepository(db)
//	srv := server.New(repo)
//	defer srv.Close()
func New(repo Repository) *Server {
	return &Server{repo: repo}
}

// OpenDatabase creates and configures a database connection based on the provided config.
// This is a factory function that handles all database connection setup.
// It follows the Single Responsibility Principle by separating connection creation
// from Server construction.
func OpenDatabase(config Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", config.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Close closes the server and releases resources
func (s *Server) Close() error {
	if s.repo != nil {
		return s.repo.Close()
	}
	return nil
}

// CreateUser creates a new user and returns their ID
func (s *Server) CreateUser(ctx context.Context, name string) (int, error) {
	return s.repo.CreateUser(ctx, name)
}

// GetUserName retrieves a user's name by their ID
func (s *Server) GetUserName(ctx context.Context, userID int) (string, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// CreateUserGroup creates a new user group and returns its ID
func (s *Server) CreateUserGroup(ctx context.Context, name string) (int, error) {
	return s.repo.CreateUserGroup(ctx, name)
}

// GetUserGroupName retrieves a user group's name by its ID
func (s *Server) GetUserGroupName(ctx context.Context, userGroupID int) (string, error) {
	return s.repo.GetUserGroupByID(ctx, userGroupID)
}

// AddUserToGroup adds a user to a user group
func (s *Server) AddUserToGroup(ctx context.Context, userID, userGroupID int) error {
	return s.repo.AddUserToGroup(ctx, userID, userGroupID)
}

// GetUsersInGroup returns all users directly in the specified group
func (s *Server) GetUsersInGroup(ctx context.Context, userGroupID int) ([]int, error) {
	return s.repo.GetUsersInGroup(ctx, userGroupID)
}

// AddUserGroupToGroup adds a child group to a parent group
// Returns an error if this would create a cycle
// Uses a database transaction to ensure atomicity of cycle check and insert
func (s *Server) AddUserGroupToGroup(ctx context.Context, childUserGroupID, parentUserGroupID int) error {
	return s.repo.AddGroupToGroup(ctx, childUserGroupID, parentUserGroupID)
}

// GetUserGroupsInGroup returns all groups directly in the specified group
func (s *Server) GetUserGroupsInGroup(ctx context.Context, userGroupID int) ([]int, error) {
	return s.repo.GetGroupsInGroup(ctx, userGroupID)
}

// GetUsersInGroupTransitive returns all users in the group and all nested subgroups
func (s *Server) GetUsersInGroupTransitive(ctx context.Context, userGroupID int) ([]int, error) {
	return s.repo.GetUsersInGroupTransitive(ctx, userGroupID)
}

// AddUserToUserPermission grants a user permission to access another user
func (s *Server) AddUserToUserPermission(ctx context.Context, sourceUserID, targetUserID int) error {
	return s.repo.AddPermission(ctx, "user", "user", sourceUserID, targetUserID)
}

// AddUserToUserGroupPermission grants a user permission to access a user group
func (s *Server) AddUserToUserGroupPermission(ctx context.Context, sourceUserID, targetUserGroupID int) error {
	return s.repo.AddPermission(ctx, "user", "group", sourceUserID, targetUserGroupID)
}

// AddUserGroupToUserPermission grants a user group permission to access a user
func (s *Server) AddUserGroupToUserPermission(ctx context.Context, sourceUserGroupID, targetUserID int) error {
	return s.repo.AddPermission(ctx, "group", "user", sourceUserGroupID, targetUserID)
}

// AddUserGroupToUserGroupPermission grants a user group permission to access another user group
func (s *Server) AddUserGroupToUserGroupPermission(ctx context.Context, sourceUserGroupID, targetUserGroupID int) error {
	return s.repo.AddPermission(ctx, "group", "group", sourceUserGroupID, targetUserGroupID)
}

// GetUserNameWithPermissionCheck retrieves a user's name if the context user has permission
func (s *Server) GetUserNameWithPermissionCheck(ctx context.Context, contextUserID, targetUserID int) (string, error) {
	// Check if contextUser has permission to access targetUser
	hasPermission, err := s.repo.HasUserPermissionOnUser(ctx, contextUserID, targetUserID)
	if err != nil {
		return "", fmt.Errorf("failed to check permission: %w", err)
	}

	if !hasPermission {
		return "", &PermissionDeniedError{
			SourceUserID: contextUserID,
			TargetType:   "user",
			TargetID:     targetUserID,
		}
	}

	// If permission check passes, get the user name
	return s.GetUserName(ctx, targetUserID)
}

// GetUserGroupNameWithPermissionCheck retrieves a user group's name if the context user has permission
func (s *Server) GetUserGroupNameWithPermissionCheck(ctx context.Context, contextUserID, targetUserGroupID int) (string, error) {
	// Check if contextUser has permission to access targetUserGroup
	hasPermission, err := s.repo.HasUserPermissionOnGroup(ctx, contextUserID, targetUserGroupID)
	if err != nil {
		return "", fmt.Errorf("failed to check permission: %w", err)
	}

	if !hasPermission {
		return "", &PermissionDeniedError{
			SourceUserID: contextUserID,
			TargetType:   "group",
			TargetID:     targetUserGroupID,
		}
	}

	// If permission check passes, get the user group name
	return s.GetUserGroupName(ctx, targetUserGroupID)
}
