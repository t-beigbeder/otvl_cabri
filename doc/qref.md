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
    ...
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0

Check copy using checksums:

    $ cabri cli sync fsy:linux-5.9@ fsy:copy-v5.9@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

## Synchronize files from remote like `rsync` does

Launch the remote server (add TLS and authentication when relevant)
mapping reference dataset to URL path `remote-sample`:

    $ ssh remotehost    
    # Download reference dataset a previous
    $ cabri webapi fsy+http://remotehost:3000/home/guest/Downloads/linux-5.9@remote-sample

On the client side synchronize remote source with local target:

    $ mkdir copy-of-remote-v5.9
    $ cabri cli sync wfsapi+http://remotehost:3000/remote-sample@ fsy:copy-of-remote-v5.9@ -rvn --summary
    ...
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0

Check copy using checksums:

    $ cabri cli sync wfsapi+http://remotehost:3000/remote-sample@ fsy:copy-v5.9@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

## Backup files into an `olf` DSS

Create an `olf` (object-like files) DSS with an index:

    $ mkdir /home/guest/Downloads/olf-sample
    $ cabri cli dss make olf:/home/guest/Downloads/xolf-sample -s m --ximpl bdb

Synchronize source files with target DSS, not using checksums:

    $ cabri cli sync fsy:linux-5.9@ olf:/home/guest/Downloads/xolf-sample@ -rvn --summary
    ...
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0

Check copy using checksums:

    $ cabri cli sync fsy:linux-5.9@ olf:/home/guest/Downloads/xolf-sample@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

## Incremental backup of files into an `olf` DSS

Download an update to the reference dataset:

    $ wget https://github.com/torvalds/linux/archive/v6.5.zip && unzip -q v6.5.zip

Evaluate need to synchronize target DSS, using checksums:

    $ cabri cli sync fsy:linux-6.5@ olf:/home/guest/Downloads/olf-sample@ -rd --summary
    ...
    created: 22295, updated 39093, removed 10660, kept 0, touched 24930, error(s) 0

Do it:

    $ cabri cli sync fsy:linux-6.5@ olf:/home/guest/Downloads/olf-sample@ -rv --summary
    ...
    created: 22295, updated 39093, removed 10660, kept 0, touched 24930, error(s) 0

Check backup:

    $ cabri cli sync fsy:linux-6.5@ olf:/home/guest/Downloads/olf-sample@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

## Restore from `olf` DSS history

Display the DSS history:

    $ cabri cli dss scan olf:/home/guest/Downloads/olf-sample --summary -r m
    2023-10-26T12:36:00/2023-10-26T12:38:00    79394
    2023-10-26T12:49:00/2023-10-26T12:52:00    88744

Seems like the version 5.9 was backed up until `2023-10-26T12:38:00`.

Display it briefly:

    $ cabri cli lsns olf:/home/guest/Downloads/olf-sample@ -t --lasttime 2023-10-26T12:38:00Z
    ...
    1965 2020-10-11 23:15:50 scripts/

Restore version v6.5 (latest state of the DSS) and check it with original:

    $ mkdir restore6.5
    $ cabri cli sync olf:/home/guest/Downloads/olf-sample@ fsy:restore6.5@ -rv --summary
    ...
    created: 86317, updated 1, removed 0, kept 0, touched 0, error(s) 0
    $ cabri cli sync fsy:linux-6.5@ fsy:restore6.5@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

Restore version v5.9 and check it:

    $ mkdir restore5.9
    $ cabri cli sync olf:/home/guest/Downloads/olf-sample@ fsy:restore5.9@ --lefttime 2023-10-26T12:38:00Z -rv --summary
    ...
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0
    $ cabri cli sync fsy:linux-5.9@ fsy:restore5.9@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0
