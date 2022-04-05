package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"p2p-chat/domain"
)

func ReadPeerListFromFile(filename string) (peers []domain.Peer, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewReader(bytes.NewReader(data))
	for {
		line, _, err := scanner.ReadLine()
		if err != nil {
			break
		}
		parts := strings.Split(string(line), ":")

		port, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("can't parse port in %s addr", string(line)))
		}

		peers = append(peers, domain.Peer{
			Host: parts[0],
			Port: uint16(port),
		})
	}

	return peers, nil
}
