.PHONY: all clean setup network-reset
all:
	chmod +x scripts/*
	sudo ./scripts/view.sh

clean:
	chmod +x scripts/clean.sh
	sudo ./scripts/clean.sh || :

setup:
	chmod +x scripts/setup-network.sh
	sudo ./scripts/setup-network.sh

network-reset:
	sudo ./scripts/teardown-network.sh
