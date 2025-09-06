echo "Test 1: No arguments"
go run main.go

echo "Test 2: Help"
go run main.go -help

echo "Test 3: Custom VFS"
go run main.go -vfs D:\custom_vfs

echo "Test 4: Name of startup script is wrong"
go run main.go -script Wrong-script.txt