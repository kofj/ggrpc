package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"

	apiUsers "github.com/kofj/ggrpc/api/users"
	"github.com/peterh/liner"
)

const server = "127.0.0.1:2333"
const DefaultPrompt = "> "
const help = "new - create a new user\nfind - find user by id.\nquit - exit."

func main() {
	// connect to gRPC server
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot connect to server:", err)
	}
	defer conn.Close()
	client := apiUsers.NewUserClient(conn)

	// command user interface
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	var input string
	var user = &apiUsers.UserInfo{}
	var uid int32
	fmt.Println(help)
mainloop:
	for {
		if input, err = line.Prompt(DefaultPrompt); err == liner.ErrPromptAborted {
			log.Print("Aborted\n")
			break mainloop
		} else if err != nil {
			log.Print("Error reading line: ", err)
		}

		switch input {
		case "quit", "q":
			break mainloop

		case "new", "n":
			fmt.Println("Please answer questions to create a new user.")
			user.Name, _ = line.Prompt("user name is? ")
			user.Email, _ = line.Prompt("email is? ")
			for {
				var phone, isOffice, more string
				phone, _ = line.Prompt("give a phone number: ")
				isOffice, _ = line.Prompt("is a office phone? Yes/No: ")
				more, _ = line.Prompt("give more phones? Yes/No: ")
				user.Phones = append(user.Phones, &apiUsers.UserInfo_Phone{
					Number:   phone,
					IsOffice: isYes(isOffice),
				})
				if !isYes(more) {
					break
				}
			}
			createUser(client, user)

		case "find", "f":
			id, _ := line.Prompt("Please input user id to find:")
			fmt.Sscan(id, &uid)
			if uid < 1 {
				fmt.Println("invalid user id, try later.")
				continue
			}
			findUser(client, uid)

		case "help", "h":
			fmt.Println(help)
		case "":
			continue
		default:
			fmt.Println(help)
		}
	}

}

// Call CreateUser function on Server via gPRC
func createUser(client apiUsers.UserClient, user *apiUsers.UserInfo) {
	resp, err := client.CreateUser(context.Background(), user)
	if err != nil {
		log.Println("Could not create user:", err)
		return
	}
	if resp.Success {
		log.Println("A new user has been added, id:", resp.Id)
	}
}

// Call GetUser function on Server via gPRC
func findUser(client apiUsers.UserClient, uid int32) {
	filter := &apiUsers.UserFilter{Id: uid}
	stream, err := client.GetUser(context.Background(), filter)
	if err != nil {
		log.Println("failed on get user info:", err)
		return
	}
	for {
		user, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("%v.GetPersons(_) = _, %v", client, err)
		}
		fmt.Println(user)
	}
}

func isYes(input string) bool {
	switch input {
	case "YES", "Yes", "yes", "y":
		return true
	default:
		return false
	}
}
