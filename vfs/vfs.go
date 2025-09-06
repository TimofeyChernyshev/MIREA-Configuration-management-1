package vfs

import (
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
	Root    *VFSNode `json:"root"` // Корневой узел
	Mounted bool     `json:"-"`    // Загружена ли VFS в память
}
