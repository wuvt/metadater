package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/r3labs/sse"
	"github.com/shkh/lastfm-go/lastfm"
)

// Times in the Trackman API may be either RFC1123 or RFC3339
func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC1123, value)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return parsed, err
		}
	}
	return parsed, nil
}

func parseConfig(path string) (c *Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal(body, &c)
	if err != nil {
		return
	}

	return
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "/etc/metadater/config.json", "path to config file")
	flag.Parse()

	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	client := sse.NewClient(config.LiveUrl)

	client.Subscribe("messages", func(msg *sse.Event) {
		if len(msg.Data) > 0 {
			var data MessageData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Printf("Error unmarshalling JSON data: %s\n", err)
				return
			}
			if data.Event == "track_change" {
				played, err := parseTime(data.TrackLog.Played)
				if err != nil {
					log.Printf("Error parsing played time: %s\n", err)
				}
				log.Printf("%s - %s played at %s", data.TrackLog.Track.Artist, data.TrackLog.Track.Title, played)

				// Icecast
				go func() {
					for _, mount := range config.IcecastMounts {
						err := updateIcecast(config.IcecastAdmin, mount, data.TrackLog.Track)
						if err != nil {
							log.Printf("Failed to update Icecast mount %s: %s", mount, err)
						}
					}
				}()

				// RDS
				if len(config.RdsHost) > 0 {
					go func() {
						err := updateRds(config.RdsHost, data.TrackLog)
						if err != nil {
							log.Printf("Failed to update RDS: %s", err)
						}
					}()
				}

				// TuneIn
				if len(config.TuneInPartnerId) > 0 && len(config.TuneInPartnerKey) > 0 && len(config.TuneInStationId) > 0 {
					go func() {
						err := sendGetWebhook(
							"https://air.radiotime.com/Playing.ashx",
							map[string]string{
								"partnerId": config.TuneInPartnerId,
								"parnerKey": config.TuneInPartnerKey,
								"id":        config.TuneInStationId,
								"title":     data.TrackLog.Track.Title,
								"artist":    data.TrackLog.Track.Artist,
							})
						if err != nil {
							log.Printf("Failed to update TuneIn: %s", err)
						}
					}()
				}

				// Last.fm
				if len(config.LastfmApiKey) > 0 && len(config.LastfmSecret) > 0 && len(config.LastfmUsername) > 0 && len(config.LastfmPassword) > 0 {
					api := lastfm.New(config.LastfmApiKey, config.LastfmSecret)
					err := api.Login(config.LastfmUsername, config.LastfmPassword)
					if err == nil {
						p := lastfm.P{
							"album":  data.TrackLog.Track.Album,
							"artist": data.TrackLog.Track.Artist,
							"track":  data.TrackLog.Track.Title,
						}
						_, err = api.Track.UpdateNowPlaying(p)
						if err != nil {
							log.Printf("Failed to update Last.fm now playing: %s", err)
						}

						p["timestamp"] = played.Unix()
						_, err = api.Track.Scrobble(p)
						if err != nil {
							log.Printf("Failed to scrobble track to Last.fm: %s", err)
						}
					} else {
						log.Printf("Failed to update Last.fm: %s", err)
					}
				}
			} else if data.Event == "track_edit" {
				log.Print("Track edit")

				// Icecast
				go func() {
					for _, mount := range config.IcecastMounts {
						err := updateIcecast(config.IcecastAdmin, mount, data.TrackLog.Track)
						if err != nil {
							log.Printf("Failed to update Icecast mount %s: %s", mount, err)
						}
					}
				}()

				// RDS
				if len(config.RdsHost) > 0 {
					go func() {
						err := updateRds(config.RdsHost, data.TrackLog)
						if err != nil {
							log.Printf("Failed to update RDS: %s", err)
						}
					}()
				}
			}

			if len(config.HealthCheckWebhook) > 0 {
				go func() {
					err := sendGetWebhook(config.HealthCheckWebhook, map[string]string{})
					if err != nil {
						log.Printf("Failed to send healthcheck webhook: %s", err)
					}
				}()
			}
		}
	})
}
