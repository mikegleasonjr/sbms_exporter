[![CircleCI](https://circleci.com/gh/mikegleasonjr/sbms_exporter.svg?style=svg)](https://circleci.com/gh/mikegleasonjr/sbms_exporter)

# SBMS Exporter

```
$ ./sbms_exporter -h
usage: sbms_exporter --serial-port=SERIAL-PORT [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --telemetry-path="/metrics"
                                 Path under which to expose metrics.
      --listen-address=":9101"   Address to listen on for web interface and telemetry.
      --serial-port=SERIAL-PORT  The serial port to read metrics from.
      --log.level="info"         Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"
                                 Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version                  Show application version.
```
