package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/nbvehbq/go-password-keeper/internal/client"
	"github.com/nbvehbq/go-password-keeper/internal/model"
	"github.com/nbvehbq/go-password-keeper/internal/storage"
)

type Keeper interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) error
	ListSecrets(ctx context.Context, resourceType string) ([]model.Secret, error)
	CreateSecret(ctx context.Context, data *model.Secret) (int64, error)
	GetSecret(ctx context.Context, ID int64) (*model.Secret, error)
	DeleteSecret(ctx context.Context, ID int64) error
	UpdateSecret(ctx context.Context, ID int64, data *model.Secret) (int64, error)
}

var (
	resorces = []string{"All", "Login & password", "Text", "Binary (file)", "Bank card"}
)

func SetupCommands(ctx context.Context, keeper Keeper) *ishell.Shell {
	shell := ishell.New()

	shell.Println(`Welcome to password-keeper utility!
	Write "help" for more information!\n`)

	// Register user cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "register",
		Help: "Register as new user & create public/private key pair",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			c.Print("Login: ")
			login := c.ReadLine()
			c.Print("Password: ")
			password := c.ReadPassword()

			if err := keeper.Register(ctx, login, password); err != nil {
				switch {
				case errors.Is(err, client.ErrUserExists):
					c.Println("User already exists. Try another login.")
				case errors.Is(err, client.ErrInternal):
					c.Println("Internal error. Try later.")
				default:
					c.Println("Unexpected error:", err)
				}
			}

			// TODO: Create public/private key pair
		},
	})

	// Login user cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "login",
		Help: "Login to password-keeper server component",
		Func: func(c *ishell.Context) {
			var err error
			for i := 0; i <= 2; i++ {
				c.ShowPrompt(false)
				defer c.ShowPrompt(true)

				c.Print("Login: ")
				login := c.ReadLine()
				c.Print("Password: ")
				password := c.ReadPassword()

				if err = keeper.Login(ctx, login, password); err == nil {
					c.Println("Authentication successful.")
					break
				}

				switch {
				case errors.Is(err, client.ErrUnauthorized):
					c.Println("Authentication Failed. Try again...")
					continue
				case errors.Is(err, client.ErrInternal):
					c.Println("Internal error. Try later.")
					break
				default:
					c.Println("Unexpected error:", err)
					break
				}
			}

			if err != nil {
				c.Println("Got error:", err)
			}
		},
	})

	// List secrets cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "List resources saved by current user",
		Func: func(c *ishell.Context) {
			choice := c.MultiChoice(resorces, "Witch resorce you want to list?")
			type_ := fmt.Sprintf("%d", choice)
			if choice == 0 {
				type_ = ""
			}
			list, err := keeper.ListSecrets(ctx, type_)
			if err != nil {
				switch {
				case errors.Is(err, client.ErrUnauthorized):
					c.Println("Please login first.")
				default:
					c.Println("Unexpected error:", err)
				}
				return
			}

			if len(list) == 0 {
				c.Println("No resources found.")
				return
			}

			text := make([]string, len(list))
			for i, v := range list {
				text[i] = fmt.Sprintf("| %4d | %10s |", v.ID, v.Name)
			}
			c.ShowPaged(strings.Join(text, "\n"))
		},
	})

	// Create secret cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "create",
		Help: "Create new resource",
		Func: func(c *ishell.Context) {
			choice := c.MultiChoice(resorces[1:], "Witch resorce you want to create?")
			// c.Println("You choose: ", choice)
			rtype := model.ResourceType(choice + 1)
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			var payload []byte
			var err error

			c.Print("Name: ")
			name := c.ReadLine()
			c.Print("Matadata: ")
			meta := c.ReadMultiLines("EOF")

			switch rtype {
			case model.LoginPasswordType:
				entity := &model.LoginPassword{}
				c.Print("Login: ")
				entity.Login = c.ReadLine()
				c.Print("Password: ")
				entity.Password = c.ReadPassword()
				payload, err = json.Marshal(entity)
				if err != nil {
					c.Println("Unexpected error:", err)
					return
				}
			case model.TextType:
				entity := &model.Text{}
				c.Print("Text: ")
				entity.Value = c.ReadMultiLines("EOF")
				payload, err = json.Marshal(entity)
				if err != nil {
					c.Println("Unexpected error:", err)
					return
				}
			case model.BinaryType:
				entity := &model.Binary{}
				c.Print("File path: ")
				path := c.ReadLine()
				entity.Name = path
				entity.Value, err = os.ReadFile(path)
				if err != nil {
					c.Println("Unexpected error:", err)
					return
				}
				payload, err = json.Marshal(entity)
				if err != nil {
					c.Println("Unexpected error:", err)
					return
				}
			case model.BankCardType:
				entity := &model.BankCard{}
				c.Print("Number: ")
				entity.Number = c.ReadLine()
				c.Print("ExpireAt: ")
				entity.ExpireAt = c.ReadLine()
				c.Print("Name: ")
				entity.Name = c.ReadLine()
				c.Print("Surname: ")
				entity.Surname = c.ReadLine()
				payload, err = json.Marshal(entity)
				if err != nil {
					c.Println("Unexpected error:", err)
					return
				}
			}

			id, err := keeper.CreateSecret(ctx, &model.Secret{
				Name:    name,
				Type:    model.ResourceType(choice + 1),
				Payload: payload,
				Meta:    []byte(meta),
			})
			if err != nil {
				switch {
				case errors.Is(err, client.ErrUnauthorized):
					c.Println("Please login first.")
				case errors.Is(err, storage.ErrSecretExists):
					c.Println("Secret already exists")
				default:
					c.Println("Unexpected error:", err)
				}
				return
			}

			c.Println("Secret created with ID:", id)
		},
	})

	// Get secret cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "Get resource by ID",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			c.Print("ID: ")
			id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err != nil {
				c.Println("Unexpected error:", err)
				return
			}

			secret, err := keeper.GetSecret(ctx, id)
			if err != nil {
				switch {
				case errors.Is(err, client.ErrUnauthorized):
					c.Println("Please login first.")
				case errors.Is(err, storage.ErrSecretNotFound):
					c.Println("Secret not found")
				default:
					c.Println("Unexpected error:", err)
				}
				return
			}

			c.Println("ID: ", secret.ID)
			c.Println("Name: ", secret.Name)
			c.Println("Type: ", secret.Type)

			// TODO: decrypt payload & meta

			c.Println("Metadata: ", string(secret.Meta))
			rtype := model.ResourceType(secret.Type)
			switch rtype {
			case model.LoginPasswordType:
				lp := model.LoginPassword{}
				if err = json.Unmarshal(secret.Payload, &lp); err != nil {
					c.Println("Unexpected error:", err)
					return
				}
				c.Println("Login: ", lp.Login)
				c.Println("Password: ", lp.Password)
			case model.TextType:
				t := model.Text{}
				if err = json.Unmarshal(secret.Payload, &t); err != nil {
					c.Println("Unexpected error:", err)
					return
				}
				c.Println("Text: ", t.Value)
			case model.BinaryType:
				b := model.Binary{}
				if err = json.Unmarshal(secret.Payload, &b); err != nil {
					c.Println("Unexpected error:", err)
					return
				}
				c.Println("Binary name: ", b.Name)
			case model.BankCardType:
				bc := model.BankCard{}
				if err = json.Unmarshal(secret.Payload, &bc); err != nil {
					c.Println("Unexpected error:", err)
					return
				}
				c.Println("Number: ", bc.Number)
				c.Println("Expire at: ", bc.ExpireAt)
				c.Println("Name: ", bc.Name)
				c.Println("Surname: ", bc.Surname)
			default:
				c.Println("Unknown type")
			}
		},
	})

	// Delete secret cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "Delete secret by ID",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			c.Print("ID: ")
			ID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err != nil {
				c.Println("Unexpected error:", err)
				return
			}

			err = keeper.DeleteSecret(ctx, ID)
			if err != nil {
				switch {
				case errors.Is(err, client.ErrUnauthorized):
					c.Println("Please login first.")
				case errors.Is(err, storage.ErrSecretNotFound):
					c.Println("Secret not found")
				default:
					c.Println("Unexpected error:", err)
				}
				return
			}

			c.Println("Secret deleted")
		},
	})

	// Update secret cmd
	shell.AddCmd(&ishell.Cmd{
		Name: "update",
		Help: "Update secret by ID",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			c.Print("ID: ")
			ID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err != nil {
				c.Println("Unexpected error:", err)
				return
			}

			secret, err := keeper.GetSecret(ctx, ID)
			if err != nil {
				switch {
				case errors.Is(err, client.ErrUnauthorized):
					c.Println("Please login first.")
				case errors.Is(err, storage.ErrSecretNotFound):
					c.Println("Secret not found")
				default:
					c.Println("Unexpected error:", err)
				}
				return
			}

			c.Println("ID: ", secret.ID)
			c.Println("Name: ", secret.Name)
			c.Println("Type: ", secret.Type)

			// TODO: decrypt payload & meta

			// c.ReadLineWithDefault()
		},
	})

	return shell
}
