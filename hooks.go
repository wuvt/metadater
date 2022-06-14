package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func updateIcecast(icecastUrl string, mount string, track Track) error {
	v := url.Values{}
	v.Add("mount", mount)
	v.Add("mode", "updinfo")
	v.Add("album", track.Album)
	v.Add("artist", track.Artist)
	v.Add("title", track.Title)

	u, err := url.Parse(icecastUrl)
	if err != nil {
		return err
	}
	u.Path += "metadata"
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Icecast returned a status code (%d) other than 200", resp.StatusCode)
	}

	return nil
}

func sendGetWebhook(webhookUrl string, params map[string]string) error {
	u, err := url.Parse(webhookUrl)
	if err != nil {
		return err
	}

	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	u.RawQuery = values.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Webhook returned a status code (%d) other than 200", resp.StatusCode)
	}

	return nil
}
