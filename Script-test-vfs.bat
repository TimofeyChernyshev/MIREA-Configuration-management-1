@echo off
echo Testing VFS functionality...

REM Создаем тестовый скрипт
echo vfs-load test_vfs > test_script.txt
echo ls >> test_script.txt
echo vfs-save saved_vfs >> test_script.txt

REM Запускаем shell с скриптом
go build -o MIREA-Configuration-management-3.exe main.go
MIREA-Configuration-management-3.exe -script test_script.txt

echo.
echo Original VFS:
dir test_vfs

echo.
echo Saved VFS:
dir saved_vfs

pause