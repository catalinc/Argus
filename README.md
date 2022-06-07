# Argus

Motion detection and alerting.

## MacOS Setup
```shell
# Install prerequisites
brew install pkg-config opencv4

# Clone this repo
git clone git@github.com:catalinc/argus.git

# Build
cd argus
make test
make build

# Run
./argus
```