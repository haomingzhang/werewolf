package client

import (
	"encoding/json"
	"github.com/haomingzhang/werewolf/game"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	longPollingInterval = 120 * time.Second
)

type WerewolfClient struct {
	client *http.Client
	uri    *url.URL
}

func CreateWerewolfClient(serverHost string) (*WerewolfClient, error) {
	uri, err := url.Parse("http://" + serverHost + game.ClientEndpoint)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &WerewolfClient{
		client: &http.Client{
			Timeout: longPollingInterval,
		},
		uri: uri,
	}, nil
}

func (w *WerewolfClient) Start() {
	w.poll()
}

func (w *WerewolfClient) poll() {
	for {
		res, err := w.client.Get(w.uri.String())
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				continue
			}
			log.Fatal(err.Error())
			return
		}
		if res.StatusCode == http.StatusOK {
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			res.Body.Close()
			clientRes := &game.ClientResponse{}
			err = json.Unmarshal(resBytes, clientRes)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			game.SleepAndPlayAudio(clientRes.TurnCode)
		}
	}
}
