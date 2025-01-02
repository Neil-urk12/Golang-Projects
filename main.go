package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

var pl = fmt.Println

type Account struct {
	username      string
	password      string
	balance       float64
	accountType   string
	accountNumber int
}

var accounts []Account

func main() {
	reader := bufio.NewReader(os.Stdin)
	pl("==Welcome to Golangking System!==")
	loginScreen(reader)
}

func mainMenu(reader *bufio.Reader, account Account) {
	for {
		pl("1. Check Balance  2. Deposit  3. Withdraw  4. Fund Transfer  5. Logout")
		pl("Enter your choice : ")
		choice, err := reader.ReadByte()
		if err != nil {
			pl("Invalid input!\nPlease try again!")
			continue
		}

		switch choice {
		case '1':
			checkBalance(account)
		case '2':
			deposit(reader, account)
		case '3':
			withdraw(reader, account)
		case '4':
			fundTransfer(reader, account)
		case '5':
			pl("Logging out...")
			loginScreen(reader)
			return
		default:
			pl("Invalid input!\nPlease try again!")
			continue
		}
	}
}

func deposit(reader *bufio.Reader, account Account) {
	for {
		pl("Enter the amount to deposit : ")
		input, err := reader.ReadString('\n')

		if err != nil {
			pl("Invalid input!\nPlease try again!")
			return
		}

		amount, err := strconv.ParseFloat(input, 64)
		if err != nil {
			pl("Invalid input!\nPlease try again!")
			continue
		}

		account.balance += amount
		pl("Deposit successful!")
		pl("Your new balance is : ", account.balance)
	}
}

func checkBalance(account Account) {
	pl("Your balance is : ", account.balance)
}

func loginScreen(reader *bufio.Reader) {
	pl("1. Login  2. Register  3. Exit")
	pl("Enter you choice: ")
	choice, err := reader.ReadByte()
	if err != nil {
		pl("Invalid input")
		return
	}
	switch choice {
	case '1':
		login(reader)
	case '2':
		register(reader)
	case '3':
		pl("Exiting...")
		return
	default:
		pl("Invalid input")
	}
}

func login(reader *bufio.Reader) {
	pl("Login")
	pl("Enter your username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		pl("Invalid input", err)
		return
	}
	pl("Enter your password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		pl("Invalid input", err)
		return
	}
	pl(username, password)
	for _, account := range accounts {
		if account.username == username {
			if account.password == password {
				pl("Login successful!")
				pl("Welcome", username)
				mainMenu(reader, account)
			} else {
				pl("Incorrect password!\nPlease try again!")
				return
			}
		}
	}
}

func register(reader *bufio.Reader) {
	for {
		pl("Register")

		pl("Enter your username (min 8 characters): ")
		username, err := reader.ReadString('\n')
		if err != nil {
			pl("Invalid input", err)
			continue
		} else if len(username) < 8 {
			pl("Username must be at least 8 characters long!\nPlease try again!")
			continue
		}

		pl("Enter your password (min 8 characters) : ")
		password, err := reader.ReadString('\n')
		if err != nil {
			pl("Invalid input", err)
			continue
		} else if len(password) < 8 {
			pl("Password must be at least 8 characters long!\nPlease try again!")
			continue
		}

		pl("Enter you password again : ")
		password2, err := reader.ReadString('\n')
		if err != nil {
			pl("Invalid input", err)
			continue
		} else if len(password2) < 8 {
			pl("Password must be at least 8 characters long!\nPlease try again!")
			continue
		} else if password != password2 {
			pl("Passwords do not match!\nPlease try again!")
			continue
		}

		pl("Enter your account type (savings/checkings) : ")
		accountType, err := reader.ReadString('\n')
		if err != nil {
			pl("Invalid input", err)
			continue
		}
		accountNumber := accountNumberGenerator()
		balance := 0.0

		newAccount := Account{
			username:      username,
			password:      password,
			balance:       balance,
			accountType:   accountType,
			accountNumber: accountNumber,
		}
		pl("Account created successfully!")
		pl("Your account number is: ", accountNumber)
		accounts = append(accounts, newAccount)
		break
	}
	return
}

func accountNumberGenerator() int {
	return rand.Intn(1000000)
}
