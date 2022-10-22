# Cabri Data Storage System

Cabri is an hybrid (local and cloud) Data Storage System,
also providing a fast and secure synchronization service.

Cabri is currently in beta release, encryption implementation is not yet finished.

## General presentation

### Open Source

Cabri is provided as
[FOSS](https://en.wikipedia.org/wiki/Free_and_open-source_software)
under the [BSD 3-Clause License](LICENSE)

### Data Storage System

A Cabri Data Storage System (DSS) can be compared to a filesystem with respect to its ability
to store data along with its metadata: hierarchical naming, modification time
and access control information.

Indeed, a portion of a native filesystem can be handled as a DSS
and then synchronized with other kinds of DSS.

Those other kinds of DSS may in turn provide additional features:

- historization, enabling to keep snapshots of a full data hierarchy
- deduplication, improving storage utilization and enabling a more efficient synchronization
when data is just renamed
- encryption, enabling the storage of confidential data on unsecure media such as USB drives,
public cloud object stores and other unprotected storage systems

Those other kinds of storage are independent of the operating system.

### Object storage in the cloud

Cabri supports storage in S3 enabled object stores,
such as Openstack Swift containers or of course Amazon S3 buckets.
By design, object stores only provide eventual consistency, and Cabri takes care of data consistency in such
conditions.

### Remote access

Cabri provides an HTTP API that primarily enables remote access to specific physical devices,
but also enables multi-user consistent access to shared data.

When S3 API for cloud object storage is not considered reliable, secure, or fast enough from a local network,
remote access via a proxy in the cloud can also be used, so that S3 calls are performed fully in the cloud.

### Data synchronization

Apart from basic storage services, Cabri provides a data synchronization service between DSS.
Synchronization may be unidirectional or bidirectional.

The service thus enables data backup, data distribution or both,
along with historization, deduplication or encryption as already mentioned. 

### Encryption

Both the data and the metadata (such as hierarchical naming) may be encrypted using
[public key encryption](https://en.wikipedia.org/wiki/Public-key_cryptography),
which means that the encrypted data content may be shared efficiently with several users
each owning a personal secret key.
Only the corresponding public keys need to be shared, secret keys are kept confidential as intended.

The users are the owners of their secret keys.
Secret keys are never used outside the scope of the component requesting or updating data.
This also means that when using encryption, confidential data is never exposed to third parties
neither in transit nor at rest.

Internally, Cabri makes use of [age](https://age-encryption.org/) technology
whose specification can be found [here](https://github.com/C2SP/C2SP/blob/main/age.md).
Many thanks to its author Filippo Valsorda!

Data encryption is incompatible with deduplication because the same content
is never encrypted the same twice.

### Not a filesystem

Cabri does not provide a POSIX like filesystem API nor does it provide the highest I/O rates.
However, its components try to make the best use of the underlying system I/O capabilities
both concerning the storage and the network if it is involved in the transfer,
in particular by parallelizing processing as much as possible.

### Indexing

Cabri makes use of indexes to enable fast access to metadata:

- in the cloud
- from the history
- encrypted, in which case it is kept local

Indexes can be rebuilt if broken or lost by performing a full scan of the repository.

## Details

### Concepts and terminology

- DSS: a Data Storage System is a repository
  providing technology neutral storage services for data along with its metadata:
  hierarchical naming, modification time and access control information. A DSS can be
  - fsy: a portion if a native filesystem (no history, no deduplication, no encryption)
  - obj: a portion of an object store (Swift container or Amazon S3 bucket)
    providing history, deduplication or encryption, and supporting eventual consistency limitations
  - olf: object-like files on a native filesystem to provide history, deduplication or encryption
  - smf: object storage mocked as files for development and tests
- namespace: namespaces provide a hierarchical naming scheme, as POSIX directories do,
  and as POSIX directories, they are composed of names separated by the character "/"
  - by convention, names ending with "/" are considered to be namespaces,
    their content is the list of their children names, which in turn are either namespaces or data 
  - name not ending with "/" are data entries
  - by convention, the root of the DSS is the empty string, it is an exception to the rules above
- access control lists: they describe the access rights to the requested entries
  - POSIX user, group and "other" for "fsy" DSSs along with their access rights
  - `age` public keys along with their access rights for encrypted DSSs
    (but the ability to decrypt the encrypted data and metadata is obviously a prerequisite)
  - access control may be bypassed like for tools such as `tar` or `rsync`;
    if access control has to be enforced, DSS files must be kept out of direct access
    and a reverse proxy ensuring authentication has to be used along with remote access through the HTTP API  
  - simple labels can be used and mapped to users, groups and public keys if wanted, this can be useful
    for synchronizing data between DSSs using different conventions

### API

Cabri comes with a Go API, an HTTP API, and a Go HTTP client (same API as Go native)
providing technology neutral storage services similar to POSIX file access API:

- stat entry: size, mtime, access control lists
- namespaces creation and update (namespaces are like POSIX directories or Windows folders)
- data content creation, update and retrieval
- data or namespace removal is simply achieved by updating the parent namespace list of children

Additional services concern:

- synchronization between DSS
- the management of the repository itself
- the management of the repository history
- the management of the indexes
- the activation of an HTTP API server

The Go API is documented here:
[pkg.go.dev](https://pkg.go.dev/github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss). 

### UI

Cabri currently provides a [CLI](https://en.wikipedia.org/wiki/Command-line_interface)
for all the services of the API, and also for performing synchronization.

Cabri will soon come with a Web User Interface for the same services.
This interface will be designed to be able to run locally,
thus never exposing secrets nor confidential data to third parties.
