# Cabri Data Storage System

Cabri enables fast and secure data synchronization between people, medias and places.

It is mainly available as a command-line tool, but also provides an API (Golang or REST) for simple data storage.

Cabri is currently in beta release. A GUI is under development.

Documentation links are available at the [bottom](#read-the-documentation) of this page.

## Simple presentation

To make it simple, Cabri can be compared to the synchronization command-line tool: `rsync`

- using HTTP instead of SSH for remote access
- enabling unidirectional or bidirectional synchronization 
- providing synchronization between local files and S3 compatible object storage
- but also with a data store on local storage
- providing in both cases data historization, deduplication or encryption,
on a system neutral storage system
- and enabling multi-user data sharing through remote server
- taking care of data confidentiality in the Cloud or any unsafe environment,
relying on public keys for data encryption and related sharing

all of that as a no-dependency single binary of less than 30 MB.

To get an idea, have a look at the [quick start](doc/qstart.md).

## Tooling

- a Golang or a REST API are available for access to data storage services
- a basic configurable scheduler is provided,
enabling automatic data synchronization among users,
or between users and hosted applications for the support of DevOps practices

## Implementation information

- Most actions are performed in parallel,
synchronization is fast if the infrastructure provides enough resources
- Cloud storage is natively "eventually consistent" and Cabri takes care of not trusting successful updates
- Cabri makes use of indexes for fast synchronization or data retrieval, indexes may be rebuilt if broken
- Encrypted data is only decrypted on the user endpoint even when using a remote http server

This [blog article](https://blog.otvl.org/blog/cabri-tech-ovw) provides detailed technical information.

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
