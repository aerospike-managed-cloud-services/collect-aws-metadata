# collect-aws-metadata
Reads AWS instance meta-data and creates a Prometheus .prom text file with upcoming maintenance events.

## Installation

1. Download the [latest release]. You most likely want the file named like: `collect-aws-metadata-vX.Y.Z.tar.gz`

1. Unpack

    ```
    tar xvfz collect-aws-metadata-v*.tar.gz
    ```

1. Copy `./collect-aws-metadata` to somewhere in your PATH

## Integrate with prometheus and systemd

To be effective, this tool must be run by systemd to output to a Prometheus
node_exporter textfiles directory.

#### Prometheus

You will need the 
[Prometheus node_exporter](https://github.com/prometheus/node_exporter) plugin
installed and configured correctly.

In particular, you must be sure that you are using this parameter:
```
--collector.textfile.directory
```

Whatever directory you have this set to will be needed by the systemd service config.

#### systemd

You should create both a service and a timer for this tool.

Set up a systemd *service* to find the binary and pass arguments to it.

<details>
<summary>systemd service file (also find this in <i>doc/sample</i>)</summary>

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
<summary>systemd timer file (also find this in <i>doc/sample</i>)</summary>

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


----

## Maintainer section: releasing

This repo supports builds for x86_64 and ARM64, use the corresponding make target to create the correct binary. 
To cut a release of this software, automated tests must pass. Check under `Actions` for the latest commit.

#### Create an RC branch and test

- We use the Gitflow process. For a release, this means that you should have a v1.2.3-rc branch under your 
  develop branch. Like this:
  ```
    main  
    └── develop  
        └── v1.2.3-rc
  ```

- Update *this file*.
  
  1. Confirm that the docs make sense for the current release.
  1. Check links!
  1. Update the Changelog section at the bottom.

- Perform whatever tests are necessary.

#### Tag and cut the release with Github Actions

- Once you have tested in this branch, create a tag in the v1.2.3-rc branch:
  ```
  git tag -a -m v1.2.3 v1.2.3
  git push --tags
  ```

- Navigate to [collect-aws-metadata Actions](https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/actions) and run the action labeled `collect-aws-metadata release`.

    - You will be asked to choose a branch. Choose your rc branch, e.g. `v1.2.3-rc`

    - If you run this action without creating a tag on v1.2.3-rc first, the action will fail with an error and nothing will happen.

  If you have correctly tagged a commit and chosen the right branch, this will run and create a new release on the [Releases page].

- Edit the release on that page 

#### Merge up

- Finish up by merging your `-rc` branch into 
  1. `main` and then 
  2. `develop`.


## Changelog

<details><summary>(About: Keep-a-Changelog text format)</summary>

The format is based on [Keep a Changelog], and this project adheres to [Semantic
Versioning].
</details>

### [1.1.0] - 2021-12-10

#### Fixed

- Renamed `instance` metric label to `cloud_instance` and other labels
  similarly, to prevent Prometheus from clobbering them.

#### Added

- `event_state` metric label, `event_date` and `days_hence`, to make it easier
  to produce a useful alert message.

### [1.0.0] - 2021-12-01

#### Changed

- This repo is now public.

### [0.9.0] - 2021-12-01

#### Added
- Docs for setting up prometheus and systemd to run this tool

#### Changed
- Prep for a first public release


[Unreleased]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/compare/v0.9.0...HEAD
[1.1.0]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/compare/v0.9.0...v1.0.0
[0.9.0]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/compare/v0.0...v0.9.0
[0.0]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/releases/tag/v0.0


[latest release]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/releases/latest
[Releases page]: https://github.com/aerospike-managed-cloud-services/collect-aws-metadata/releases
[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
