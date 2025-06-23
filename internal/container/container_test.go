package container

import (
	"fmt"
	"testing"
)

type UserService struct {
}

type UserApi struct {
	userService *UserService `autowire:"-"`
}

func TestName(t *testing.T) {
	userApi := &UserApi{}
	err := Inject(userApi)
	if err != nil {
		t.Errorf("Failed to inject dependencies: %v", err)
		return
	}
	fmt.Printf("%+v\n", userApi)
}
