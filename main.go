package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	// "strings"

	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
)

// Command struct to hold command and description
type Command struct {
	Command     string
	Description string
}

// List of commands
// var commands = []Command{
// 	{"ls", "List directory contents"},
// 	{"cd", "Change the current directory"},
// 	{"grep", "Search text using patterns"},
// 	{"awk", "Pattern scanning and processing language"},
// 	{"sed", "Stream editor for filtering and transforming text"},
// 	{"find", "Search for files in a directory hierarchy"},
// 	{"tar", "Archive files"},
// 	{"curl", "Transfer data from or to a server"},
// }

// Function to read commands from CSV file
func readCommandsFromCSV(filename string) ([]Command, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var commands []Command
	for _, record := range records[1:] { // skip the header
		commands = append(commands, Command{
			Command:     record[0],
			Description: record[1],
		})
	}
	return commands, nil
}

// func main() {
// 	// Use fuzzyfinder to search and select a command
// 	idx, err := fuzzyfinder.Find(
// 		commands,
// 		func(i int) string {
// 			return fmt.Sprintf("%s: %s", commands[i].Command, commands[i].Description)
// 		},
// 		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
// 			if i == -1 {
// 				return ""
// 			}
// 			return fmt.Sprintf("Command: %s\nDescription: %s", commands[i].Command, commands[i].Description)
// 		}),
// 	)
// 	if err != nil {
// 		if err == fuzzyfinder.ErrAbort {
// 			fmt.Println("Search aborted")
// 			os.Exit(0)
// 		} else {
// 			log.Fatal(err)
// 		}
// 	}

// 	selectedCommand := commands[idx].Command

// 	// Copy the selected command to the clipboard
// 	err = clipboard.WriteAll(selectedCommand)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("\nSelected command copied to clipboard: %s\n", selectedCommand)

// 	// Automatically paste the command to the terminal using AppleScript
// 	script := fmt.Sprintf(`tell application "System Events" to keystroke "v" using {command down}`)
// 	cmd := exec.Command("osascript", "-e", script)
// 	err = cmd.Run()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func main() {
	// Check for arguments
	args := os.Args
	var filter string
	if len(args) > 1 {
		filter = args[1]

		switch filter {
		case "add":

			return
		}
	}

	commands, err := readCommandsFromCSV("commands.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Use fuzzyfinder to search and select a command
	idx, err := fuzzyfinder.Find(
		commands,
		func(i int) string {
			return fmt.Sprintf("%s: %s", commands[i].Command, commands[i].Description)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Command: %s\nDescription: %s", commands[i].Command, commands[i].Description)
		}),
	)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Search aborted")
			os.Exit(0)
		} else {
			log.Fatal(err)
		}
	}

	selectedCommand := commands[idx].Command

	// Copy the selected command to the clipboard
	err = clipboard.WriteAll(selectedCommand)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nSelected command copied to clipboard: %s\n", selectedCommand)

	// Automatically paste the command to the terminal
	if err := pasteCommand(selectedCommand); err != nil {
		log.Fatal(err)
	}
}

func pasteCommand(command string) error {
	// Detect the OS and run the appropriate paste command
	switch os := runtime.GOOS; os {
	case "linux":
		cmd := exec.Command("xdotool", "type", "--delay", "1", command)
		return cmd.Run()
	case "darwin":
		script := fmt.Sprintf(`tell application "System Events" to keystroke "v" using {command down}`)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported OS: %s", os)
	}
}

// Automatically paste the command to the terminal using AppleScript
// 	script := fmt.Sprintf(`tell application "System Events" to keystroke "v" using {command down}`)
// 	cmd := exec.Command("osascript", "-e", script)
// 	err = cmd.Run()
// 	if err != nil {
// 		log.Fatal(err)
// 	}