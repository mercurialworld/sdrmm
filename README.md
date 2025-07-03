# Smart DumbRequestManager Manager

It's not as smart as I thought it would be. Made for my use case (Mix It Up allows me to call executables with arguments). 

I HATE WRITING IN GO

# TODO
- return the original message, or errors if there are any

# Setup

probably going to need MIU open so it can receive webhooks (for queue open/closed status)

# Config

```toml
[drm]
url = "http://localhost"
port = 13337

[bsr]
request-limit = 0               # set this to 0 to allow unlimited requests
newer-than = 2018-05-08         # ISO 8601 formatted date
map-age = 0                     # how many days old the map should be
min-length = 0                  # min length of the song, in seconds
max-length = 0                  # max length of the song, in seconds

[nps]
min = 0
max = 0                         # set this to 0 to just not have limits

[njs]
min = 0
max = 0                         # set this to 0 to just not have limits
```

# Usage

TODO