# Cabri quick start

Cabri is not difficult to use for basic needs.
This page provides initial information for getting quickly familiar with the tool,
either to give it a try or to progress smoothly in its usage.

Anyway you will not find a lot of explanations here,
reading the full documentation is necessary when it is missing on that page.

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

    $ wget https://github.com/torvalds/linux/archive/v5.9.zip && unzip -q v5.9.zip

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

NB: you have to use the same uid on the server than on the client,
or else you would have to map the uids, not documented here.

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
    $ cabri cli dss make olf:/home/guest/Downloads/olf-sample -s m --ximpl bdb

Synchronize source files with target DSS, not using checksums:

    $ cabri cli sync fsy:linux-5.9@ olf:/home/guest/Downloads/olf-sample@ -rvn --summary
    ...
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0

Check copy using checksums:

    $ cabri cli sync fsy:linux-5.9@ olf:/home/guest/Downloads/olf-sample@ -rd --summary
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

## Expose and access a remote `olf` DSS

Exposing a remote DSS is not different from seen above for files:

- the server maps DSS to URL path
- the client now uses the URL type `webapi` instead of `wfsapi` because the protocol is different for DSS

Create an `olf` (object-like files) DSS with an index on the remote server:

    $ ssh remotehost    
    # Create a `olf` DSS as previous
    $ mkdir /home/guest/Downloads/olf-remote-sample
    $ cabri cli dss make olf:/home/guest/Downloads/olf-remote-sample -s m --ximpl bdb

Launch the remote server (add TLS and authentication when relevant)
mapping the DSS to URL path to URL path `remote-olf-sample`:

    $ cabri webapi olf+http://remotehost:3000/home/guest/Downloads/olf-remote-sample@remote-olf-sample

On the client side synchronize local source with remote target:

    $ cabri cli sync fsy:linux-5.9@ webapi+http://remotehost:3000/remote-olf-sample@ -rvn --summary
    ...
    created: 74682, updated 1, removed 0, kept 0, touched 0, error(s) 0

Other cabri commands apply to the client view of the DSS as usual.

## Backup files into an `xobs` DSS

You could replay the previous commands using an `obs` (object storage) DSS.
That means using a DSS which, instead of using local files for its storage,
relies on object storage accessed through internet.

In this section, we switch directly to the use of an `xobs` (encrypted object storage) DSS.
The commands are still the same, only the DSS content is different, fully encrypted in that case,
just for the illustration.

We use a different reference dataset:

    $ wget https://raw.githubusercontent.com/nltk/nltk_data/gh-pages/packages/corpora/framenet_v15.zip
    $ unzip -q framenet_v15.zip

Create an Object container (or S3 bucket), here hosted for instance by OVH Cloud in France:

- Public Cloud / Object Storage / Create an Object container
- Solution Standard Object Storage - S3 API
- Region Gravelines (GRA)
- Link a user, create or reuse a user, copy its S3 access and secret keys to be used just below
- Container name `xobs-sample`

You can check the connectivity using `s3tools`:

    $ cabri cli s3tools --obsrg gra --obsep https://s3.gra.io.cloud.ovh.net/ --obsct xobs-sample \
    --obsak <access-key> --obssk <secret-key> --cnx

Create an `xobs` (encrypted object storage) DSS:

    $ mkdir /home/guest/Downloads/xobs-sample
    $ cabri cli dss make --obsrg gra --obsep https://s3.gra.io.cloud.ovh.net/ --obsct xobs-sample \
    --obsak <access-key> --obssk <secret-key> xobs:/home/guest/Downloads/xobs-sample

Synchronize source files with target DSS, not using checksums:

    $ cabri cli sync fsy:framenet_v15@ xobs:/home/guest/Downloads/xobs-sample@ -rvn --summary
    ...
    created: 12957, updated 1, removed 0, kept 0, touched 0, error(s) 0

Check copy using checksums:

    $ cabri cli sync fsy:framenet_v15@ xobs:/home/guest/Downloads/xobs-sample@ -rd --summary
    created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0

Use content as you want:

    $ cabri cli lsns xobs:/home/debian/Downloads/xobs-sample@ -t
    ...
      57 2010-09-10 22:02:23 docs/
    1564 2010-09-14 16:10:50 README.txt
    $ cabri cli dss get xobs:/home/debian/Downloads/xobs-sample@README.txt README.txt
    $ cat README.txt
    Welcome to Release 1.5 of the FrameNet data!
    ...
