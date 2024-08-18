package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/thomas-introini/pocket-cli/config"
	"github.com/thomas-introini/pocket-cli/db"
	"github.com/thomas-introini/pocket-cli/globals"
	"github.com/thomas-introini/pocket-cli/views/root"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		os.Exit(2)
	}
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "info")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	pocketConsumerKey := os.Getenv("POCKET_CONSUMER_KEY")
	if pocketConsumerKey == "" {
		fmt.Println("set POCKET_CONSUMER_KEY environment variable")
		os.Exit(1)
	}
	config.InitConfig(pocketConsumerKey)
	err = db.ConnectDB()
	if err != nil {
		fmt.Println("error connecting to database:", err)
		os.Exit(1)
	}
	user, err := db.GetLoggedUser()
	if err != nil && err != db.NoUserErr {
		fmt.Println("Error while retrieving user from database:", err)
		os.Exit(1)
	}
	p := tea.NewProgram(root.New(user), tea.WithAltScreen(), tea.WithMouseCellMotion())
	globals.InitProgram(p)
	if _, err = p.Run(); err != nil {
		fmt.Println("Could not run the program", err)
		os.Exit(1)
	}
}
