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

// CreateUser creates a new user and returns their ID
func (r *MySQLRepository) CreateUser(ctx context.Context, name string) (int, error) {
	result, err := r.db.ExecContext(ctx, queryInsertUser, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

// GetUserByID retrieves a user's name by their ID
func (r *MySQLRepository) GetUserByID(ctx context.Context, userID int) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, querySelectUser, userID).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", &UserNotFoundError{UserID: userID}
		}
		return "", fmt.Errorf("failed to get user name: %w", err)
	}

	return name, nil
}

// CreateUserGroup creates a new user group and returns its ID
func (r *MySQLRepository) CreateUserGroup(ctx context.Context, name string) (int, error) {
	result, err := r.db.ExecContext(ctx, queryInsertUserGroup, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create user group: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

// GetUserGroupByID retrieves a user group's name by its ID
func (r *MySQLRepository) GetUserGroupByID(ctx context.Context, groupID int) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, querySelectUserGroup, groupID).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", &UserGroupNotFoundError{UserGroupID: groupID}
		}
		return "", fmt.Errorf("failed to get user group name: %w", err)
	}

	return name, nil
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
	rows, err := r.db.QueryContext(ctx, querySelectUsersInGroup, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users in group: %w", err)
	}
	defer rows.Close()

	userIDs := make([]int, 0)
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return userIDs, nil
}

// GetUsersInGroupTransitive returns all users in the group and all nested subgroups
func (r *MySQLRepository) GetUsersInGroupTransitive(ctx context.Context, groupID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, querySelectUsersInGroupTransitive, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users in group transitive: %w", err)
	}
	defer rows.Close()

	userIDs := make([]int, 0)
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return userIDs, nil
}

// AddGroupToGroup adds a child group to a parent group
func (r *MySQLRepository) AddGroupToGroup(ctx context.Context, childID, parentID int) error {
	_, err := r.db.ExecContext(ctx, queryInsertGroupToGroup, childID, parentID)
	if err != nil {
		return fmt.Errorf("failed to add group to group: %w", err)
	}

	return nil
}

// GetGroupsInGroup returns all groups directly in the specified group
func (r *MySQLRepository) GetGroupsInGroup(ctx context.Context, groupID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, querySelectGroupsInGroup, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups in group: %w", err)
	}
	defer rows.Close()

	groupIDs := make([]int, 0)
	for rows.Next() {
		var gid int
		if err := rows.Scan(&gid); err != nil {
			return nil, fmt.Errorf("failed to scan group id: %w", err)
		}
		groupIDs = append(groupIDs, gid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return groupIDs, nil
}

// WouldCreateCycle checks if adding child to parent would create a cycle
func (r *MySQLRepository) WouldCreateCycle(ctx context.Context, childID, parentID int) (bool, error) {
	// If they're the same, it's definitely a cycle
	if childID == parentID {
		return true, nil
	}

	var exists int
	err := r.db.QueryRowContext(ctx, queryCheckCycle, childID, parentID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check for cycle: %w", err)
	}

	return true, nil
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
	var exists int
	err := r.db.QueryRowContext(ctx, queryCheckUserPermissionOnUser,
		sourceUserID, targetUserID, // Scenario 1
		sourceUserID, targetUserID, // Scenario 2
		targetUserID, sourceUserID, // Scenario 3
		sourceUserID, targetUserID, // Scenario 4
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check user permission on user: %w", err)
	}

	return true, nil
}

// HasUserPermissionOnGroup checks if a user has permission to access a group
func (r *MySQLRepository) HasUserPermissionOnGroup(ctx context.Context, sourceUserID, targetGroupID int) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, queryCheckUserPermissionOnGroup,
		sourceUserID, targetGroupID, // Scenario 1
		sourceUserID, targetGroupID, // Scenario 2
		targetGroupID, sourceUserID, // Scenario 3
		sourceUserID, targetGroupID, // Scenario 4
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check user permission on group: %w", err)
	}

	return true, nil
}

// Close closes the database connection
func (r *MySQLRepository) Close() error {
	return r.db.Close()
}
