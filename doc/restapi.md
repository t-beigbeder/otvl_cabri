# Cabri REST API documentation

The [CLI](gscli.md) page explains how to enable remote access to a DSS using a Web server using the CLI.
A Web server can also be used to provide the same kind of utilities than the CLI
through a Web API based on HTTP REST services.

This page explains how to work with the REST API.

## REST API server

It is important to have in mind that the REST API server behaves the same as the CLI:

- it has access to the local configuration containing user identities
- it has access to the local metadata indexes
in which information is always decrypted
- it can access a remote DSS served by another Cabri HTTP server

On the contrary, a Cabri HTTP server for remote access:

- does not need to use user identities, as it only uses public keys
provided by the client when needed
- keep remote indexes encrypted if the DSS are encrypted

which guarantees that deployments in cloud environments keep confidential data secret
in any circumstances.

As a consequence, the REST API server will, in most cases, be used
on the local host in order to provide a Cabri API to other languages than Golang.

## Launching the REST API server

Options to launch the Web server are

- by default _remote_, as we already saw to enable CLI multi user access
- `rest`: `webapi` subcommand providing a REST Web API along with the set of DSS client options

For instance:

    $ cabri webapi rest olf+http://localhost:3000/home/guest/olf_server@demo \
        olf+http://localhost:3000/home/guest/olf_server2@demo2 &

will launch a REST Web API server to access those two local `olf` DSS using
their respective URL paths `demo` and `demo2`.

Accessing a remote server involves a more complicated syntax, such as:

    $ cabri webapi rest webapi+https+http://localhost:3000/remotehost:3443/rdemo@demo

to launch a REST Web API server to access a remote webapi server 
`webapi+https://remotehost:3443/rdemo`

## The REST API

The REST API follows the simple rules below:

- namespace or content path is appended at the end of the URL
- create or update a namespace with POST 
- create or update content with PUT
- get namespace or content data with GET
- get namespace or content metadata with GET adding the query parameter `meta`
- delete namespace or content with DELETE

For POST or PUT, the following query parameters are used:

- `mtime`: modification time, either RFC3339 (eg 2020-08-13T11:56:41Z) or a unix time integer
- `acl`: see the CLI reference for syntax

for POST, the namespace children are provided as `child` query parameter, with the child name
ending with "/" in the case of a sub-namespace.

Following sample helps to clarify:

    cabri cli dss make olf:/home/guest/cabri_olf/olfsimpleacl -s s --ximpl bdb --pfile /tmp/pf
    cabri webapi rest olf+http://localhost:3000/home/guest/cabri_olf/olfsimpleacl@demo/ --pfile /tmp/pf --haslog
    
    curl -X POST  -H "Content-Type: application/json" "http://0.0.0.0:3000/demo/?mtime=2023-06-14T19:04:44Z&child=d1/&child=f1"
    curl -X GET "http://0.0.0.0:3000/demo/"
    ["d1/","f1"]
    curl -X GET "http://0.0.0.0:3000/demo/?meta"
    {"path":"/","mtime":1686769484,"size":7,"ch":"521ecf89977e207c7528c94f6afa99b4","isNs":true,"children":["d1/","f1"],"acl":null,"itime":1686762504746623437,"ech":"","emid":""}
    date > /tmp/guest.sample
    curl -X PUT  -H "Content-Type: application/octet-stream" "http://0.0.0.0:3000/demo/f1?mtime=2023-06-14T19:05:45Z" --data-binary @/tmp/guest.sample
    curl -X GET "http://0.0.0.0:3000/demo/f1?meta"
    {"path":"f1","mtime":1686769545,"size":33,"ch":"103d8beb9d0f106325c788860e1c6ef9","isNs":false,"children":null,"acl":null,"itime":1686762800350494070,"ech":"","emid":""}
    curl -X GET  "http://0.0.0.0:3000/demo/f1"
    Wed 14 Jun 2023 07:12:54 PM CEST
    curl -X DELETE "http://0.0.0.0:3000/demo/f1"
    curl -X GET "http://0.0.0.0:3000/demo/"
    ["d1/"]