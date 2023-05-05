# Client configuration

Here is a default client configuration created the first time it is required:

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

- The client id is used when sharing content with other users through a remote service,
but it is also used for any use of an encrypted DSS:
its role is to identify what (content and metadata) has been changed in the DSS
since the last time that the client accessed it,
and to load the index of those changes (metadata only) locally
- Identities are public keys users can use to encrypt content for themselves or others
  - each identity may be used by CLI tools through its alias
  - decrypting the content requires the use of the corresponding secret key
  - the empty alias is the user's default key-pair
- The `__internal__` alias is used to encrypt the configurations of the DSS you create locally

## Managing encryption keys

The CLI `config` subcommand is used to manage encryption keys

Display the on-line help:

    $ cabri cli config -h
    manage application configuration
    
    Usage:
    cabri cli config [args...] [flags]
    
    Flags:
    -d, --decrypt   decrypts the configuration file with master password
    --dump      dumps the configuration file
    -e, --encrypt   encrypts the configuration file with master password
    --gen       generate a new identity for one or several aliases
    --get       display an identity for one or several aliases
    -h, --help      help for config
    --put       <alias> <pkey> [<secret>] import or update an identity for an alias, secret may be unknown
    --remove    remove an identity alias

CAUTION: when you dump or display identities, make sure to keep the private key confidential.

Generate a public and a private keys with alias _u1_:

    cabri cli config --gen u1

Declares the public key of a user with alias _u2_:

    cabri cli config --put u2 age1<user u2 public key>
