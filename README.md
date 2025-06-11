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
url = "http://localhost"        # IP of the machine the BS client is on, if you have a different one
port = 13337                    # Port of the DRM API

[bsr]
request-limit = 0               # number of requests allowed for each person per session 
newer-than = 2018-05-08         # ISO 8601 formatted date (YYYY-MM-DD)
```

# Usage

```
Usage: sdrmm.exe <command> [<args>]

Options:
  --help, -h             display this help and exit

Commands:
  request                Put a map in the queue
  mtt                    Put a user's request to the top of the queue
  wip                    Put a WIP in the queue
  getqueue               Get the queue
  queue                  Toggle/get queue open/closed status
  clear                  Clear the queue
  ban                    Ban a map from being requested
  unban                  Unban a map from being requested
```

## Request

```
Usage: sdrmm.exe request [--user USER] ID

Positional arguments:
  ID                     The BeatSaver ID of the map to be requested

Options:
  --user USER, -u USER   The username of the requester
```

## MTT (Mine To Top)

```
Usage: sdrmm.exe mtt USER

Positional arguments:
  USER                   The username whose map you want to put on top of the queue
```

## WIP

```
Usage: sdrmm.exe wip --user USER ID

Positional arguments:
  ID                     The WIP code of the map to be requested

Options:
  --user USER, -u USER   The username of the requester
```

## Get Queue

```
Usage: sdrmm.exe getqueue [USER]

Positional arguments:
  USER                   The username who invoked the command
```

## Queue

```
Usage: sdrmm.exe queue [--status]

Options:
  --status, -s           Get queue status
```

## Clear

```
Usage: sdrmm.exe clear [--save]

Options:
  --save                 Whether to save the queue as it was before deleting all of it, in case you want to refund points or something [default: false]
```

## Ban

```
Usage: sdrmm.exe ban ID

Positional arguments:
  ID                     The BeatSaver ID of the map to be banned
```

## Unban

```
Usage: sdrmm.exe unban ID

Positional arguments:
  ID                     The BeatSaver ID of the map to be unbanned
```

