package public_ip

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func Get() (string, error) {
	if response, err := http.Get("https://ipinfo.io/ip"); err == nil {
		if bytes, err := ioutil.ReadAll(response.Body); err == nil {
			ip := string(bytes)
			return ip, nil
		}
	}

	return "", errors.New("not available")
}
