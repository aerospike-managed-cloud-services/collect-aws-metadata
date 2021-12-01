# collect-aws-metadata
Reads AWS instance meta-data and creates a Prometheus .prom text file with upcoming maintenance events.

## Installation

- Download [the latest release](https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/releases/download/latest/collect-aws-metadata-latest.tar.gz)
- Unpack

    ```
    tar xvfz collect-aws-metadata-latest.tar.gz
    ```

- Copy `./collect-aws-metadata` to somewhere in your PATH

## Integrate with prometheus and systemd

### Prometheus

You will need the 
[Prometheus node_exporter](https://github.com/prometheus/node_exporter) plugin
installed and configured correctly.

In particular, you must be sure that you are using this parameter:
```
--collector.textfile.directory
```

Whatever directory you have this set to will be needed by the systemd service config.

### systemd

You should create both a service and a timer for this tool.

Set up a systemd *service* to find the binary and pass arguments to it.

<details>
<summary>systemd service file (also find this in <i>doc/samples</i>)</summary>

```
[Unit]
Description=Collect AWS maintenance events
Wants=collect-aws-metadata.timer
After=collect-aws-metadata.timer

[Service]
ExecStart=/opt/my_deployment/bin/collect-aws-metadata --textfiles-path=/opt/node_exporter/textfile_collector/ --metric-prefix=my_org_

User=prometheus
Group=nodeexporter
Type=oneshot

[Install]
WantedBy=multi-user.target
```

</details>

Set up a system *timer* to run the service on a timed schedule.

<details>
<summary>systemd timer file (also find this in <i>doc/samples</i>)</summary>

```
[Unit]
Description=Collect AWS maintenance events timer
Requires=collect-aws-metadata.service
After=network-online.target

[Timer]
Unit=collect-aws-metadata.service
# every 5 minutes
OnCalendar=*:0/5

Persistent=true
AccuracySec=1s

[Install]
WantedBy=timers.target
```

</details>


## Maintainer section: releasing

To cut a release of this software, automated tests must pass. Check under `Actions` for the latest commit.

In addition:

- We use the Gitflow process. For a release, this means that you should have a v1.2.3-rc branch under your 
  develop branch. Like this:
  ```
    main  
    └── develop  
        └── v1.2.3-rc
  ```

- Once you have tested in this branch, create a tag in the v1.2.3-rc branch:
  ```
  git tag -a -m v1.2.3 v1.2.3
  git push --tags
  ```

- Navigate to [collect-aws-metadata Actions](https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/actions) and run the action labeled `collect-aws-metadata release`.

    - You will be asked to choose a branch. Choose your rc branch, e.g. `v1.2.3-rc`

    - If you run this action without creating a tag on v1.2.3-rc first, the action will fail with an error and nothing will happen.

  If you have correctly tagged a commit and chosen the right branch, this will run and create a new release on the [Releases page](https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/releases).

- TBD: update docs

- Finish up by merging your `-rc` branch into first `main` and then `develop`.


## Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- unreleased features

## [1.0.0] - 2021-11-99
### Added
- added

### Changed
- changed

### Removed
- removed

[Unreleased]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/compare/v0.0...v1.0.0
[0.0]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/releases/tag/v0.0
