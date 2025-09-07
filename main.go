package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/TimofeyChernyshev/MIREA-Configuration-management-1/vfs"
)

// Структура для хранения команд - карта, где ключи - имена команд, а значения - функции
type Shell struct {
	commands map[string]func([]string)
	vfs      vfs.VFS
}

func NewShell() *Shell {
	shell := &Shell{}
	shell.vfs = vfs.VFS{
		Root: &vfs.VFSNode{
			Name:     "/",
			IsDir:    true,
			ModTime:  time.Now(),
			Children: []*vfs.VFSNode{},
		},
		IsLoaded: false,
	}
	shell.commands = map[string]func([]string){
		"ls":       shell.lsCommand,
		"cd":       shell.cdCommand,
		"exit":     shell.exitCommand,
		"vfs-save": shell.vfsSaveCommand,
		"vfs-load": shell.vfsLoadCommand,
	}
	return shell
}

// SHELL METHODS
func (s *Shell) lsCommand(args []string) {
	fmt.Printf("Command: ls, arguments: %v\n", args)
}
func (s *Shell) cdCommand(args []string) {
	fmt.Printf("Command: cd, arguments: %v\n", args)
}
func (s *Shell) exitCommand(args []string) {
	os.Exit(0)
}
func (s *Shell) vfsLoadCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("vfs-load: need path to directory")
		return
	}
	err := s.vfs.LoadFromDisk(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
func (s *Shell) vfsSaveCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("vfs-save: need path to save")
		return
	}
	if !s.vfs.IsLoaded {
		fmt.Println("vfs-save: VFS isn`t loaded")
		return
	}
	err := s.vfs.SaveToDisk(args[0])
	if err != nil {
		fmt.Printf("vfs-save: save error: %v\n", err)
	}
	fmt.Printf("VFS saved to %v\n", args[0])
}
func (s *Shell) executeCommand(cmd string, args []string) error {
	if handler, exists := s.commands[cmd]; exists {
		handler(args)
	} else {
		return errors.New("сommand doesn`t exists")
	}
	return nil
}
func (s *Shell) executeScript(scriptPath string) error {
	file, err := os.Open(scriptPath)
	if err != nil {
		return err
	}
	defer file.Close()
	var cmd_err error
	cmd_err_flag := 0
	fmt.Printf("Startup script started work\n")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		// пропускаем пустые строки и комментарии
		if input == "" || strings.HasPrefix(input, "#") {
			continue
		}
		fmt.Printf("%s%s\n", getInvitation(), input)
		cmd, args := parser(input)
		cmd_err = s.executeCommand(cmd, args)
		if cmd_err != nil {
			cmd_err_flag = 1
			fmt.Printf("Error: %v\n", cmd_err)
		}
	}
	if cmd_err_flag == 1 {
		return errors.New("сommand doesn`t exists")
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Кастомное приглашение к вводу
func getInvitation() string {
	// Имя пользователя
	var username string
	currentUser, err := user.Current()
	if err != nil {
		username = os.Getenv("USER")
	} else {
		username = currentUser.Name
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
	var vfsPath string
	var startupScript string
	var help bool

	// vfs - параметр, -vfs аргументы, если нет аргументов, то vfsPath = ".vfs", "Path to VFS" - текст справки при вызове -help
	flag.StringVar(&vfsPath, "vfs", "", "Path to VFS")
	flag.StringVar(&startupScript, "script", "", "Path to startup script")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&help, "h", false, "Show help")

	flag.Parse()

	if help {
		flag.Usage()
	}
	fmt.Print("Commands:\nls {arguments}\ncd {arguments}\nexit\nvfs-save {path}\nvfs-load {path}\n")

	shell := NewShell()
	if vfsPath != "" {
		err := shell.vfs.LoadFromDisk(vfsPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	if startupScript != "" {
		if err := shell.executeScript(startupScript); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Println("Script ended with error")
			os.Exit(1)
		} else {
			fmt.Println("Script successfully ended")
		}
	}

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

		if cmd != "" {
			err := shell.executeCommand(cmd, args)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
