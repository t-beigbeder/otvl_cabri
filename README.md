# Cabri Data Storage System

## Share data with confidence using Cabri

Cabri is a free and open source tool designed specifically to store data
and synchronize it on various media, among various places, with the people you want.

It is both fast and secure, providing confidentiality in unsecured environments.

It is mainly available as a command-line tool, but also provides an API (Golang and REST).
A GUI is under development.

Cabri is currently in beta release.

## Main features

- Cabri manages data storage on external devices such as USB drives
and using Cloud Storage services compatible with Amazon S3
- Cabri synchronizes local data files with those external storage systems 
and external storage systems between each other
possibly in both directions at the same time 
- Storage is incremental
- Storage may be encrypted, relying on public keys, meaning no secrets need to be shared

## Tooling

- Access to external storage systems can be provided through a remote http server,
enabling among others multi-user sharing and synchronization of common data
from different locations
- a REST API is available for access to data storage services
- a basic configurable scheduler is provided enabling automatic data synchronization among users,
but also from developers to hosted applications concerning data feeding

## Read the documentation

- [Introduction](doc/intro.md)
- [Getting started](doc/gscli.md)

Other documentation is referenced from these pages, including:

- [Tuning synchronization parameters](doc/synctune.md)
- [Simple HTTPS implementation](doc/https.md)
- [ACL (Access Control List)](doc/acl.md)
- [Data encyption](doc/encrypt.md)
- [DSS management](doc/mng.md)
- [Reference documentation for the CLI](doc/cliref.md)
- [Client configuration](doc/cliconf.md)
- [REST API](doc/restapi.md)
- [Building the application for various platforms](doc/build.md)
