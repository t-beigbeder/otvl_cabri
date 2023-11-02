# Cabri Data Storage System

Cabri enables fast and secure data synchronization between people, medias and places.

It is mainly available as a command-line tool, but also provides an API (Golang or REST) for simple data storage.

Cabri is currently in beta release. A GUI is under development.

Documentation links are available at the [bottom](#read-the-documentation) of this page.

## Technical simple presentation

To make it simple, Cabri can be compared to the synchronization command-line tool: `rsync`

- using HTTP instead of SSH for remote access
- enabling mono or bidirectional synchronization 
- providing synchronization between local files and S3 compatible object storage
- but also with a data store on local storage
- providing in both cases data historization, deduplication or encryption, on a system neutral storage system
- and enabling multi-user data sharing through remote server
- taking care of data confidentiality in the Cloud or any unsafe environment 

as a no-dependency single binary of less than 30 MB.

It is fast if the infrastructure is available, using parallel processing.

To get an idea, have a look at the [quick start](doc/qstart.md).

## Functional simple presentation

To get an overview, have a look at the article
[using Cabri to share data with confidence](https://blog.otvl.org/blog/cabri-share-conf).

- Cabri manages data storage on local and external devices such as USB drives,
or using Cloud Storage services compatible with Amazon S3
- Access to storage systems can be provided through a remote http server,
  enabling among others multi-user sharing and synchronization of common data
  from different locations
- Cabri provides a data synchronization service between those storage systems,
  synchronization may be unidirectional or bidirectional 
- Storage is incremental, no data is lost unless it is wanted to remove some parts of the history
- Storage may be encrypted, relying on public keys, meaning no secrets need to be shared

## Tooling

- a Golang or a REST API are available for access to data storage services
- a basic configurable scheduler is provided enabling automatic data synchronization among users,
but also from developers to hosted applications concerning data feeding
(can be useful as a Kubernetes pod sidecar to pull initial data)

## Implementation

- Most actions are performed in parallel, synchronization is as fast as the infrastructure permits
- Cloud storage is natively "eventually consistent" and Cabri takes care of not trusting successful updates
- Cabri makes use of indexes for fast synchronization or data retrieval, indexes may be rebuilt if broken
- Encrypted data is only decrypted on the user endpoint even when using a remote http server

## Read the documentation

The tool is documented through the links below:

- [Introduction](doc/intro.md)
- [Getting started](doc/gscli.md)
- [Quick start](doc/qstart.md)

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
- [Cabri technical overview](https://blog.otvl.org/blog/cabri-tech-ovw)
