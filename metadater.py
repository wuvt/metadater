#!/usr/bin/python3

import dateutil.parser
import defaults
import hashlib
import json
import logging
import os.path
import pylast
import requests
import sseclient


class Config(dict):
    def load_from_object(self, obj):
        for key in dir(obj):
            if key.isupper():
                self[key] = getattr(obj, key)


def update_stream(track):
    for mount in config['ICECAST_MOUNTS']:
        r = requests.get(config['ICECAST_ADMIN'] + 'metadata', params={
            'mount': mount,
            'mode': 'updinfo',
            'album': track['album'],
            'artist': track['artist'],
            'title': track['title'],
        })
        if r.status_code != 200:
            logger.warning("Update stream mount {0} failed: {1}".format(
                mount, r.status_code))


def update_tunein(track):
    if len(config['TUNEIN_PARTNERID']) > 0:
        r = requests.get('http://air.radiotime.com/Playing.ashx', params={
            'partnerId': config['TUNEIN_PARTNERID'],
            'partnerKey': config['TUNEIN_PARTNERKEY'],
            'id': config['TUNEIN_STATIONID'],
            'title': track['title'],
            'artist': track['artist'],
        })
        if r.status_code != 200:
            logger.warning("Update TuneIn failed: {0}".format(r.status_code))


def update_lastfm(track, timestamp):
    if len(config['LASTFM_APIKEY']) > 0:
        h = hashlib.md5()
        h.update(config['LASTFM_PASSWORD'].encode('utf-8'))
        password_hash = h.hexdigest()

        try:
            network = pylast.LastFMNetwork(
                api_key=config['LASTFM_APIKEY'],
                api_secret=config['LASTFM_SECRET'],
                username=config['LASTFM_USERNAME'],
                password_hash=password_hash)
            network.scrobble(
                artist=track['artist'],
                title=track['title'],
                timestamp=timestamp,
                album=track['album'])
        except Exception as exc:
            logger.warning("Last.fm scrobble failed: {}".format(exc))

config = Config()
config.load_from_object(defaults)

if os.path.exists('config.py'):
    import config as _configobj
    config.load_from_object(_configobj)

logger = logging.getLogger(__name__)

messages = sseclient.SSEClient(config['LIVE_URL'])
for msg in messages:
    data = json.loads(msg.data)
    if data['event'] == 'track_change':
        track = data['tracklog']['track']
        played = dateutil.parser.parse(data['tracklog']['played'])
        logger.info("{track} played at {played}".format(track=track,
                                                        played=played))

        update_stream(track)
        update_tunein(track)
        update_lastfm(track, played)
