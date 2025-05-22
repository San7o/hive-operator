.PHONY: all clean setup network-reset generate
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

generate:
	awk -v content="$$(awk '{print "      " $$0}' scripts/k8s-lab.sh)" '{gsub("%setup-content%", content)}1' cloud-init/master/user-data-template > cloud-init/master/user-data
	awk -v content="$$(awk '{print "      " $$0}' scripts/k8s-lab.sh)" '{gsub("%setup-content%", content)}1' cloud-init/worker1/user-data-template > cloud-init/worker1/user-data
