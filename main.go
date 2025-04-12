package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Command struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

type CommandList struct {
	Commands []Command `yaml:"commands"`
}

const (
	AppName  = "Command Search Tool"
	Version  = "2.0.0"
	YAMLFile = "../commands.yml"
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

func readCommandsFromYAML() ([]Command, error) {
	// Get the directory of the executable and create the path to the YAML file
	execDir := getExecDir()
	filename := filepath.Join(execDir, YAMLFile)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cmdList CommandList
	if err := yaml.Unmarshal(data, &cmdList); err != nil {
		return nil, err
	}

	return cmdList.Commands, nil
}

// Function to write a new command to the YAML file
func writeCommandToYAML(input []string) error {
	if len(input) == 4 {
		newCommand := Command{
			Command:     strings.Trim(input[2], "\""),
			Description: strings.Trim(input[3], "\""),
		}

		// Get the directory of the executable and create the path to the YAML file
		execDir := getExecDir()
		filename := filepath.Join(execDir, YAMLFile)

		// Read existing commands
		commands, _ := readCommandsFromYAML()

		// Append new command
		commands = append(commands, newCommand)

		cmdList := CommandList{Commands: commands}
		data, err := yaml.Marshal(&cmdList)
		if err != nil {
			return err
		}

		return os.WriteFile(filename, data, 0644)
	}

	return fmt.Errorf("invalid input (size: %d), expected format: add \"command\" \"description\"", len(input))
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
			if err := writeCommandToYAML(args); err != nil {
				log.Fatalf("Failed to write to YAML: %v", err)
			}
			fmt.Println("Command added successfully.")
			return
		}
	}

	// Otherwise, it will search through
	commands, err := readCommandsFromYAML()
	if err != nil {
		log.Fatal(err)
	}

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
			command := wrapText(commands[i].Command, w/3)
			description := wrapText(commands[i].Description, w/3)
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
	if err := clipboard.WriteAll(selectedCommand); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nSelected command copied to clipboard: %s\n", selectedCommand)

	// Automatically paste the command to the terminal
	if err := pasteCommand(selectedCommand); err != nil {
		log.Fatal(err)
	}
}
