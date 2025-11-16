package server

import (
	"errors"
	"fmt"
)

// Sentinel errors for common cases
var (
	// ErrUserNotFound indicates that the requested user does not exist
	ErrUserNotFound = errors.New("user not found")

	// ErrUserGroupNotFound indicates that the requested user group does not exist
	ErrUserGroupNotFound = errors.New("user group not found")

	// ErrCycleDetected indicates that an operation would create a cycle in the group hierarchy
	ErrCycleDetected = errors.New("operation would create a cycle in group hierarchy")

	// ErrPermissionDenied indicates that the user does not have permission to perform the action
	ErrPermissionDenied = errors.New("permission denied")
)

// UserNotFoundError wraps user ID information
type UserNotFoundError struct {
	UserID int
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user not found: %d", e.UserID)
}

func (e *UserNotFoundError) Is(target error) bool {
	return target == ErrUserNotFound
}

// UserGroupNotFoundError wraps user group ID information
type UserGroupNotFoundError struct {
	UserGroupID int
}

func (e *UserGroupNotFoundError) Error() string {
	return fmt.Sprintf("user group not found: %d", e.UserGroupID)
}

func (e *UserGroupNotFoundError) Is(target error) bool {
	return target == ErrUserGroupNotFound
}

// CycleDetectedError wraps cycle information
type CycleDetectedError struct {
	ChildGroupID  int
	ParentGroupID int
}

func (e *CycleDetectedError) Error() string {
	return fmt.Sprintf("adding group %d to group %d would create a cycle", e.ChildGroupID, e.ParentGroupID)
}

func (e *CycleDetectedError) Is(target error) bool {
	return target == ErrCycleDetected
}

// PermissionDeniedError wraps permission denial information
type PermissionDeniedError struct {
	TargetType   string // "user" or "group"
	SourceUserID int
	TargetID     int
}

func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("user %d does not have permission to access %s %d", e.SourceUserID, e.TargetType, e.TargetID)
}

func (e *PermissionDeniedError) Is(target error) bool {
	return target == ErrPermissionDenied
}
