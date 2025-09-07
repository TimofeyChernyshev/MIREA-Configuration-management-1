@echo off
echo Creating test VFS structure...

REM Создаем корневую директорию
mkdir test_vfs
cd test_vfs

REM Создаем файл motd
echo Welcome to Test VFS This is the Message Of The Day > motd

REM Создаем поддиректории и файлы
mkdir dir1
mkdir dir2
mkdir dir3

echo File in dir1 > dir1\file1.txt
echo Another file in dir1 > dir1\file2.txt

mkdir dir2\subdir1
mkdir dir2\subdir2

echo File in subdir1 > dir2\subdir1\subfile1.txt
echo File in subdir2 > dir2\subdir2\subfile2.txt

mkdir dir3\level1
mkdir dir3\level1\level2
mkdir dir3\level1\level2\level3

echo Deeply nested file > dir3\level1\level2\level3\deepfile.txt

REM Создаем несколько файлов в корне
echo Root file 1 > root1.txt
echo Root file 2 > root2.txt

echo Test VFS structure created successfully!
cd ..