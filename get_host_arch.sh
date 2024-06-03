uname_arch=$(uname -m)

architecture="amd64"

if [ "$uname_arch" == "x86_64" ] || [ "$uname_arch" == "amd64" ]; then
    architecture="amd64"
elif [ "$uname_arch" == "aarch64" ] || [ "$uname_arch" == "arm64" ]; then
    architecture="arm64"
fi

echo $architecture
