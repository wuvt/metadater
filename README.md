# metadater

metadater is a tool to update metadata on an Icecast stream, on TuneIn, and on
Last.fm from data submitted to Trackman's SSE endpoint at /playlists/live.

This functionality was previously handled by a Celery task in Trackman, but
was split out into a separate project to reduce complexity and increase
flexibility.

Unlike Trackman itself, this code is available under the GPLv3 to allow
modifications to be made for specific production uses without a requirement to
offer the source code.
