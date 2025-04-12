package main

import (
	"bufio"
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

func getYAMLPath() string {
	execPath, _ := os.Executable()
	return filepath.Join(filepath.Dir(execPath), YAMLFile)
}

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

func readCommands() ([]Command, error) {
	data, err := os.ReadFile(getYAMLPath())
	if err != nil {
		return nil, err
	}
	var cmdList CommandList
	if err := yaml.Unmarshal(data, &cmdList); err != nil {
		return nil, err
	}
	return cmdList.Commands, nil
}

func writeCommands(commands []Command) error {
	data, err := yaml.Marshal(CommandList{Commands: commands})
	if err != nil {
		return err
	}
	return os.WriteFile(getYAMLPath(), data, 0644)
}

func pasteCommand(command string) error {
	switch os := runtime.GOOS; os {
	case "linux":
		return exec.Command("xdotool", "type", "--delay", "1", command).Run()
	case "darwin":
		script := `tell application "System Events" to keystroke "v" using {command down}`
		return exec.Command("osascript", "-e", script).Run()
	default:
		return fmt.Errorf("unsupported OS: %s", os)
	}
}

func promptEdit(c Command) Command {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Edit command [%s]: ", c.Command)
	cmdInput, _ := reader.ReadString('\n')
	cmdInput = strings.TrimSpace(cmdInput)
	if cmdInput != "" {
		c.Command = cmdInput
	}

	fmt.Printf("Edit description [%s]: ", c.Description)
	descInput, _ := reader.ReadString('\n')
	descInput = strings.TrimSpace(descInput)
	if descInput != "" {
		c.Description = descInput
	}
	return c
}

func main() {
	// Add new command from args (e.g. `add "ls -la" "List files"`)
	args := os.Args
	if len(args) == 4 && args[1] == "add" {
		cmd := Command{Command: args[2], Description: args[3]}
		cmds, _ := readCommands()
		cmds = append(cmds, cmd)
		_ = writeCommands(cmds)
		fmt.Println("Command added.")
		return
	}

	cmds, err := readCommands()
	if err != nil {
		log.Fatal(err)
	}

	// Fuzzy select
	index, err := fuzzyfinder.Find(
		cmds,
		func(i int) string { return cmds[i].Command },
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 {
				return ""
			}
			// w is the width of the entire terminal, so the wrapping needs to be
			// less than half of w to fit into the preview window
			cmd := wrapText(cmds[i].Command, w/3)
			desc := wrapText(cmds[i].Description, w/3)
			return fmt.Sprintf("Command: %s\n\nDescription: %s", cmd, desc)
		}),
	)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Search aborted")
			return
		} else {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	selected := cmds[index]
	fmt.Printf("\nSelected:\n%s - %s\n", selected.Command, selected.Description)

	fmt.Print("Press [Enter] to copy/paste, [d] to delete, [e] to edit: ")
	choice := ""
	fmt.Scanln(&choice)

	switch choice {
	case "d":
		cmds = append(cmds[:index], cmds[index+1:]...)
		if err := writeCommands(cmds); err != nil {
			log.Fatal("Delete failed:", err)
		}
		fmt.Println("Deleted successfully.")
	case "e":
		cmds[index] = promptEdit(cmds[index])
		if err := writeCommands(cmds); err != nil {
			log.Fatal("Edit failed:", err)
		}
		fmt.Println("Edited and saved.")
	default:
		if err := clipboard.WriteAll(selected.Command); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Copied to clipboard: %s\n", selected.Command)
		if err := pasteCommand(selected.Command); err != nil {
			log.Fatal(err)
		}
	}
}
