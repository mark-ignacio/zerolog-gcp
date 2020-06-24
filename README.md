# zerolog-gcp

a (hopefully) straightforward LevelWriter for using [zerolog](github.com/rs/zerolog) with Google Cloud Logging of Google Operations, all of which used to be named Stackdriver.

Some notable features:

* The first log written to Cloud Logging is a slow, blocking write to confirm connectivity + permissions, but all subsequent writes are non-blocking.
* Handles converting `zerolog.WarnLevel` to `logging.Warning`.
* Zerolog's trace level maps to Cloud Logging's Default level.
* Cloud Logging's Alert and Emergency levels are not used.
