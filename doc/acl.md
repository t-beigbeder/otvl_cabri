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

TODO: to be completed
