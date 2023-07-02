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
- obs: an object store (Swift container or Amazon S3 bucket),
in which case `<location>` is the directory where
the DSS configuration and index are stored.
  - example: `obs:/home/guest/cabri_config/cloud_backup`
- smf: object storage mocked as files for development and tests,
in which case `<location>` is the directory where DSS configuration, index,
metadata and data files are stored.
  - example: `smf:/home/guest/cabri_tests/smf_backup`
- webapi+http, webapi+https: remote access to object or object-like DSS,
in which case `<location>` takes the form `://<host>:<port>/<url-path>`
  - example: `webapi+http://localhost:3000/demo`
- xolf, xobs and xsmf stand for encrypted DSS of the corresponding type
- xwebapi+http, xwebapi+https must be used as well if the remote DSS is encrypted

## Object storage creation parameters

You will need to provide the following information for access to the object storage:

    --obsrg <object storage region>
    --obsep <object storage endpoint>
    --obsct <object storage swift container or aws bucket>
    --obsak <object storage access key>
    --obssk <object storage secret key>

This information will be kept encrypted in the local DSS configuration
for further use.
NB: it is encrypted with the client `__internal__` identity,
so the client configuration should be encrypted itself with a master password
if your local environment is not secure.
See [client configuration](cliconf.md) for more information.

## Namespace specification

A namespace specification simply has the form:

    <DSS specification>@<path_to_the_namespace>

knowing that the trailing slash has to be omitted for commands
dealing exclusively with namespaces.

Examples:
- fsy:/home/guest/simple_directory@
- olf:/media/guest/usbkey/simple_backup@
- obs:/home/guest/cabri_config/simple_backup@sub_namespace
