go build -o build/stacks main.go
path=$(pwd)
echo "alias stacks='$path/build/stacks'" >> ~/.zshrc
echo "Stacks has been added to your aliases. Restart your terminal and type 'stacks' to run the program."
