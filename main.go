package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"path/filepath"
	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
)

// Command struct to hold command and description
type Command struct {
	Command     string
	Description string
}

const (
    AppName    = "Command Search Tool"
    Version    = "1.0.0"
    CSVFile    = "../commands.csv"
)

// Gets the directory path of the executable binary 
func getExecDir() string {
	// Get the path of the executable
    execPath, err := os.Executable()
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

	return filepath.Dir(execPath)
}

// Function to wrap text to a specific width
func wrapText(text string, width int) string {
    if len(text) <= width {
        return text
    }
    var wrappedText string
    for len(text) > width {
        splitPos := strings.LastIndex(text[:width], " ")
        if splitPos == -1 {
            splitPos = width
        }
        wrappedText += text[:splitPos] + "\n"
        text = strings.TrimSpace(text[splitPos:])
    }
    wrappedText += text
    return wrappedText
}

// Function to read commands from CSV file
func readCommandsFromCSV() ([]Command, error) {
    // Get the directory of the executable and create the path to the CSV file
    execDir := getExecDir()
    filename := filepath.Join(execDir, CSVFile)

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

// Function to write a new command to the CSV file
func writeCommandToCSV(input []string) error {
	if len(input) == 4 {
		command := strings.Trim(input[2], "\"")
		description := strings.Trim(input[3], "\"")

		// Get the directory of the executable and create the path to the CSV file
		execDir := getExecDir()
		filename := filepath.Join(execDir, CSVFile)

		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return err
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		return writer.Write([]string{command, description})
	}
	
	return fmt.Errorf("invalid input (size: %d), needs to be of format: add \"command\" \"description\"", len(input))
}

func pasteCommand(command string) error {
	// Detect the OS and run the appropriate paste command
	switch os := runtime.GOOS; os {
	case "linux":
		cmd := exec.Command("xdotool", "type", "--delay", "1", command)
		return cmd.Run()
	case "darwin":
		script := `tell application "System Events" to keystroke "v" using {command down}`
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported OS: %s", os)
	}
}

func main() {
	// Checking for arguments (e.g. adding new commands). Exits program after
	args := os.Args
	if len(args) > 1 {
		switch args[1] {
		case "add":
			if err := writeCommandToCSV(args); err != nil {
				log.Fatalf("Failed to write to CSV: %v", err)
			}
			return
		}
	}

	// Otherwise, it will search through
	commands, err := readCommandsFromCSV()
	if err != nil {
		log.Fatal(err)
	}

	// Use fuzzyfinder to search and select a command
	idx, err := fuzzyfinder.Find(
		commands,
		func(i int) string {
			// return fmt.Sprintf("%s: %s", commands[i].Command, commands[i].Description)
			return commands[i].Command

		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			
			// w is the width of the entire terminal, so the wrapping needs to be
			// less than half of w to fit into the preview window
			command := wrapText(commands[i].Command, w / 3)
			description := wrapText(commands[i].Description, w / 3)

			return fmt.Sprintf("Command: %s\n\nDescription: %s", command, description)
		}),
	)
	// Quitting with ctrl + c
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
