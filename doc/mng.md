# DSS management

The on-line help for subcommands that manage DSS is provided with

    $ cabri cli dss
    Cabri DSS management calling subcommands
    
    Usage:
    cabri cli dss [command]
    
    Available Commands:
    audit       audit a DSS check files against index
    config      updates and/or displays the DSS configuration
    lshisto     list namespace or entry full history information
    make        create a new DSS
    mkns        create a namespace
    reindex     reindex a DSS
    rmhisto     removes history entries for a given time period
    scan        scan a DSS
    unlock      unlock a DSS

Creating a new DSS or a namespace and management of the DSS configuration
are documented in "getting started" pages,
this page focuses on other subcommands, dealing with indexes and history.
As `fsy` filesystem native DSS don't have indexes and don't manage history,
only object-like `olf` or object storage `obj` DSS are supported by these commands.

## Index management

Indexes are necessary to provide correct performance when accessing data in DSS.
There are two kind of indexes:

- DSS index: its role is to index the metadata for the DSS
  - when the DSS is encrypted, the metadata in the DSS and the index are too
  - when the DSS is shared over http(s), the index stays on the server side
  and it tracks what updates are known by which clients
- client index: its role is to index the metadata on the client side
  - when the DSS is encrypted, a client index is always created,
  the metadata in the client index appears clear, as well as the client data
  extracted from the DSS for local copy (no security issue here)
  - when the DSS is accessed over http(s), a client index is always created,
  extracting from the server metadata information that is in use by the client

As indexes duplicate DSS metadata, they may be rebuild if broken or inconsistent,
but this is a time-consuming operation as it requires a full scan of the DSS data.

The following subcommands are provided:

- audit: performs an audit of the DSS index, check DSS files against its index
and displays inconsistencies
- reindex: rebuilds the DSS index from a full scan
- unlock: removes the lock of a DSS index (local) or client index ((x)webapi+http(s))

For instance:

    mkdir /home/guest/olf_server
    cabri cli dss make olf:/home/guest/olf_server -s s --ximpl bdb
    cabri webapi olf+http://localhost:3000/home/guest/olf_server@demo &
    cabri cli sync fsy:/home/guest/cabri_samples/consistent@ webapi+http://localhost:3000/demo@ -rv

Stop the server to unlock the index, then

    cabri cli dss audit olf:/home/guest/olf_server

In case of inconsistencies, simply reindex

    cabri cli dss reindex olf:/home/guest/olf_server

In case of remote access, the client index can be managed the same way, for instance:

    cabri webapi olf+http://localhost:3000/home/guest/olf_server@demo &
    cabri cli dss audit webapi+http://localhost:3000/demo
    cabri cli dss reindex webapi+http://localhost:3000/demo

## History management

DSS store all history for namespaces and content entries.
This can be displayed with the lshisto subcommand.
For instance:

    $ cabri cli dss make olf:/home/guest/cabri_olf/olfsimpleacl -s s --ximpl bdb
    $ cabri cli sync fsy:/home/guest/cabri_samples/simple@ olf:/home/guest/cabri_olf/olfsimpleacl@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -r
    $ cabri cli dss lshisto olf:/home/guest/cabri_olf/olfsimpleacl@d1/d11/ -r
    "d1/d11/"
    2023-05-07T08:38:44/....-..-..T..:..:..            3 2ba85baaa7922ff4c0dfdbc00fd07bd6 d1/d11/
    "d1/d11/f3"
    2023-05-07T08:38:44/....-..-..T..:..:..            3 2ba85baaa7922ff4c0dfdbc00fd07bd6 d1/d11/f3

Update the source DSS, synchronize again and display the `olf` history:

    $ date >> /home/guest/cabri_samples/simple/d1/d11/f3
    $ date >> /home/guest/cabri_samples/simple/d1/d11/f3bis
    $ cabri cli sync fsy:/home/guest/cabri_samples/simple@ olf:/home/guest/cabri_olf/olfsimpleacl@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -r
    $ cabri cli dss lshisto olf:/home/guest/cabri_olf/olfsimpleacl@d1/d11/ -r
    "d1/d11/"
    2023-05-07T08:38:44/2023-05-07T08:46:19            3 2ba85baaa7922ff4c0dfdbc00fd07bd6 d1/d11/
    2023-05-07T08:46:20/....-..-..T..:..:..            9 01cf5bbc10aba25ed41d6ea371192377 d1/d11/
    "d1/d11/f3"
    2023-05-07T08:38:44/2023-05-07T08:46:19            3 2ba85baaa7922ff4c0dfdbc00fd07bd6 d1/d11/f3
    2023-05-07T08:46:20/....-..-..T..:..:..           36 e32621954bcf857cb453abcc28f57a25 d1/d11/f3
    "d1/d11/f3bis"
    2023-05-07T08:46:20/....-..-..T..:..:..           33 0884ebdf8cae1d8a937bb7813591e425 d1/d11/f3bis

Unused DSS entries can be removed in the history with the `rmhisto` subcommand:

    $ cabri cli dss rmhisto
    Usage:
    cabri cli dss rmhisto [flags]
    
    Flags:
    -d, --dryrun      don't remove the history, just report work to be done
    --et string   the inclusive index time below which entries must be removed, default to all future entries
    -h, --help        help for rmhisto
    -r, --recursive   recursively remove the history of all namespace children
    --st string   inclusive index time above which entries must be removed, default to all past entries

In the example above, entries up to `2023-05-07T08:46:19` may be removed with:

    $ cabri cli dss rmhisto --et 2023-05-07T08:46:19Z olf:/home/guest/cabri_olf/olfsimpleacl@d1/d11/ -r

## Remove unused content

The `dss scan` subcommand can be used to locate and remove 
unused data content after having removed some entries with the
`dss rmhisto` subcommand:

    $ cabri cli dss scan
    Usage:
    cabri cli dss scan [flags]
    
    Flags:
    -f, --full           if summary requested, performs a full scan
    -h, --help           help for scan
    --hidden         also purge hidden metadata and content
    --purge          purge unused content
    -r, --resol string   if summary requested, resolution s, m, h, d from seconds to days to display the result (default "s")
    -s, --summary        don't scan, only provide a summary of time periods

Using the `--purge` flag:

    $ cabri cli dss scan olf:/home/guest/cabri_olf/olfsimpleacl --purge --pfile /home/guest/secrets/cabri
    Error: Collected errors:
        Error 0: /home/guest/cabri_olf/olfsimpleacl/content/9c/71185977b6dfe6a2023af4401f91f8 (ch 9c71185977b6dfe6a2023af4401f91f8) is not used anymore
