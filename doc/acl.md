# Using ACL with cabri DSS

## Mapping system users to DSS users

When synchronizing files among different systems it is often useful to map
system users with DSS users.

On unix-like systems, the system access rights names associated
to the files are respectively:

- x-uid:<uid> for the user permissions, uid is the file's owner user id,
by convenience the empty name can also be used to refer to it
- x-gid:<gid> for the group permissions, gid is the file's group id
- x-other for the _other_ permissions

On non-unix systems, the file owner name is mapped to an empty name.

DSS users are simply textual labels, except when using encryption
where they refer to identities' aliases.

Synchronization ACLs take the form `--acl <user>:<rights>` where the
rights take one of more letters in:
- r: read access
- w: write access
- x: execute access, or file access in a directory

Empty right means `rwx`

Synchronization also uses a list of ACL mappings to map source users with target users.
They take the form `<left-user>:<right-user>`. Unmapped users are kept as they are in the source.

For instance, sharing a source directory `cabri_samples/simple`
with two users: `u1` in read-write and `u2` in read-only
into an `olf` DSS `cabri_olf/olfsimpleacl` 

    cabri cli dss make olf:/home/guest/cabri_olf/olfsimpleacl -s s
    cabri cli sync fsy:/home/guest/cabri_samples/simple@ olf:/home/guest/cabri_olf/olfsimpleacl@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -r

To synchronize it back with both users in dedicated directories, simply use

    cabri cli sync olf:/home/guest/cabri_olf/olfsimpleacl@ fsy:/home/guest/cabri_samples/simpleback1@ --macl u1: -r
    cabri cli sync olf:/home/guest/cabri_olf/olfsimpleacl@ fsy:/home/guest/cabri_samples/simpleback2@ --acl :rx --macl u2: -r

## Multi-user synchronization

Using the previously discussed mapping between users and rights,
we can easily implement a multi-user synchronization.
Given the previous example, assuming that `cabri_samples/simple` contains
three directories

- `d1` that is owned by user `u1`,
- `d2` that is owned by user `u2`,
- `d3` that is shared by both

Sharing into an `olf` DSS `cabri_olf/olfsimplesas` 

    cabri cli dss make olf:/home/guest/cabri_olf/olfsimplesas -s s
    cabri cli dss mkns olf:/home/guest/cabri_olf/olfsimplesas@ -c d1/ -c d2/ -c d3/ --acl u1: --acl u2:
    cabri cli sync fsy:/home/guest/cabri_samples/simple@d1 olf:/home/guest/cabri_olf/olfsimplesas@d1 --acl u1: --acl u2:rx --macl :u1 --macl :u2 -r
    cabri cli sync fsy:/home/guest/cabri_samples/simple@d2 olf:/home/guest/cabri_olf/olfsimplesas@d2 --acl u1:rx --acl u2: --macl :u1 --macl :u2 -r
    cabri cli sync fsy:/home/guest/cabri_samples/simple@d3 olf:/home/guest/cabri_olf/olfsimplesas@d3 --acl u1: --acl u2: --macl :u1 --macl :u2 -r

To synchronize them back with both users in dedicated directories, simply use

    mkdir /home/guest/cabri_samples/simplesync1/d1 /home/guest/cabri_samples/simplesync1/d2 /home/guest/cabri_samples/simplesync1/d3
    cabri cli sync fsy:/home/guest/cabri_samples/simplesync1@d1 olf:/home/guest/cabri_olf/olfsimplesas@d1 --acl u1: --acl u2:rx --macl :u1 --macl :u2 --bidir -u u1 -r
    cabri cli sync olf:/home/guest/cabri_olf/olfsimplesas@d2 fsy:/home/guest/cabri_samples/simplesync1@d2 --acl :rx --macl u1: --leftuser u1 -r
    cabri cli sync fsy:/home/guest/cabri_samples/simplesync1@d3 olf:/home/guest/cabri_olf/olfsimplesas@d3 --acl u1: --acl u2: --macl :u1 --macl :u2 --bidir -u u1 -r

    mkdir /home/guest/cabri_samples/simplesync2/d1 /home/guest/cabri_samples/simplesync2/d2 /home/guest/cabri_samples/simplesync2/d3
    cabri cli sync olf:/home/guest/cabri_olf/olfsimplesas@d1 fsy:/home/guest/cabri_samples/simplesync2@d1 --acl :rx --macl u2: --leftuser u2 -r
    cabri cli sync fsy:/home/guest/cabri_samples/simplesync2@d2 olf:/home/guest/cabri_olf/olfsimplesas@d2 --acl u1:rx --acl u2: --macl :u1 --macl :u2 --bidir -u u2 -r
    cabri cli sync fsy:/home/guest/cabri_samples/simplesync2@d3 olf:/home/guest/cabri_olf/olfsimplesas@d3 --acl u1: --acl u2: --macl :u1 --macl :u2 --bidir -u u2 -r
