.PHONY: all dep build test

all: dep build test

dep:
	go mod download

build:
	go build -o ./p2p-chat ./cmd

test:
	go test -race ./...



PUBLIC_IP=$(shell curl -s ipinfo.io/ip)
LOCAL_IP=$(shell bash local-ip.sh)
FREE_PORT=$(shell bash freeport.sh)
#FREE_PORT=16998

#setup:
#	sudo ufw allow from any to any port ${FREE_PORT} proto tcp
#	sudo ufw allow from any to any port ${FREE_PORT} proto udp
	#sudo iptables -A INPUT -p tcp --dport ${FREE_PORT} -j ACCEPT
	#sudo iptables -A INPUT -p udp --dport ${FREE_PORT} -j ACCEPT

start: build
	./p2p-chat foo-bar ${PUBLIC_IP} ${FREE_PORT}