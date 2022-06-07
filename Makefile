USER=pi
SYSTEMD_SERVICES=/lib/systemd/system
INSTALL_DIR=/home/$(USER)/argus

.PHONY: all test build install_raspi clean

all: test build

test:
	@echo "Running tests"
	go test

build:
	@echo "Building executable"
	go build -o argus cmd/argus/main.go

install_raspi: all
	@echo "Installing service"
	rm -fr $(INSTALL_DIR)
	mkdir -p $(INSTALL_DIR)
	mv argus $(INSTALL_DIR)
	cp config.json $(INSTALL_DIR)
	cat argus.service | sed "s/USER/$(USER)/g" > $(INSTALL_DIR)/argus.service
	sudo mv $(INSTALL_DIR)/argus.service $(SYSTEMD_SERVICES)
	@echo "Service installed to $(INSTALL_DIR)"
	@echo "To test service run   : sudo systemctl start|stop argus.service"
	@echo "To enable service run : sudo systemctl enable argus.service"

clean:
	rm argus
