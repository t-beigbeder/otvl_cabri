# Cabri Data Storage System

## Share data with confidence using Cabri

Cabri is a free and open source tool designed specifically to store data
and synchronize it on various media, between different places, with the people you want.

It is both fast and secure, providing confidentiality in untrusted environments.

It is mainly available as a command-line tool, but also provides an API (Golang and REST).
A GUI is under development.

Cabri is currently in beta release.

## Main features

- Cabri manages data storage on local and external devices such as USB drives,
or using Cloud Storage services compatible with Amazon S3
- Access to storage systems can be provided through a remote http server,
  enabling among others multi-user sharing and synchronization of common data
  from different locations
- Cabri provides a data synchronization service between those storage systems,
  synchronization may be unidirectional or bidirectional 
- Storage is incremental, no data is lost until you decide to remove some parts of the history
- Storage may be encrypted, relying on public keys, meaning no secrets need to be shared

## Tooling

- a REST API is available for access to data storage services
- a basic configurable scheduler is provided enabling automatic data synchronization among users,
but also from developers to hosted applications concerning data feeding

## Implementation

- Most actions are performed in parallel, synchronization is as fast as the infrastructure permits
- Cloud storage is natively "eventually consistent" and Cabri takes care of not trusting successful updates
- Cabri makes use of indexes for fast synchronization or data retrieval, indexes may be rebuilt if broken
- Encrypted data is only decrypted on the user endpoint even when using a remote http server

## Read the documentation

The tool is documented through the links below:

- [Introduction](doc/intro.md)
- [Getting started](doc/gscli.md)
- [Quick reference](doc/qref.md)

Other documentation is referenced from these pages, including:

- [Tuning synchronization parameters](doc/synctune.md)
- [Information for configuring and using a secured HTTP server](doc/https.md)
- [ACL (Access Control List)](doc/acl.md)
- [Data encyption](doc/encrypt.md)
- [DSS management](doc/mng.md)
- [Reference documentation for Cabri DSS CLI](doc/cliref.md)
- [Client configuration](doc/cliconf.md)
- [REST API](doc/restapi.md)
- [Building the application for various platforms](doc/build.md)

External blog:

- [Using Cabri to share data with confidence](https://blog.otvl.org/blog/cabri-share-conf)
