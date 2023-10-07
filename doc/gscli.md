# Getting started with Cabri DSS using CLI

## Getting the application

The [repository](https://github.com/t-beigbeder/otvl_cabri) for this tool is hosted on GitHub,
where its binaries can also be [downloaded](https://github.com/t-beigbeder/otvl_cabri/releases)
for some target platforms.

For other platforms or any other reason, you can build the application from source code,
as explained on this [page](build.md).

## Syntax

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
By the way this mapping is applied by default so there is no need to mention it in that case.

You will then find all files from the `fsy` DSS synchronized in the `olf` DSS:

    $ cabri cli lsns -rs olf:/media/guest/usbkey/simple_backup@

See also [Tuning synchronization parameters](synctune.md).

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
        --obsrg gra --obsep https://s3.gra.cloud.ovh.net \
        --obsct simple_backup_container \
        --obsak access_key --obssk secret_key \
        dss make obs:/home/guest/cabri_config/simple_backup

In the case of aws, the container would be the bucket name, and the endpoint would be built similarly,
for instance: `https://s3.eu-west-3.amazonaws.com/`

Please refer to your cloud provider documentation for configuring and securing the access
to your object storage.

Once the DSS is created, object storage information is kept in its configuration,
no need to mention it again for further use, except if it changes, for instance the access or secret keys.
To display or change this configuration, use the `dss config` command, for instance:

    $ cabri cli --password \
        --obsak new_access_key --obssk new_secret_key \
        dss config obs:/home/guest/cabri_config/simple_backup

Synchronization is performed by providing a source and a target namespaces as seen above:

    $ cabri cli sync -r fsy:/home/guest/simple_directory@ obs:/home/guest/cabri_config/simple_backup@

You will then find all files from the `fsy` DSS synchronized in the `obs` DSS:

    $ cabri cli lsns -rs obs:/home/guest/cabri_config/simple_backup@

## Multi-user synchronization with an HTTP server

An HTTP server can make a common DSS available to be synchronized
with the respective DSS of several users.
The flow of data can be from one user to the others through the common server,
but the synchronization may also be performed in both directions between each DSS,
using the `--bidir` flag.

The following describes the setup and use of a simple HTTP server for the demonstration,
however, in most cases, a secure transport should be used with HTTPS.
A simple [HTTPS implementation](https.md) is provided,
but a dedicated proxy in front of the Cabri server, like [traefik](https://traefik.io/traefik/),
apache or nginx, may be required or more effective.

Concerning the use of a HTTPS proxy, providing the authentication credentials is also
explained on the [page mentioned above](https.md).

The server setup is very simple:

- create a DSS, local or cloud object storage, on the server host
- launch the cabri server with a mapping between a URL path and the DSS
- use DSS commands referring to the server's DSS using a special `webapi+http` prefix for the DSS type,
and the URL path chosen above

Local DSS need to be indexed for the server to be able to communicate efficiently
with its clients, you have to use the `--ximpl bdb` flag at creation time, as mentioned above.

For instance:

    $ mkdir /home/guest/olf_server
    $ cabri cli dss make olf:/home/guest/olf_server -s s --ximpl bdb
    $ cabri webapi olf+http://localhost:3000/home/guest/olf_server@demo &
    $ cabri cli dss mkns webapi+http://localhost:3000/demo@

A single server may serve several DSS at the same time, for instance, extending the previous example:

    $ mkdir /home/guest/olf_server2
    $ cabri cli dss make olf:/home/guest/olf_server2 -s s --ximpl bdb
    $ cabri webapi olf+http://localhost:3000/home/guest/olf_server@demo \
        olf+http://localhost:3000/home/guest/olf_server2@demo2 &
    $ cabri cli dss mkns webapi+http://localhost:3000/demo2@
    $ cabri cli lsns webapi+http://localhost:3000/demo@

Now you can synchronize a full directory with the server as seen above, for instance:

    $ cabri cli sync -r fsy:/home/guest/simple_directory@ webapi+http://localhost:3000/demo@

And another user will retrieve it:

    $ cabri cli sync -r webapi+http://localhost:3000/demo@ fsy:/home/reader/retrieved_directory@

You will generally want to control access to data when using DSS in multi-user mode.
ACL (Access Control List) can be used for such a purpose and their basic use is explained on a dedicated [page](acl.md).

## Remote HTTP server for native filesystem DSS

It may sometimes be useful to access a remote native filesystem ("fsy" DSS)
using cabri.
The webapi is different for "fsy" DSS and for object or object-like DSS.
The server will take care of it because it has the information,
for instance as in:

    $ cabri webapi fsy+http://localhost:3000/home/guest/simple_directory@simple_demo &

but on the client-side, the usual `webapi+` type cannot be used
because it is intended to access a remote object or object-like DSS;
use the `wfsapi+` type prefix instead, as in:

    $ cabri cli lsns wfsapi+http://localhost:3000/simple_demo@

## Encrypting your data for secure storage on unsafe media

If your data is confidential, and you have to store it on unsafe media
such as unencrypted laptop drive, USB drive, public cloud storage,
you have the option to encrypt a full Cabri DSS.
It may be worth noting that encryption
can be subject to legal concerns in various countries, take legal advice in doubt.

In that case, all data and corresponding metadata such as entity names are encrypted.
The encryption uses public key, meaning that the data may be decrypted by any of the owners
of the corresponding public keys you have enabled during encryption (including yours, this may help).

Moreover, as Cabri is Open Source and because unencrypted content never
quits the workplace where you launch Cabri commands,
you can be confident that your data is never exposed to unexpected access.
This rule remains true even if you make use of the previously described HTTP server,
which is valuable because a [MITM attack](https://en.wikipedia.org/wiki/Man-in-the-middle_attack)
still wouldn't expose anything.

To create an encrypted DSS, you make use of CLI commands as usual,
but prefix the DSS type with an 'x",
for instance, a local encrypted DSS using `xolf`:

    $ mkdir /media/guest/usbkey/encrypted_backup
    $ cabri cli dss make -s s xolf:/media/guest/usbkey/encrypted_backup

Or using object storage with encryption using `xobs`:

    $ cabri cli --password \
        --obsrg gra --obsep https://s3.gra.cloud.ovh.net \
        --obsct encrypted_backup_container \
        --obsak access_key --obssk secret_key \
        dss make xobs:/home/guest/cabri_config/encrypted_backup

Now you are ready to synchronize your confidential data with this encrypted DSS.
As it is a rather technical topic, it deserves a [page](encrypt.md) on its own.
