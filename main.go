package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"
)

// Структура для хранения команд - карта, где ключи - имена команд, а значения - функции
type Shell struct {
	commands map[string]func([]string)
}

func NewShell() *Shell {
	shell := &Shell{}
	shell.commands = map[string]func([]string){
		"ls":   shell.lsCommand,
		"cd":   shell.cdCommand,
		"exit": shell.exitCommand,
	}
	return shell
}
func (s *Shell) lsCommand(args []string) {
	fmt.Printf("Command: ls, arguments: %v\n", args)
}
func (s *Shell) cdCommand(args []string) {
	fmt.Printf("Command: cd, arguments: %v\n", args)
}
func (s *Shell) exitCommand(args []string) {
	os.Exit(0)
}
func (s *Shell) executeCommand(cmd string, args []string) {
	if handler, exists := s.commands[cmd]; exists {
		handler(args)
	} else {
		fmt.Printf("Command doesn`t exists\n")
	}
}

// Кастомное приглашение к вводу
func getInvitation() string {
	// Имя пользователя
	var username string
	currentUser, err := user.Current()
	if err != nil {
		username = os.Getenv("USER")
	} else {
		username = currentUser.Username
	}
	// Имя хоста
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	return fmt.Sprintf("%s@%s:~$ ", username, hostname)
}

// Парсер, который обрабатывает аргументы в кавычках
func parser(cmd string) (string, []string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "", nil
	}
	// регулярное выражение для разбивки с учетом кавычек
	// [^\s"']+ - последовательность символов без кавычек
	// "([^"]*)" - последовательность в двойных кавычках
	// '([^']*)' - в одинарных
	re := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	matches := re.FindAllString(cmd, -1)
	if len(matches) == 0 {
		return "", nil
	}
	arg := make([]string, 0, len(matches)-1)
	for i, match := range matches {
		if i == 0 {
			continue
		}
		if len(match) >= 2 && (match[0] == '"' && match[len(match)-1] == '"' || match[0] == '\'' && match[len(match)-1] == '\'') {
			match = match[1 : len(match)-1]
		}
		arg = append(arg, match)
	}

	return matches[0], arg
}

func main() {
	shell := NewShell()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Кастомное приглашение к вводу
		fmt.Print(getInvitation())
		if !scanner.Scan() {
			break
		}
		// считывание ввода
		input := scanner.Text()
		cmd, args := parser(input)

		shell.executeCommand(cmd, args)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения ввода: %v\n", err)
	}
}
