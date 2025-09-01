# Smart DumbRequestManager Manager

A wrapper for TheBlackParrot's [DumbRequestManager](https://github.com/TheBlackParrot/DumbRequestManager) that adds filters and request limits. Written in Rust because the first iteration was written in Go and I hated every second of it.

This is a CLI program, as it was meant to be called by Mix It Up's "Executable" action. You also should make sure you give it time to run in your bot software of choice.

**THIS REQUIRES DRM v0.6.0 OR HIGHER TO WORK.**

> [!IMPORTANT]
> This is beta software, so if you aren't me and you're using this, there may be unexpected behaviour. Please let me know if there's anything by opening an issue in this repository.

# Commands

- `sdrmm new`: Starts a new session.
- `sdrmm request <BSR ID> <USERNAME> (--service/-s <PLATFORM>) (--modadd)`: Attempts to add a request of `<BSR ID>` to the queue. Goes through filters. If the `--modadd` flag is present, forces the map in.
- `sdrmm ban <BSR ID>`: Bans `<BSR ID>` from ever being requested.
- `sdrmm unban <BSR ID>`: Allows `<BSR ID>` to be requested again.
- `sdrmm wip <link/WIPBot site code> <USERNAME>`: Adds a WIP to the queue, if it's an allowed link or there's a code.
- `sdrmm queue <open/close/toggle>`: Changes state of queue (open/close).
- `sdrmm getqueue (--user/-u <USERNAME>)`: Returns a formatted message with how many songs in the queue and how long the queue is. Optionally, shows where a user's requests are in the queue.
- `sdrmm clear`: Clears the queue. **If you're using this program and have `session_max` set, it is highly recommended to use this whenever you clear the queue.**
- `sdrmm top <USERNAME>`: Moves the most recent request of a user to the top of queue. **The user must have requested something first.**
- `sdrmm oops <USERNAME>`: Undoes the most recent request of a user.
- `sdrmm refund <USERNAME>`: If streamer has request limits and a map is banned/skipped or the queue is cleared, adds 1 to the number of requests a user has.

# Config

The config is written in YAML. sdrmm expects the config to be called `config.yaml`.

```yaml
drm:
    url: "http://localhost"    # The URL where the DRM server is hosted.
    port: 13337                # The port of the DRM server.
    new_session_length: 60     # Time, in minutes, to consider a session "new".
queue:
    session_max: 0    # How many maps can be requested per session. Set to 0 to ignore.
    queue_max: 0      # How many maps of a user can be in the queue. Set to 0 to ignore.
    repeat: false     # Whether to add the same map more than once to the queue.
    replay: false     # Whether to play maps that have already been played this session.
bsr:
    censors:
        deny_censored: false   # Whether to deny a map if any of the "Censor" fields are true.
        deny_urls: false       # Whether to deny a map if it contains any domains. Depends on `deny_censored`.
    allow_ai: false            # Whether to allow Beat Sage/other such AI maps. Honestly, leave this false.
    min_rating: 0.0            # Minimum rating on BeatSaver. Set to 0 to ignore.
    date:
        earliest: 1970-01-01   # YYYY-MM-DD format
        min_age: 0             # How old a map is, in days. Useful if you don't want users to request new maps.
    length:
        min: 0                 # Minimum length, in seconds. Set to 0 to ignore.
        max: 0                 # Maximum length, in seconds. Set to 0 to ignore.
    nps:
        min: 0.0               # Minimum NPS. Set to 0 to ignore.
        max: 0.0               # Maximum NPS. Set to 0 to ignore.
    njs: 
        min: 0.0               # Minimum NJS. Set to 0 to ignore.
        max: 0.0               # Maximum NJS. Set to 0 to ignore.
map_vote:
    allow_liked: true          # Whether to auto-allow maps that have been liked.
    deny_disliked: true        # Whether to deny maps that have been disliked.
playlists:                     # A list of local playlists a map can be in, in order to auto-allow the map. Can be empty.
    - ""    
```

# TODO
- [ ] more configuration
  - [X] make `new_session_length` which mirrors `AssumeNewSessionAfterMinutes` 
  - [X] `!oops`
  - [ ] `hide_map_metadata` for `!link`
- [ ] configurable string format (look into localization?)