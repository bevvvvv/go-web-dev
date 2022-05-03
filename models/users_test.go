package models

import (
	"fmt"
	"testing"
	"time"
)

func testingUserService() (*UserService, error) {
	const (
		host     = "host.docker.internal"
		port     = 5432
		user     = "postgres"
		password = "secretpass"
		dbname   = "fakeoku_test"
	)

	connectionInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	userService, err := NewUserService(connectionInfo)
	if err != nil {
		return nil, err
	}
	userService.db.LogMode(false)

	// Clear table
	userService.DestructiveReset()
	return userService, nil
}

func TestCreateUser(t *testing.T) {
	userService, err := testingUserService()
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		Name:  "Michael Scott",
		Email: "michael@dundermifflin.net",
	}
	err = userService.Create(&user)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Errorf("Expected ID > 0. Receieved %d", user.ID)
	}
	if time.Since(user.CreatedAt) > time.Duration(5*time.Second) {
		t.Errorf("Expected recent creation. Received %s", user.CreatedAt)
	}
	if time.Since(user.UpdatedAt) > time.Duration(5*time.Second) {
		t.Errorf("Expected recent creation. Received %s", user.UpdatedAt)
	}
}

func TestUpdateUser(t *testing.T) {
	userService, err := testingUserService()
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		Name:  "Michael Scott",
		Email: "michael@dundermifflin.net",
	}
	err = userService.Create(&user)
	if err != nil {
		t.Fatal(err)
	}
	user.Email = "michael@michaelscottpaperco.com"
	err = userService.Update(&user)
	if err != nil {
		t.Fatal(err)
	}
	dbUser, err := userService.ById(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dbUser.Email != user.Email {
		t.Errorf("Expected email michael@michaelscottpaperco.com. Received %s", dbUser.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	userService, err := testingUserService()
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		Name:  "Michael Scott",
		Email: "michael@dundermifflin.net",
	}
	err = userService.Create(&user)
	if err != nil {
		t.Fatal(err)
	}
	userService.Delete(1)
	dbUser, err := userService.ById(user.ID)
	if dbUser.Name != "" {
		t.Errorf("Expected user to be deleted. Found %s", dbUser.Name)
	}
}
