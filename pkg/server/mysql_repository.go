package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// SQL queries as package-level constants for better maintainability
const (
	queryInsertUser      = "INSERT INTO users (name) VALUES (?)"
	querySelectUser      = "SELECT name FROM users WHERE id = ?"
	queryInsertUserGroup = "INSERT INTO user_groups (name) VALUES (?)"
	querySelectUserGroup = "SELECT name FROM user_groups WHERE id = ?"

	queryInsertUserToGroup = `
		INSERT INTO user_group_members (user_id, user_group_id) 
		VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE user_id = user_id`

	querySelectUsersInGroup = `
		SELECT user_id 
		FROM user_group_members 
		WHERE user_group_id = ? 
		ORDER BY user_id`

	queryInsertGroupToGroup = `
		INSERT INTO user_group_hierarchy (child_group_id, parent_group_id) 
		VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE child_group_id = child_group_id`

	querySelectGroupsInGroup = `
		SELECT child_group_id 
		FROM user_group_hierarchy 
		WHERE parent_group_id = ? 
		ORDER BY child_group_id`

	queryCheckCycle = `
		WITH RECURSIVE descendants AS (
			SELECT child_group_id FROM user_group_hierarchy WHERE parent_group_id = ?
			UNION ALL
			SELECT h.child_group_id 
			FROM user_group_hierarchy h
			INNER JOIN descendants d ON h.parent_group_id = d.child_group_id
		)
		SELECT 1 FROM descendants WHERE child_group_id = ? LIMIT 1`

	querySelectUsersInGroupTransitive = `
		WITH RECURSIVE all_groups AS (
			SELECT ? as group_id
			UNION ALL
			SELECT h.child_group_id
			FROM user_group_hierarchy h
			INNER JOIN all_groups ag ON h.parent_group_id = ag.group_id
		)
		SELECT DISTINCT m.user_id
		FROM user_group_members m
		INNER JOIN all_groups ag ON m.user_group_id = ag.group_id
		ORDER BY m.user_id`

	queryInsertPermission = `
		INSERT INTO permissions (source_type, source_id, target_type, target_id) 
		VALUES (?, ?, ?, ?) 
		ON DUPLICATE KEY UPDATE source_id = source_id`

	queryCheckUserPermissionOnUser = `
		SELECT 1 FROM (
			-- Scenario 1: Direct user-to-user permission
			SELECT 1 as has_perm
			FROM permissions
			WHERE source_type = 'user' AND source_id = ?
			  AND target_type = 'user' AND target_id = ?
			
			UNION
			
			-- Scenario 2: Source user in group (transitively) -> target user
			SELECT 1 as has_perm
			FROM permissions p
			INNER JOIN (
				WITH RECURSIVE user_groups AS (
					SELECT user_group_id FROM user_group_members WHERE user_id = ?
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN user_groups ug ON h.child_group_id = ug.user_group_id
				)
				SELECT user_group_id FROM user_groups
			) source_groups ON p.source_id = source_groups.user_group_id
			WHERE p.source_type = 'group'
			  AND p.target_type = 'user' AND p.target_id = ?
			
			UNION
			
			-- Scenario 3: Source user -> target user in group (transitively)
			SELECT 1 as has_perm
			FROM permissions p
			INNER JOIN (
				WITH RECURSIVE user_groups AS (
					SELECT user_group_id FROM user_group_members WHERE user_id = ?
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN user_groups ug ON h.child_group_id = ug.user_group_id
				)
				SELECT user_group_id FROM user_groups
			) target_groups ON p.target_id = target_groups.user_group_id
			WHERE p.source_type = 'user' AND p.source_id = ?
			  AND p.target_type = 'group'
			
			UNION
			
			-- Scenario 4: Source user in group (transitively) -> target user in group (transitively)
			SELECT 1 as has_perm
			FROM permissions p
			INNER JOIN (
				WITH RECURSIVE user_groups AS (
					SELECT user_group_id FROM user_group_members WHERE user_id = ?
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN user_groups ug ON h.child_group_id = ug.user_group_id
				)
				SELECT user_group_id FROM user_groups
			) source_groups ON p.source_id = source_groups.user_group_id
			INNER JOIN (
				WITH RECURSIVE user_groups AS (
					SELECT user_group_id FROM user_group_members WHERE user_id = ?
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN user_groups ug ON h.child_group_id = ug.user_group_id
				)
				SELECT user_group_id FROM user_groups
			) target_groups ON p.target_id = target_groups.user_group_id
			WHERE p.source_type = 'group' AND p.target_type = 'group'
		) as perm_check
		LIMIT 1`

	queryCheckUserPermissionOnGroup = `
		SELECT 1 FROM (
			-- Scenario 1: Direct user-to-group permission
			SELECT 1 as has_perm
			FROM permissions
			WHERE source_type = 'user' AND source_id = ?
			  AND target_type = 'group' AND target_id = ?
			
			UNION
			
			-- Scenario 2: Source user in group (transitively) -> target group
			SELECT 1 as has_perm
			FROM permissions p
			INNER JOIN (
				WITH RECURSIVE user_groups AS (
					SELECT user_group_id FROM user_group_members WHERE user_id = ?
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN user_groups ug ON h.child_group_id = ug.user_group_id
				)
				SELECT user_group_id FROM user_groups
			) source_groups ON p.source_id = source_groups.user_group_id
			WHERE p.source_type = 'group'
			  AND p.target_type = 'group' AND p.target_id = ?
			
			UNION
			
			-- Scenario 3: Source user -> target group is transitively in another group
			SELECT 1 as has_perm
			FROM permissions p
			INNER JOIN (
				WITH RECURSIVE parent_groups AS (
					SELECT ? as group_id
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN parent_groups pg ON h.child_group_id = pg.group_id
				)
				SELECT group_id FROM parent_groups
			) target_groups ON p.target_id = target_groups.group_id
			WHERE p.source_type = 'user' AND p.source_id = ?
			  AND p.target_type = 'group'
			
			UNION
			
			-- Scenario 4: Source user in group (transitively) -> target group in group (transitively)
			SELECT 1 as has_perm
			FROM permissions p
			INNER JOIN (
				WITH RECURSIVE user_groups AS (
					SELECT user_group_id FROM user_group_members WHERE user_id = ?
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN user_groups ug ON h.child_group_id = ug.user_group_id
				)
				SELECT user_group_id FROM user_groups
			) source_groups ON p.source_id = source_groups.user_group_id
			INNER JOIN (
				WITH RECURSIVE parent_groups AS (
					SELECT ? as group_id
					UNION ALL
					SELECT h.parent_group_id
					FROM user_group_hierarchy h
					INNER JOIN parent_groups pg ON h.child_group_id = pg.group_id
				)
				SELECT group_id FROM parent_groups
			) target_groups ON p.target_id = target_groups.group_id
			WHERE p.source_type = 'group' AND p.target_type = 'group'
		) as perm_check
		LIMIT 1`
)

// MySQLRepository implements the Repository interface using MySQL
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQLRepository creates a new MySQL repository with the given database connection
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Helper methods to reduce repetition

// execInsert executes an insert query and returns the last insert ID
func (r *MySQLRepository) execInsert(ctx context.Context, query, errorMsg string, args ...interface{}) (int, error) {
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", errorMsg, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

// queryString queries a single string value with custom error handling for not found
func (r *MySQLRepository) queryString(ctx context.Context, query string, notFoundErr error, errorMsg string, args ...interface{}) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", notFoundErr
		}
		return "", fmt.Errorf("%s: %w", errorMsg, err)
	}
	return value, nil
}

// queryIDs queries a list of integer IDs
func (r *MySQLRepository) queryIDs(ctx context.Context, query, errorMsg string, args ...interface{}) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorMsg, err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan id: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ids, nil
}

// queryExists checks if a query returns any rows
func (r *MySQLRepository) queryExists(ctx context.Context, query, errorMsg string, args ...interface{}) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s: %w", errorMsg, err)
	}
	return true, nil
}

// CreateUser creates a new user and returns their ID
func (r *MySQLRepository) CreateUser(ctx context.Context, name string) (int, error) {
	return r.execInsert(ctx, queryInsertUser, "failed to create user", name)
}

// GetUserByID retrieves a user's name by their ID
func (r *MySQLRepository) GetUserByID(ctx context.Context, userID int) (string, error) {
	return r.queryString(ctx, querySelectUser, &UserNotFoundError{UserID: userID}, "failed to get user name", userID)
}

// CreateUserGroup creates a new user group and returns its ID
func (r *MySQLRepository) CreateUserGroup(ctx context.Context, name string) (int, error) {
	return r.execInsert(ctx, queryInsertUserGroup, "failed to create user group", name)
}

// GetUserGroupByID retrieves a user group's name by its ID
func (r *MySQLRepository) GetUserGroupByID(ctx context.Context, groupID int) (string, error) {
	return r.queryString(ctx, querySelectUserGroup, &UserGroupNotFoundError{UserGroupID: groupID}, "failed to get user group name", groupID)
}

// AddUserToGroup adds a user to a group
func (r *MySQLRepository) AddUserToGroup(ctx context.Context, userID, groupID int) error {
	_, err := r.db.ExecContext(ctx, queryInsertUserToGroup, userID, groupID)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

// GetUsersInGroup returns all users directly in the specified group
func (r *MySQLRepository) GetUsersInGroup(ctx context.Context, groupID int) ([]int, error) {
	return r.queryIDs(ctx, querySelectUsersInGroup, "failed to get users in group", groupID)
}

// GetUsersInGroupTransitive returns all users in the group and all nested subgroups
func (r *MySQLRepository) GetUsersInGroupTransitive(ctx context.Context, groupID int) ([]int, error) {
	return r.queryIDs(ctx, querySelectUsersInGroupTransitive, "failed to get users in group transitive", groupID)
}

// AddGroupToGroup adds a child group to a parent group with cycle detection
// Uses a database transaction to ensure atomicity of cycle check and insert
func (r *MySQLRepository) AddGroupToGroup(ctx context.Context, childID, parentID int) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // Rollback if not committed

	// Check for self-cycle
	if childID == parentID {
		return &CycleDetectedError{
			ChildGroupID:  childID,
			ParentGroupID: parentID,
		}
	}

	// Check for cycle within transaction
	var exists int
	err = tx.QueryRowContext(ctx, queryCheckCycle, childID, parentID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check for cycle: %w", err)
	}

	// If we found a row, it means adding this would create a cycle
	if err == nil {
		return &CycleDetectedError{
			ChildGroupID:  childID,
			ParentGroupID: parentID,
		}
	}

	// No cycle detected, insert the relationship
	_, err = tx.ExecContext(ctx, queryInsertGroupToGroup, childID, parentID)
	if err != nil {
		return fmt.Errorf("failed to add group to group: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetGroupsInGroup returns all groups directly in the specified group
func (r *MySQLRepository) GetGroupsInGroup(ctx context.Context, groupID int) ([]int, error) {
	return r.queryIDs(ctx, querySelectGroupsInGroup, "failed to get groups in group", groupID)
}

// WouldCreateCycle checks if adding child to parent would create a cycle
func (r *MySQLRepository) WouldCreateCycle(ctx context.Context, childID, parentID int) (bool, error) {
	// If they're the same, it's definitely a cycle
	if childID == parentID {
		return true, nil
	}

	return r.queryExists(ctx, queryCheckCycle, "failed to check for cycle", childID, parentID)
}

// AddPermission adds a permission record
func (r *MySQLRepository) AddPermission(ctx context.Context, sourceType, targetType string, sourceID, targetID int) error {
	_, err := r.db.ExecContext(ctx, queryInsertPermission, sourceType, sourceID, targetType, targetID)
	if err != nil {
		return fmt.Errorf("failed to add permission: %w", err)
	}

	return nil
}

// HasUserPermissionOnUser checks if a user has permission to access another user
func (r *MySQLRepository) HasUserPermissionOnUser(ctx context.Context, sourceUserID, targetUserID int) (bool, error) {
	return r.queryExists(ctx, queryCheckUserPermissionOnUser, "failed to check user permission on user",
		sourceUserID, targetUserID, // Scenario 1
		sourceUserID, targetUserID, // Scenario 2
		targetUserID, sourceUserID, // Scenario 3
		sourceUserID, targetUserID, // Scenario 4
	)
}

// HasUserPermissionOnGroup checks if a user has permission to access a group
func (r *MySQLRepository) HasUserPermissionOnGroup(ctx context.Context, sourceUserID, targetGroupID int) (bool, error) {
	return r.queryExists(ctx, queryCheckUserPermissionOnGroup, "failed to check user permission on group",
		sourceUserID, targetGroupID, // Scenario 1
		sourceUserID, targetGroupID, // Scenario 2
		targetGroupID, sourceUserID, // Scenario 3
		sourceUserID, targetGroupID, // Scenario 4
	)
}

// Close closes the database connection
func (r *MySQLRepository) Close() error {
	return r.db.Close()
}
