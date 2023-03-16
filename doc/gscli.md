# Getting started with Cabri DSS using CLI

The reference documentation for the CLI is provided [here](cliref.md).

When in doubt with a command syntax, use the `--help` or `-h` flag, for instance:

    cabri cli -h
    cabri cli sync --help

## Creating a local DSS

Use the command `cabri cli dss make`

For instance this will create an `olf` DSS with a small size on the provided directory:

    $ mkdir /media/guest/usbkey/simple_backup
    $ cabri cli dss make -s s olf:/media/guest/usbkey/simple_backup

`olf` DSS don't need to be indexed because local access to metadata is reasonably fast.
Anyway if you wish to create such a DSS with an index, then use the following command,
`bdb` standing for the indexing technology buntdb:

    $ cabri cli dss make -s s --ximpl bdb olf:/media/guest/usbkey/simple_backup

## Creating the root namespace

A DSS namespace is provided after the DSS name using the `@` separator, the root namespace is empty:

    $ cabri cli dss mkns olf:/media/guest/usbkey/simple_backup@

You can list it with the lsns command:

    $ cabri cli lsns olf:/media/guest/usbkey/simple_backup@

## Synchronizing a local directory with the DSS

Local directory access is possible using a `fsy` DSS referring to this directory,
the following namespace is the root in the DSS, so the directory itself:

    $ cabri cli lsns fsy:/home/guest/simple_directory@

Synchronization is performed by providing a source and a target namespaces in their respective DSS:

    $ cabri cli sync -r fsy:/home/guest/simple_directory@ olf:/media/guest/usbkey/simple_backup@ --macl :

The `--macl` flags maps a source ACL user with a target one.
For a `fsy` DSS, the empty ACL user corresponds to the current user id.
For the target `olf` DSS, we map the previous one to the empty ACL user.
This enables to back up files with their original system permission as metadata
so that a reverse synchronization will restore the files with the same permissions.

You will then find all files from the `fsy` DSS synchronized in the `olf` DSS:

    $ cabri cli lsns -rs olf:/media/guest/usbkey/simple_backup@

## Handling secrets securely

Storing content in the cloud requires access to object storage secret keys.
Encrypting content requires using the user's secret key.
The CLI must have convenient but secure access to both kind of secrets,
under the user's control.
They are stored in a configuration file named `clientConfig`,
stored by default in the `.cabri` directory of the user's home directory. 

This file should be protected with a master password, as following

    $ cabri cli config --encrypt
    please enter the master password:
    please enter the master password again: 

WARNING: if you loose this password, you will definitely loose stored secrets
that you didn't backup by other means, for instance your encryption's secret keys
if you generated them with the CLI.

Once the configuration file is encrypted, you must provide the master password
to any CLI command that needs accessing it, for instance:

    $ cabri cli config --dump
    Error: password required to perform this action

This can be done interactively:

    $ cabri cli config --dump --password
    please enter the master password:

or through a password file:

    $ mkdir /home/guest/secrets && chmod go-rwx /home/guest/secrets
    $ echo "mysecret" > /home/guest/secrets/cabri
    $ cabri cli config --dump --pfile /home/guest/secrets/cabri

Dumping the configuration is done as

    $ cabri cli config --dump --password
    please enter the master password:
    {
    "clientId": "<a unique id for this CLI client's configuration>",
    "Identities": [
    {
    "alias": "",
    "pKey": "<user's default public key>",
    "secret": "<user's default secret key>"
    }
    ],
    "Internal": {
    "alias": "__internal__",
    "pKey": "<default public key for this CLI client's configuration>",
    "secret": "<an internal secret key for this CLI client's configuration>"
    }
    }

Explanations about the configuration are provided on the page
[client configuration](cliconf.md).

You can manage those identities, including the `__internal__` one, with the same command,
using proposed flags:

    $ cabri cli config --help
    ...
    Flags:
    -d, --decrypt   decrypts the configuration file with master password
    --dump      dumps the configuration file
    -e, --encrypt   encrypts the configuration file with master password
    --gen       generate a new identity for one or several aliases
    --get       display an identity for one or several aliases
    --put       <alias> <pkey> [<secret>] import or update an identity for an alias, secret may be unknown
    --remove    remove an identity alias

## Synchronizing a local directory with cloud object storage

First create the DSS with the command `cabri cli dss make`,
referring to a DSS type `obs`,
and providing required information (see the [CLI reference](cliref.md)):

- `<DSS local storage location>`
- object storage access information

for instance:
    
    $ cabri cli --password \
        --obsrg GRA --obsep https://s3.gra.cloud.ovh.net \
        --obsct simple_backup_container \
        --obsak access_key --obssk secret_key \
        dss make obs:/home/guest/cabri_config/simple_backup

Once the DSS is created, object storage information is kept in its configuration,
no need to mention it again for further use.
Synchronization is performed by providing a source and a target namespaces as seen above:

    $ cabri cli sync -r fsy:/home/guest/simple_directory@ obs:/home/guest/cabri_config/simple_backup@

You will then find all files from the `fsy` DSS synchronized in the `obs` DSS:

    $ cabri cli lsns -rs obs:/home/guest/cabri_config/simple_backup@
