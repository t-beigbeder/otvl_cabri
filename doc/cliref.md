# Reference documentation for Cabri DSS CLI

## General help

Cabri CLI is built with coral (see [here](dev.md))
which helps to build a CLI with commands and sub-commands along with their specific flags and arguments.
Help for their usage is always available using for instance

    $ cabri help
    $ cabri cli help

## DSS specification

DSS management commands use the specification of a DSS as following:

    <type>:<location>

Where type can be:

- fsy: a portion if a native filesystem, in which case `<location>` is its path. 
  - example: `fsy:/home/guest`
- olf: object-like files on a native filesystem, in which case `<location>`
is the directory where DSS configuration, index, metadata and data files are stored.
  - example: `olf:/media/guest/usbkey/simple_backup`
- obj: a portion of an object store (Swift container or Amazon S3 bucket),
in which case `<location>` is the directory where
DSS configuration and index are stored.
  - example: `obs:/home/guest/cabri_config/cloud_backup`
- smf: object storage mocked as files for development and tests,
in which case `<location>` is the directory where DSS configuration, index,
metadata and data files are stored.
  - example: `/home/guest/cabri_tests/smf_backup`
- xolf, xobj and xsmf stand for encrypted DSS of the corresponding type