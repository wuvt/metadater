#!/usr/bin/python3

import dateutil.parser
from dateutil import tz
import defaults
import hashlib
import json
import logging
import os.path
import pylast
import requests
import sseclient
import time


class Config(dict):
    def load_from_object(self, obj):
        for key in dir(obj):
            if key.isupper():
                self[key] = getattr(obj, key)

    def load_from_json(self, path):
        with open(path) as f:
            self.update(json.load(f))


def update_stream(track):
    for mount in config['ICECAST_MOUNTS']:
        r = requests.get(config['ICECAST_ADMIN'] + 'metadata', params={
            'mount': mount,
            'mode': 'updinfo',
            'album': track['album'],
            'artist': track['artist'],
            'title': track['title'],
        }, timeout=config['REQUEST_TIMEOUT'])
        if r.status_code != 200:
            logger.warning("Update stream mount {0} failed: {1}".format(
                mount, r.status_code))


def update_tunein(track):
    if len(config['TUNEIN_PARTNERID']) > 0:
        try:
            r = requests.get('https://air.radiotime.com/Playing.ashx', params={
                'partnerId': config['TUNEIN_PARTNERID'],
                'partnerKey': config['TUNEIN_PARTNERKEY'],
                'id': config['TUNEIN_STATIONID'],
                'title': track['title'],
                'artist': track['artist'],
            }, timeout=config['REQUEST_TIMEOUT'])
            r.raise_for_status()
        except Exception as exc:
            logger.warning("Update TuneIn failed: {0}".format(exc))


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


if __name__ == '__main__':
    config = Config()
    config.load_from_object(defaults)

    config_path = os.environ.get('APP_CONFIG_PATH', 'config.py')
    if os.path.exists(config_path):
        if config_path.endswith('.py'):
            import config as _configobj
            config.load_from_object(_configobj)
        else:
            config.load_from_json(config_path)

    logger = logging.getLogger(__name__)

    messages = sseclient.SSEClient(config['LIVE_URL'], chunk_size=1024)
    for msg in messages:
        try:
            data = json.loads(msg.data)
            if data['event'] == 'track_change':
                track = data['tracklog']['track']
                played = dateutil.parser.parse(data['tracklog']['played'])
                logger.info("{track} played at {played}".format(track=track,
                                                                played=played))

                update_stream(track)
                update_tunein(track)
                update_lastfm(
                    track,
                    int(time.mktime(played.replace(tzinfo=tz.tzutc()).
                                    astimezone(tz.tzlocal()).timetuple())))
            elif data['event'] == 'track_edit':
                track = data['tracklog']['track']
                update_stream(track)

            if len(config['HEALTHCHECK_WEBHOOK']) > 0:
                try:
                    r = requests.get(config['HEALTHCHECK_WEBHOOK'],
                                     timeout=config['REQUEST_TIMEOUT'])
                except requests.exceptions.RequestException as e:
                    logger.warning(
                        "Healthcheck webhook failed: {}".format(e))
        except Exception as e:
            logger.warning("Failed to process message: {}".format(e))
