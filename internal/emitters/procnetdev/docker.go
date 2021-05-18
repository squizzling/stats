package procnetdev

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

type container struct {
	Id string
}

type containerDetail struct {
	Id    string
	Name  string
	State *containerState
}

type containerState struct {
	Status  string
	Running bool
	Pid     int
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")

		},
	},
}

func dockerEnabled() bool {
	_, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// fail hard, probably a permissions thing
		panic(err)
	}
	return true
}

func queryDocker(url string, output interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(output)
	if err != nil {
		return err
	}
	return err
}

func getDockerContainerDetail(id string) *containerDetail {
	var detail containerDetail
	if err := queryDocker(fmt.Sprintf("http://x/containers/%s/json", id), &detail); err != nil {
		panic(err)
	}
	detail.Name = strings.TrimLeft(detail.Name, "/")
	return &detail
}

func getDockerContainerIDs() []string {
	if !dockerEnabled() {
		return nil
	}
	var containers []*container
	if err := queryDocker("http://x/containers/json", &containers); err != nil {
		panic(err)
	}

	var ids []string
	for _, container := range containers {
		ids = append(ids, container.Id)
	}
	return ids
}
