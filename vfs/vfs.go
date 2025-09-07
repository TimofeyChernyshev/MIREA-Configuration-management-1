package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Узел виртуальной файловой системы
type VFSNode struct {
	Name     string     `json:"name"`               // Имя файла/папки
	IsDir    bool       `json:"isDir"`              // true - папка, false - файл
	Content  string     `json:"content,omitempty"`  // Содержимое файла (для файлов)
	Children []*VFSNode `json:"children,omitempty"` // Дочерние узлы (для папок)
	ModTime  time.Time  `json:"modTime"`            // Время последнего изменения
}

// Виртуальная файловая система
type VFS struct {
	Root     *VFSNode `json:"root"` // Корневой узел
	IsLoaded bool     `json:"-"`    // Загружена ли VFS в память
}

func (v *VFS) LoadFromDisk(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	v.Root = &VFSNode{
		Name:    filepath.Base(absPath),
		IsDir:   true,
		ModTime: time.Now(),
	}
	// Рекурсивный обход всех файлов и папок
	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filePath == absPath {
			return nil // Пропускаем корневую директорию
		}
		relPath, err := filepath.Rel(absPath, filePath)
		if err != nil {
			return err
		}
		node := &VFSNode{
			Name:    info.Name(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime(),
		}
		if !info.IsDir() {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			node.Content = string(content)
		}
		v.addNode(relPath, node)
		return nil
	})
	v.IsLoaded = true
	if err == nil {
		fmt.Printf("VFS loaded from: %v\n", path)
		v.PrintMOTD()
	}
	return err
}

func (v *VFS) addNode(path string, node *VFSNode) {
	// Разбиваем путь на части
	parts := strings.Split(path, string(filepath.Separator))
	current := v.Root
	// Проход по частям пути
	for i, part := range parts {
		// Если последняя часть пути, то это файл. Добавляем узел в текущую папку
		if i == len(parts)-1 {
			current.Children = append(current.Children, node)
			return
		}
		// Ищем существующую папку
		var found *VFSNode
		for _, child := range current.Children {
			if child.Name == part && child.IsDir {
				found = child
				break
			}
		}
		// Если папка не найдена, то создаем новую
		if found == nil {
			found = &VFSNode{
				Name:     part,
				IsDir:    true,
				ModTime:  time.Now(),
				Children: []*VFSNode{},
			}
			current.Children = append(current.Children, found)
		}
		current = found
	}
}

func (v *VFS) SaveToDisk(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	// Рекурсивно создаем все папки в пути
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return err
	}
	return v.saveNode(v.Root, absPath)
}

func (v *VFS) saveNode(node *VFSNode, basePath string) error {
	nodePath := filepath.Join(basePath, node.Name)
	if node.IsDir {
		if err := os.MkdirAll(nodePath, 0755); err != nil {
			return err
		}
		for _, child := range node.Children {
			if err := v.saveNode(child, nodePath); err != nil {
				return err
			}
		}
	} else {
		if err := os.WriteFile(nodePath, []byte(node.Content), 0644); err != nil {
			return err
		}
		os.Chtimes(nodePath, time.Now(), node.ModTime)
	}
	return nil
}

func (v *VFS) FindNode(path string) (*VFSNode, error) {
	if path == "" || path == "/" {
		return v.Root, nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := v.Root

	for _, part := range parts {
		if part == "." {
			continue
		}
		if part == ".." {
			current = v.Root
			continue
		}
		found := false
		for _, child := range current.Children {
			if child.Name == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("file or directory doesn`t found %s", part)
		}
	}
	return current, nil
}

func (v *VFS) PrintMOTD() {
	motdNode, err := v.FindNode("motd")
	if err == nil && !motdNode.IsDir {
		fmt.Printf("%s\n", motdNode.Content)
	}
}
