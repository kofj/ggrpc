package main

import (
	"context"
	"log"
	"net"

	apiUsers "github.com/kofj/ggrpc/api/users"
	"google.golang.org/grpc"
)

const address = ":2333"

func main() {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("failed to listen", err)
	}
	log.Println("listening", address)

	s := grpc.NewServer()
	apiUsers.RegisterUserServer(s, &Users{})
	log.Println(s.Serve(l))
}

// Users implement apiUsers.UserServer interface.
type Users struct {
	users []*apiUsers.UserInfo
}

func (u *Users) GetUser(filter *apiUsers.UserFilter, stream apiUsers.User_GetUserServer) error {
	log.Println("find user via id:", filter.Id)
	for _, user := range u.users {
		if filter.Id == user.Id {
			err := stream.Send(user)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *Users) CreateUser(c context.Context, user *apiUsers.UserInfo) (state *apiUsers.CreateUserState, err error) {
	user.Id = int32(len(u.users) + 1)
	u.users = append(u.users, user)
	log.Println("new user:", user.Id)
	return &apiUsers.CreateUserState{Id: user.Id, Success: true}, nil
}

var _ apiUsers.UserServer = &Users{}
