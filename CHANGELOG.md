# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## 0.3.7 - 2023-04-08
### Fixed
* Remove potential race condition caused by using `RLock` instead of `Lock`

## 0.3.6 - 2023-02-25
### Changed
* Updated Go build version
### Fixed 
* Multiple security advisories

## 0.3.5 - 2022-01-19
### Added
* Added space for contact information to user agent

## 0.3.4 - 2022-01-19
### Added
* Support for `ETag` and `Last-Modified` headers in feed responses
### Changed
* Added version number to email footer

## 0.3.3 - 2022-08-31
### Fixed
* Feed entries can now be deleted. [#1](https://github.com/codemicro/walrss/issues/1)
* Proper errors are shown when attempting to register with an in-use email address. [#2](https://github.com/codemicro/walrss/issues/2)

## 0.3.2 - 2022-08-13
### Added
* OIDC support

## 0.3.1 - 2022-08-13
This one deploys.

## 0.3.0 - 2022-08-13
### Changed
* Migrated to SQLite in favour of BoltDB. This a completely breaking change.

## 0.2.3 - 2022-06-01
### Added
* Added a user agent to any requests made by Walrss (a very basic regex for this is `walrss(\/(\d|\.){5})? \(https:\/\/github\.com\/codemicro\/walrss\)`)

## 0.2.2 - 2022-05-08
### Fixed
* Feed entries from midnight on a given day are no longer mistakenly ignored.

## 0.2.1 - 2022-04-29
### Fixed
* Digest emails no longer contain three extra days worth of feed items

## 0.2.0 - 2022-04-29
### Added
* Progress display for test emails

## 0.1.1 - 2022-04-16
### Added
* Support for OPML imports and exports

### Fixed
* Secure cookies are no longer sent when debug mode is enabled

## 0.1.0 - 2022-04-14
Initial release
