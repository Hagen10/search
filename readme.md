# Search Tool

This tool exists because quite often I need to use a command that I don't remember correctly and fail to retrieve with reverse search or by looking through history. Instead, this tool allows you to save all commands in a yaml file for easy retrieval later on. The tool has two ways of being used from a terminal:

- executing `search` will open a list of all the commands currently stored in commands.yml. Selecting one and pressing enter will exit the program and paste it in the terminal
- executing `search add command description` will save a new command as well as a description to the commands.yml. Example: `search add "git checkout master" "checkouts the master branch of the current git directory"`

## Install

The tool can be installed with `make` once you have Go installed.