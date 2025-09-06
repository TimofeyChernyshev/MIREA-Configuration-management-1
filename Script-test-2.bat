echo "Test 1: Calling script without arguments"
go run main.go -script

echo "Test 2: Script with wrong command"
go run main.go -script Script-startup-wrong.txt

echo "Test 3: Script with correct commands"
go run main.go -script Script-startup-right.txt