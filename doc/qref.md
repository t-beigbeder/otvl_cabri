# Cabri quick reference

## Configuration set up

### Initialization

Because Cabri stores some secrets such as S3 secret keys, an encryption key has to be used.

    $ cabri cli config --dump
    {
    "clientId": "0e4c659a-d084-466a-b725-21c27d069a21",
    "Identities": [
    {
    "alias": "",
    "pKey": "age1z66mme7e4mv0c6x6s3emrqgflphn3zt2pjtx7k9vck7xn0pw29usufue4r",
    "secret": "AGE-SECRET-KEY-1RS099MRK3WRMLARVESXPFLG754EF0CALV60JZYRMYX09CCRUNYQS46K82H"
    }
    ],
    "Internal": {
    "alias": "__internal__",
    "pKey": "age13nht4agnsza4f7z9exymgneecq3ndp0rdmjkaspe9jggwctkjqlsmm3hfg",
    "secret": "AGE-SECRET-KEY-1ASR7RXXALPPQLSMHQDHW3C8YM9AG0ZJMAXVSWNQD75TXUMS6S0FQCDUTL6"
    }
    }

- The empty alias key is the default identity for encrypting data in DSS.
- The `__internal__` key encrypts DSS configuration.

Copy this information in a safe place (such as Keepass) and make a backup.

### Master password encryption

The previous configuration is better being encrypted if the workstation where it is stored is at risk.

    $ cabri cli config --encrypt
    please enter the master password:
    please enter the master password again:

Now all commands need to provide a password (`--password` or `--pfile` options)

## Synchronize files like `rsync` does

Download a reference dataset:

    $ wget https://github.com/torvalds/linux/archive/v5.9.zip && unzip v5.9.zip

Synchronize source with target, not using checksums:

    $ mkdir copy-v5.9
    $ cabri cli sync fsy:linux-5.9@ fsy:copy-v5.9@ -rvn --summary
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0

Check copy using checksums:

    $ cabri cli sync fsy:linux-5.9@ fsy:copy-v5.9@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

## Synchronize remote files like `rsync` does

Launch the remote server (add TLS and authentication) mapping reference dataset to URL path `remote-sample`:

    $ ssh remotehost    
    $ cabri webapi fsy+http://remotehost:3000/home/guest/Downloads/linux-5.9@remote-sample

On the client side synchronize remote source with local target, reducing network usage:

    $ mkdir copy-of-remote-v5.9
    $ cabri cli sync wfsapi+http://remotehost:3000/remote-sample@ fsy:copy-of-remote-v5.9@ -rvn --summary --reducer 3
