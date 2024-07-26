go build -o build/stacks main.go
src=$(pwd)
echo "export PATH=\"$src/build:\$PATH\"" >> ~/.zshrc
echo "Stacks has been installed succesfully. Restart your terminal and type 'stacks' to run the program."
