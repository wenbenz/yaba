package database_test

import (
	"golang.org/x/net/context"
	"testing"
	"yaba/internal/database"
	"yaba/internal/test/helper"
)

func TestCreateUser(t *testing.T) {
	pool := helper.GetTestPool()

	database.CreateUser(context.Background(), pool, "testUser", "testPassword")
}
