package main

type Config struct {
	LiveUrl            string   `json:"LIVE_URL"`
	IcecastAdmin       string   `json:"ICECAST_ADMIN"`
	IcecastMounts      []string `json:"ICECAST_MOUNTS"`
	LastfmApiKey       string   `json:"LASTFM_APIKEY"`
	LastfmSecret       string   `json:"LASTFM_SECRET"`
	LastfmUsername     string   `json:"LASTFM_USERNAME"`
	LastfmPassword     string   `json:"LASTFM_PASSWORD"`
	TuneInPartnerId    string   `json:"TUNEIN_PARTNERID"`
	TuneInPartnerKey   string   `json:"TUNEIN_PARTNERKEY"`
	TuneInStationId    string   `json:"TUNEIN_STATIONID"`
	HealthCheckWebhook string   `json:"HEALTHCHECK_WEBHOOK"`
}

type Track struct {
	Id     uint64 `json:"id"`
	Added  string `json:"added"`
	Album  string `json:"album"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type TrackLog struct {
	Dj        string `json:"dj"`
	DjId      uint64 `json:"dj_id"`
	DjVisible bool   `json:"dj_visible"`
	DjSet     uint64 `json:"djset"`

	TrackLogId uint64 `json:"tracklog_id"`
	Listeners  uint64 `json:"listeners"`
	Played     string `json:"played"`
	RotationId uint64 `json:"rotation_id"`
	TrackId    uint64 `json:"track_id"`
	Track      Track  `json:"track"`

	New     bool `json:"new"`
	Request bool `json:"request"`
	Vinyl   bool `json:"vinyl"`
}

type MessageData struct {
	Event    string   `json:"event"`
	TrackLog TrackLog `json:"tracklog,omitempty"`
}
