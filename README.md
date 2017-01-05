# mill

mill is a Go package implementing structured logging and is designed around the
Go 1.7 context package.  Any number of loggers can be attached to a context, and
extra logging tags can be added to the context to be included in every log using
that context.

## Features

* Structured logging
* Asynchronous (but data-race free) logging
* Per-context log tags and tag key/value pairs
* Custom log entry codecs (text, JSON, ...)
* Timestamps that can be lexicographically compared
* Enabling and disabling of per-context and global runtime debug logging
* Compile time removal of all debugging using the `release` build tag.
* Semver release versions
* Permissive license

## Non-features

* Log levels

  I believe "debug" is the only extra log verbosity I will ever need or want,
  and I am not alone in this sentiment (see Dave Cheney's take on the subject
  [here](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)).  If
  something is worth logging, just log it.

  Selective debugging is not quite the same as traditional log "levels" as there
  is no heirarchy of levels to pick between.  Debug logs are neither
  "higher" nor "lower" than non-debug logs.

  If you must have log levels, something similar can likely be accomplished
  using tags.  If not, this package is probably not for you.

* File loggers, log rotation, ...

  While these could be implemented by writing a custom underlying writer, it is
  not a goal of this project to support "files as logs".  Instead, output is
  intended to be written to stdout/stderr and handled by some other log
  aggregator (syslog, splunk, etc.) if log persistence and inspection is
  necessary.  This is where structured logging really begins to shine, as it
  allows for much more detailed analysis of your logs.

  See http://adam.herokuapp.com/past/2011/4/1/logs_are_streams_not_files/ for a
  longer explanation.

* Zero allocations! Fastest logger ever!

  I like performance too but let's not go overboard.  If writing log entries to
  stdout without reaching for unsafe at every line is already too slow for you,
  you might want to reconsider your language choice.

## Requirements

Go 1.7 or later (although a fork of the project could be used with older Go
releases by switching to the golang.org/x/net/context package)

## NIH?

Absolutely

## License

mill is licensed under the liberal ISC License.
