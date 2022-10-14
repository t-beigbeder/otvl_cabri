# Cabri Data Storage System

Cabri is an hybrid (local and cloud) Data Storage System
coming with a fast and secure synchronization service.

## General presentation

### Open Source

Cabri is provided as
[FOSS](https://en.wikipedia.org/wiki/Free_and_open-source_software)
under the [BSD 3-Clause License](LICENSE)

### Data Storage System

A Cabri Data Storage System (DSS) can be compared to a filesystem with respect to its ability
to store data along with its metadata: hierarchical naming, modification time
and access control information.

Indeed, data from a native filesystem can be easily synchronized with other kinds of DSS.
And those other kinds of DSS can provide other features:

- historization, enabling snapshots of a full data hierarchy
- deduplication, improving storage utilization and enabling faster synchronization
when data is just renamed
- encryption, enabling the storage of confidential data on unsecure media such as USB drives,
public cloud object stores and other unprotected storage systems

### Remote access

Cabri provides an HTTP API that enables:

- multi user consistent access to shared data
- improving network bandwidth utilization in some circumstances

### Data synchronization

Apart from basic storage services, Cabri provides a data synchronization service between DSS.
Synchronization may be unidirectional or bidirectional.

The service thus enables data backup, data distribution or both,
along with historization, deduplication or encryption as already mentioned. 

### Object storage in the cloud

Cabri supports storage in S3 enabled object stores, such as Openstack Swift or of course Amazon S3.
Object stores only provide eventual consistency, and Cabri takes care of data consistency in such
conditions.

### Encryption

You are the owners of your secret keys.

Secret keys are never used outside the scope of the component requesting or updating data.

That also means that when using encryption, confidential data is never exposed to third parties
neither in transit nor at rest.

Internally, Cabri makes use of [age](https://age-encryption.org/) technology
whose specification can be found [here](https://github.com/C2SP/C2SP/blob/main/age.md).
Big thanks to its author [Filippo Valsorda](https://filippo.io/) !

From a non-technical point of view, `age` technology uses
[public key encryption](https://en.wikipedia.org/wiki/Public-key_cryptography),
which means that the encrypted data content may be shared efficiently with several users
each owning a personal secret key.

Both the data and the metadata (such as hierarchical naming) are encrypted.

Data encryption is incompatible with deduplication because the same content
is never encrypted the same twice.

### Not a filesystem

Cabri does not provide a POSIX like filesystem API nor does it provide high I/O rates.
However, its components try to make the best use of the underlying system I/O capabilities
both concerning the storage and the network if it is involved in the transfer.

### Indexing

Cabri makes use of indexes to enable fast access to metadata:

- in the cloud
- from the history
- encrypted, in which case it is kept local

Indexes can be rebuilt if broken or lost by performing a full scan of the repository.

## Details

### Concepts and terminology

- DSS: Data Storage System is a repository
  providing technology neutral storage services
  for data along with its metadata: hierarchical naming, modification time
  and access control information. A DSS can be
  - fsy: a portion if a native filesystem (no history, no deduplication, no encryption)
  - obj: a portion of an object store providing history, deduplication or encryption,
    and supporting eventual consistency
  - olf: object-like files on a native filesystem to provide history, deduplication or encryption
  - smf: object storage mocked as files for development and tests

### API

Cabri comes with a Go API, an HTTP API, and a Go HTTP client (same API as Go native)
providing technology neutral storage services.

Provided services:

- stat entry: type, size, mtime, user, group
- directory listing
- get content
- write content with mtime
- delete entry

Additional services concern:

- the management of the repositoy itself
- the management of the repository history
- the management of the indexes

### UI

Cabri currently provides a [CLI](https://en.wikipedia.org/wiki/Command-line_interface)
for all the services of the API, but also for performing synchronization.

Cabri will soon come with a Web User Interface for the same services.
