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

    $ cabri cli lsns fsy:/home/guest/simple_files@

Synchronization is performed by providing a source and a target namespaces in their respective DSS:

    $ cabri cli sync -r fsy:/home/guest/simple_files@ olf:/media/guest/usbkey/simple_backup@ --macl :

The `--macl` flags maps a source ACL user with a target one.
For a `fsy` DSS, the empty ACL user corresponds to the current user id.
For the target `olf` DSS, we map the previous one to the empty ACL user.
This enables to back up files with their original system permission as metadata
so that a reverse synchronization will restore the files with the same permissions.

You will then find all files from the `fsy` DSS synchronized in the `olf` DSS:

    $ cabri cli lsns -rs olf:/media/guest/usbkey/simple_backup@
