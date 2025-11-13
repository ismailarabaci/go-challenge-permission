package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var errNotImplemented = errors.New("not implemented")

type server struct {
	db *sql.DB
}

func New() *server {
	db, err := sql.Open("mysql", "blp:password@tcp(localhost:3306)/blp-coding-challenge")
	if err != nil {
		panic(fmt.Sprintf("sql.Open: %s", err.Error()))
	}

	return &server{
		db: db,
	}
}

func (s *server) CreateUser(ctx context.Context, name string) (int, error) {
	return 0, errNotImplemented
}

func (s *server) GetUserName(ctx context.Context, userID int) (string, error) {
	return "", errNotImplemented
}

func (s *server) CreateUserGroup(ctx context.Context, name string) (int, error) {
	return 0, errNotImplemented
}

func (s *server) GetUserGroupName(ctx context.Context, userGroupID int) (string, error) {
	return "", errNotImplemented
}

func (s *server) AddUserToGroup(ctx context.Context, userID, userGroupID int) error {
	return errNotImplemented
}

func (s *server) GetUsersInGroup(ctx context.Context, userGroupID int) ([]int, error) {
	return nil, errNotImplemented
}

func (s *server) AddUserGroupToGroup(ctx context.Context, childUserGroupID, parentUserGroupID int) error {
	return errNotImplemented
}

func (s *server) GetUserGroupsInGroup(ctx context.Context, userGroupID int) ([]int, error) {
	return nil, errNotImplemented
}

func (s *server) GetUsersInGroupTransitive(ctx context.Context, userGroupID int) ([]int, error) {
	return nil, errNotImplemented
}

func (s *server) AddUserToUserPermission(ctx context.Context, sourceUserID, targetUserID int) error {
	return errNotImplemented
}

func (s *server) AddUserToUserGroupPermission(ctx context.Context, sourceUserID, targetUserGroupID int) error {
	return errNotImplemented
}

func (s *server) AddUserGroupToUserPermission(ctx context.Context, sourceUserGroupID, targetUserID int) error {
	return errNotImplemented
}

func (s *server) AddUserGroupToUserGroupPermission(ctx context.Context, sourceUserGroupID, targetUserGroupID int) error {
	return errNotImplemented
}

func (s *server) GetUserNameWithPermissionCheck(ctx context.Context, contextUserID, targetUserID int) (string, error) {
	return "", errNotImplemented
}

func (s *server) GetUserGroupNameWithPermissionCheck(ctx context.Context, contextUserID, targetUserGroupID int) (string, error) {
	return "", errNotImplemented
}
