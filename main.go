package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TimofeyChernyshev/MIREA-Configuration-management-1/vfs"
)

// Структура для хранения команд - карта, где ключи - имена команд, а значения - функции
type Shell struct {
	commands    map[string]func([]string)
	vfs         vfs.VFS
	currentPath string
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
	shell.currentPath = "/"
	shell.commands = map[string]func([]string){
		"ls":       shell.lsCommand,
		"cd":       shell.cdCommand,
		"exit":     shell.exitCommand,
		"vfs-save": shell.vfsSaveCommand,
		"uniq":     shell.uniqCommand,
		"tail":     shell.tailCommand,
		"mv":       shell.mvCommand,
		"chown":    shell.chownCommand,
	}
	return shell
}

// SHELL METHODS
func (s *Shell) lsCommand(args []string) {
	// Выводит список файлов в директории
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		path = s.currentPath
	}
	node, err := s.vfs.FindNode(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !node.IsDir {
		fmt.Printf("Error: %s is not a directory\n", path)
		return
	}
	for _, child := range node.Children {
		fmt.Printf("%s\n", child.Name)
	}
}
func (s *Shell) cdCommand(args []string) {
	// Позволяет установить текущую директорию
	if len(args) == 0 {
		s.currentPath = "/"
		return
	}
	path := args[0]
	var targetPath string
	if path == "/" {
		targetPath = "/"
	} else if path == "." {
		return // остаемся в текущей директории
	} else if path == ".." {
		// поднимаемся на уровень выше
		if s.currentPath == "/" {
			return // уже в корневой
		}
		part := strings.Split(strings.Trim(s.currentPath, "/"), "/")
		if len(part) <= 1 {
			targetPath = "/"
		} else {
			targetPath = "/" + strings.Join(part[:len(part)-1], "/")
		}
	} else if strings.HasPrefix(path, "/") {
		targetPath = path // абсолютный путь
	} else {
		// путь через просто пробел
		if s.currentPath == "/" {
			targetPath = "/" + path
		} else {
			targetPath = s.currentPath + "/" + path
		}
	}
	node, err := s.vfs.FindNode(targetPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !node.IsDir {
		fmt.Printf("Error: %v is not a directory\n", targetPath)
		return
	}
	s.currentPath = targetPath
}
func (s *Shell) exitCommand(args []string) {
	os.Exit(0)
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
func (s *Shell) uniqCommand(args []string) {
	// Вывод содержимое файла без повторяющихся строк
	if len(args) == 0 {
		fmt.Println("Error: missing arguments")
		return
	}
	filePath := args[0]
	if !strings.HasPrefix(filePath, "/") {
		filePath = s.currentPath + "/" + filePath
	}
	node, err := s.vfs.FindNode(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if node.IsDir {
		fmt.Printf("Error: %v is directory", filePath)
		return
	}
	lines := strings.Split(node.Content, "\n")
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	for _, line := range result {
		fmt.Println(line)
	}
}
func (s *Shell) tailCommand(args []string) {
	// Выводит последние N строк файла (по умолчанию 10)
	if len(args) == 0 {
		fmt.Println("Error: missing arguments")
		return
	}
	lines := 10
	files := []string{}
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "-n" && i+1 < len(args) {
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n <= 0 {
				fmt.Println("Error: invalid number of lines")
				return
			}
			lines = n
			i += 2
		} else {
			files = append(files, arg)
			i++
		}
	}
	if len(files) == 0 {
		fmt.Println("Error: missing argument")
		return
	}
	for _, fileArg := range files {
		filePath := fileArg
		if !strings.HasPrefix(filePath, "/") {
			filePath = s.currentPath + "/" + filePath
		}
		node, err := s.vfs.FindNode(filePath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		if node.IsDir {
			fmt.Printf("Error: %s is directory\n", filePath)
			continue
		}
		// Вывод заголовка для нескольких файлов
		if len(files) > 1 {
			fmt.Printf("Title: %s\n", fileArg)
		}
		contentLines := strings.Split(node.Content, "\n")
		if len(contentLines) == 0 {
			continue
		}
		// Выводим последние N строк
		start := len(contentLines) - lines
		if start < 0 {
			start = 0
		}
		for i := start; i < len(contentLines); i++ {
			fmt.Println(contentLines[i])
		}
		if len(files) > 1 && fileArg != files[len(files)-1] {
			fmt.Println() // Пустая строка между файлами
		}
	}
}
func (s *Shell) mvCommand(args []string) {
	// Перемещает/переименовывает файлы и директории
	if len(args) < 2 {
		fmt.Println("Error: missing arguments")
		return
	}

	sources := args[:len(args)-1]
	destination := args[len(args)-1]

	// Проверяем, является ли назначение директорией
	destNode, err := s.vfs.FindNode(destination)
	isDestDir := err == nil && destNode.IsDir

	// Если перемещаем несколько файлов, назначение должно быть директорией
	if len(sources) > 1 && !isDestDir {
		fmt.Printf("Error: %s is not a directory\n", destination)
		return
	}
	for _, source := range sources {
		sourcePath := source
		if !strings.HasPrefix(sourcePath, "/") {
			sourcePath = s.currentPath + "/" + sourcePath
		}
		sourceNode, err := s.vfs.FindNode(sourcePath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		var destPath string
		if isDestDir {
			// Если назначение - директория, добавляем имя исходного файла/папки
			destPath = destination + "/" + sourceNode.Name
		} else {
			destPath = destination
		}
		// Делаем путь абсолютным, если он относительный
		if !strings.HasPrefix(destPath, "/") {
			destPath = s.currentPath + "/" + destPath
		}
		// Проверяем, не пытаемся ли переместить в самого себя
		if sourcePath == destPath {
			fmt.Println("Error: trying to move to itself")
			continue
		}
		// Проверяем, существует ли уже целевой путь
		existingNode, err := s.vfs.FindNode(destPath)
		if err == nil {
			// Если существует и это директория, и исходный объект тоже директория, то перемещаем внутрь с тем же именем
			if existingNode.IsDir && sourceNode.IsDir {
				destPath = destPath + "/" + sourceNode.Name
			} else {
				fmt.Printf("Error: cannot move %s to %s; File exists\n", source, destination)
				continue
			}
		}
		// Проверяем, не пытаемся ли переместить родительскую папку в дочернюю
		if strings.HasPrefix(destPath, sourcePath+"/") {
			fmt.Printf("Error: cannot move %s into its subdirectory %s\n", source, destination)
			continue
		}
		err = s.vfs.MoveNode(sourcePath, destPath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			fmt.Printf("Moved %s to %s\n", source, destination)
		}
	}
}
func (s *Shell) chownCommand(args []string) {
	// меняет владельца файла
	if len(args) < 2 {
		fmt.Println("Error: missing argument")
		return
	}

	owner := args[0]
	files := args[1:]

	for _, file := range files {
		filePath := file
		if !strings.HasPrefix(filePath, "/") {
			filePath = s.currentPath + filePath
		}

		node, err := s.vfs.FindNode(filePath)
		if err != nil {
			fmt.Printf("chown: %v\n", err)
			continue
		}

		// Изменяем владельца
		node.Owner = owner
		node.ModTime = time.Now()
		fmt.Printf("Changed owner of '%s' to '%s'\n", file, owner)
	}
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
	fmt.Printf("Startup script started work\n")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		// пропускаем пустые строки и комментарии
		if input == "" || strings.HasPrefix(input, "#") {
			continue
		}
		fmt.Printf("%s%s\n", s.getInvitation(), input)
		cmd, args := parser(input)
		cmd_err = s.executeCommand(cmd, args)
		if cmd_err != nil {
			fmt.Printf("Error: %v\n", cmd_err)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Кастомное приглашение к вводу
func (s *Shell) getInvitation() string {
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

	return fmt.Sprintf("%s@%s:~%s$ ", username, hostname, s.currentPath)
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
	flag.StringVar(&vfsPath, "vfs", ".", "Path to VFS")
	flag.StringVar(&startupScript, "script", "", "Path to startup script")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&help, "h", false, "Show help")

	flag.Parse()

	if help {
		flag.Usage()
	}
	fmt.Print("Commands:\nls {arguments}\ncd {arguments}\nexit\nvfs-save {path}\nuniq {path}\ntail {flag} {path} (supports only -n {Number of last N lines of file})\n")

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
		fmt.Print(shell.getInvitation())
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
