.PHONY: all dep build test

all: dep build test

dep:
	go mod download

build:
	go build -o ./p2p-chat ./

test:
	go test -race ./...

image:
	docker build --network=host -f .docker/Dockerfile -t docker.io/goforbroke1006/p2p-chat:latest ./

#PUBLIC_IP=$(shell curl -s ipinfo.io/ip)
#LOCAL_IP=$(shell bash local-ip.sh)
FREE_PORT=$(shell bash freeport.sh)
#ANY_IP=0.0.0.0
#FREE_PORT=16998

#setup:
#	#sudo ufw allow 18888/tcp
#	#iptables-save > IPtablesbackup.txt
#	sudo iptables -A INPUT -p tcp --dport 18888 -j ACCEPT
#	touch peers.txt

start: build
	./p2p-chat cli -u foo-bar -P ${FREE_PORT}