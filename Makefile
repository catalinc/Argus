USER=pi
SYSTEMD_SERVICES=/lib/systemd/system
INSTALL_DIR=/home/$(USER)/argus

.PHONY: all test build install_raspi clean

all: test build

test:
	@echo "\e[96mRunning tests\e[0m"
	go test

build:
	@echo "\e[96mBuilding executable\e[0m"
	go build -o argus cmd/argus/main.go

install_raspi: all
	@echo "\e[96mInstalling service\e[0m"
	rm -fr $(INSTALL_DIR)
	mkdir -p $(INSTALL_DIR)
	mv argus $(INSTALL_DIR)
	cp config.json $(INSTALL_DIR)
	cat argus.service | sed "s/USER/$(USER)/g" > $(INSTALL_DIR)/argus.service
	sudo mv $(INSTALL_DIR)/argus.service $(SYSTEMD_SERVICES)
	@echo "\e[96mService installed to $(INSTALL_DIR)\e[0m"
	@echo "\e[97mTo test service run   : sudo systemctl start|stop argus.service\e[0m"
	@echo "\e[97mTo enable service run : sudo systemctl enable argus.service\e[0m"

clean:
	rm argus
