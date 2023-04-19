# Information for configuring and using a secured HTTP server

This is a simple HTTPS implementation with Basic Authentication,
but a dedicated proxy in front of the Cabri server, like [traefik](https://traefik.io/traefik/),
apache or nginx, will generally be required or more effective.

Generate certificates for the client and the server, for instance self-signed for localhost:

    $ openssl req -newkey rsa:2048 \
        -new -nodes -x509 \
        -days 3650 \
        -out cert.pem \
        -keyout key.pem \
        -subj "/C=FR/ST=Haute Garonne/CN=localhost" -addext "subjectAltName = DNS:localhost"

The credentials for Basic Authentication are global to clients or to servers
sharing the same configuration directory, and they currently need to be provided as identities
with the alias `WebBasicAuth`.

On the server host generate keys for basic authentication, then list them:

    $ cabri cli config --gen WebBasicAuth
    $ cabri cli config --get WebBasicAuth
        PKey: age1<user>
        Secret: AGE-SECRET-KEY-<userPassword>

On the client side import this identity:
    
    $ cabri cli config --put WebBasicAuth age1<user> AGE-SECRET-KEY-<userPassword>

When no WebBasicAuth identity exists in the configuration of the client or the server:

- the server does not check clients credentials
- the client does not authenticate itself, leading to an _unauthorized_ error
  if the server is performing authentication

You can then launch a secured HTTP server using the https protocol,
and providing a certificate and its key, for instance:

    $ cabri webapi olf+https://localhost:3443/home/guest/olf_server@demo \
        --tlscrt cert.pem --tlskey key.pem

You can finally access the server remotely with the https protocol as well,
keeping other parameters as usual.
When a self-signed certificate is used,
the client must provide it as following, for instance:

    $ cabri cli lsns webapi+https://localhost:3443/demo@ --tlscrt cert.pem

Accepting any certificate from the server without any check is not recommended
although possible using the `--tlsnc` flag.
