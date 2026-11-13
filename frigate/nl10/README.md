# Frigate on Internal-Multi rootless Docker

This stack uses the service account's rootless Docker daemon (host UID 1000).
The current deployment uses CPU OpenVINO detection, with no hardware decode or
GPU/Coral device mappings. Privileged mode is disabled. Reassess device access
before adding hardware acceleration.

## Host prerequisites

The host mounts the NAS at /mnt/Frigate. Keep Docker storage and the Frigate
SQLite/config directory on local ext4; media remains on the NAS.

Install bindfs, create /mnt/Frigate-rootless, and install rootless-media.service
as /etc/systemd/system/frigate-rootless-media.service. Enable/start that service
with systemctl. This is a host administrator step; Arcane Git sync does not
install system services.

The mount maps NAS UID/GID 0 to host UID/GID 1000 for Frigate's rootless view.
Files created through that view remain root-owned on the NAS. Existing media
ownership is not recursively changed. The original NAS mount and its .nas-ready
sentinel must be present. The Compose entrypoint retains its NAS readiness guard
and existing 90-second startup delay.

FRIGATE_MEDIA_PATH can override the mapped source path when required. The default
is /mnt/Frigate-rootless. The Compose project path/name and published ports are
unchanged so Arcane can continue managing the project.

The host /etc/Frigate/config directory must be writable by UID/GID 1000. Its
credentials, SQLite state, and runtime secrets must not be committed to Git.
During migration, quiesce the original Frigate before transferring ownership
of its existing local config directory in place. Keep runtime keys at their
existing paths. Never run both daemons' Frigate containers against the same writable state.

## Validation and recovery

Verify all nine camera capture/detection processes, newly written recording
segments in SQLite and on NAS, historical playback, previews, and HTTP/RTSP
services after the container becomes healthy. Enabling Arcane Git sync is a
separate final step after the Git source and live rootless definition agree.

Keep the original rootful container and sanitized ownership recovery metadata until
validation is complete. Stop rootless Frigate before any rollback. The pinned
application version is unchanged, but reconcile any new database writes rather
than restoring an older database blindly. Keep the mapped mount available while
rootless Frigate is running.
