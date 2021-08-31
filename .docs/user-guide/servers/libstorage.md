# Configuring libStorage

Tweak this, turn that, peek behind the curtain...

---

## Overview
This page reviews how to configure libStorage to suit any environment,
beginning with the the most common use cases, exploring recommended guidelines,
and finally, delving into the details of more advanced settings.

### Client/Server Configuration
Except when specified otherwise, the configuration examples below assume the
libStorage client and server exist on the same host. However, that is not at
all a requirement. It is fully possible, and in fact the entire purpose of
libStorage, that the client and server be able to function on different
systems. One libStorage server should be able to support hundreds of clients.
Yet for the sake of completeness, the examples below show both configurations
merged.

When configuring a libStorage client and server for different systems, there
will be a few differences from the examples below:

  * The examples show libStorage configured with its server component hosted
    on a UNIX socket. This is ideal for when the client/server exist on the same
    host as it reduces security risks. However, in most real-world scenarios
    the client and server are *not* residing on the same host, the
    libStorage  server should use a TCP endpoint so it can be accessed
    remotely.

  * In a distributed configuration the actual driver configuration sections
    need only occur on the server-side. The entire purpose of libStorage's
    distributed nature is to enable clients without any knowledge of how to
    access a storage platform the ability to connect to a remote server that
    maintains that storage platform access information.

## Basic Configuration
This section outlines the most common configuration scenarios encountered by
libStorage's users.

### Simple
The first example is a simple libStorage configuration with the VirtualBox
storage driver. The below example omits the host property, but the configuration
is still valid. If the `libstorage.host` property is not found, the server is
hosted via a temporary UNIX socket file in `/var/run/libstorage`.

!!! note "note"

    Please remember to replace the placeholders in the following examples
    with values valid for the systems on which the examples are executed.

    The example below specifies the `volumePath` property as
    `$HOME/VirtualBox/Volumes`. While the text `$HOME` will be replaced with
    the actual value for that environment variable at runtime, the path may
    still be invalid. The `volumePath` property should reflect a path on the
    system on which the VirtualBox server is running, and that is not always
    the same system on which the libStorage server is running.

    So please, make sure to update the `volumePath` property for the VirtualBox
    driver to a path valid on the system on which the VirtualBox server is
    running.

    The same goes for VirtualBox property `endpoint` as the VirtualBox
    web service is not always available at `10.0.2.2:18083`.

```yaml
libstorage:
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

### TCP
The following example illustrates how to configure a libStorage client and
server running on the same host. The server has one endpoint on which it is
accessible - a single TCP port, 7979, bound to the localhost network interface.

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

### TCP+TLS
The following example illustrates how to configure a libStorage client and
server running on the same host. The server has one endpoint on which it is
accessible - a single TCP port, 7979, bound to all of the host's network
interfaces. This means that the server is accessible via external clients, not
just those running on the same host.

Because of the public nature of this libStorage server, it is a good idea to
encrypt communications between client and server.

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  client:
    tls: true
  server:
    tls:
      certFile: /etc/rexray/rexray-server.crt
      keyFile: /etc/rexray/rexray-server.key
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

!!! note "note"
    Please note that in the above example the property `libstorage.client` has
    been introduced. This property is always present, even if not explicitly
    specified. It exists to override libStorage properties for the client
    only, such as TLS settings, logging, etc.

### UNIX Socket
For the security conscious, there is no safer way to run a client/server setup
on a single system than the option to use a UNIX socket. The socket offloads
authentication and relies on the file system file access to ensure authorized
users can use the libStorage API.

```yaml
libstorage:
  host: unix:///var/run/rexray/localhost.sock
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

It is possible to apply TLS to the UNIX socket. Refer to the TCP+TLS section
for applying TLS to the UNIX sockets.

### Multiple Endpoints
There may be occasions when it is desirable to provide multiple ingress vectors
for the libStorage API. In these situations, configuring multiple endpoints
is the solution. The below example illustrates how to configure three endpoints:

 endpoint | protocol    | address | tls | localhost only
----------|-------------|---------|-----|-----------
sock | unix socket | /var/run/rexray/localhost.sock | no | yes
private | tcp | 127.0.0.1:7979 | no | yes
public | tcp | \*:7980 | yes | no

```yaml
libstorage:
  host: unix:///var/run/rexray/localhost.sock
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
    endpoints:
      sock:
        address: unix:///var/run/rexray/localhost.sock
      private:
        address: tcp://127.0.0.1:7979
      public:
        address: tcp://:7980
        tls:
          certFile: /etc/rexray/rexray-server.crt
          keyFile: /etc/rexray/rexray-server.key
          trustedCertsFile: /etc/rexray/trusted-certs.crt
          clientCertRequired: true
```

With all three endpoints defined explicitly in the above example, why leave the
property `libstorage.host` in the configuration at all? When there are no
endpoints defined, the libStorage server will attempt to create a default
endpoint using the value from the property `libstorage.host`. However, even
when there's at least one explicitly defined endpoint, the `libstorage.host`
property still serves a very important function -- it is the property used
by the libStorage client to determine which to which endpoint to connect.

### Multiple Services
All of the previous examples have used the VirtualBox storage driver as the
sole measure of how to configure a libStorage service. However, it is possible
to configure many services at the same time in order to provide access to
multiple storage drivers of different types, or even different configurations
of the same driver.

The following example demonstrates how to configure three libStorage services:

service | driver
--------|--------
virtualbox-00 | virtualbox
virtualbox-01 | virtualbox
scaleio | scaleio

Notice how the `virtualbox-01` service includes an added `integration` section.
The integration definition refers to the integration interface and parameters
specific to incoming requests through this layer. In this case we defined
`libstorage.server.services.virtualbox-01` with the
`integration.volume.operations.create.default.size` parameter set. This enables all
create requests that come in through `virtualbox-01` to have a default size of
1GB. So although it is technically the same platform below the covers,
`virtualbox-00` requests may have different default values than those defined
in `virtualbox-01`.

```yaml
libstorage:
  server:
    services:
      virtualbox-00:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes-00
          controllerName: SATA
      virtualbox-01:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes-01
          controllerName: SATA
        integration:
          volume:
            operations:
              create:
                default:
                  size: 1 # GB
      scaleio:
        driver: scaleio
        scaleio:
          endpoint: https://gateway_ip/api
          insecure: true
          userName: username
          password: password
          systemName: tenantName
          protectionDomainName: protectionDomainName
          storagePoolName: storagePoolName
```

A very important point to make about the relationship between services and
endpoints is that all configured services are available on all endpoints. In
the future this may change, and libStorage may support endpoint-specific
service definitions, but for now if a service is configured, it is accessible
via any of the available endpoint addresses.

Between the three services above, clearly one major difference is that two
services host one driver, VirtualBox, and the third service hosts ScaleIO.
However, why two services for one driver, in this case, VirtualBox? Because,
in addition to services being configured to host different types of drivers,
services can also host different driver configurations. In service
`virtualbox-00`, the volume path is `$HOME/VirtualBox/Volumes-00`,
whereas for service `virtualbox-01`, the volume path is
`$HOME/VirtualBox/Volumes-01`.

## Advanced Configuration
The following sections detail every last aspect of how libStorage works and can
be configured.

### TLS Configuration
This section reviews the several supported TLS configuration options. The table
below lists the default locations of the TLS-related files.

| Directory | File | Property | Description |
| ----------|-----|-----------|------------ |
| `$REXRAY_HOME_ETC_TLS` | `libstorage.crt` | `libstorage.tls.crtFile` | The public key. |
| | `libstorage.key` | `libstorage.tls.keyFile` | The private key. |
| | `cacerts` | `libstorage.tls.trustedCertsFile` | The trusted key ring. |
| | `known_hosts` | `libstorage.tls.knownHosts` | The system known hosts file. |

If libStorage detects any of the above files, the detected files are loaded
when necessary and without any explicit configuration. However, if a file's
related configuration property is set explicitly to some other, non-default
value, the default file will not be loaded even if it is present.

#### TLS and UNIX Sockets
TLS is disabled by default when a server's endpoint or client-side host is
configured for use with a UNIX socket. This is for two reasons:

1. TLS isn't strictly necessary to provide transport security via a file-backed
socket. The file's UNIX permissions take care of which users can read and/or
write to the socket.
2. The certificate verification step in a TLS negotiation is complicated when
the endpoint to which the client is connecting is a file path, not a FQDN or
IP address. This issue can be resolved by setting the property
`libstorage.tls.serverName` on the client so that the value matches the
server certificate's common name (CN) or one of its subject alternate names
(SAN).

To explicitly enable TLS for use with UNIX sockets, set the environment
variable `LIBSTORAGE_TLS_SOCKITTOME` to a truthy value.

#### Insecure TLS
The following example illustrates how to configure the libStorage client to
skip validation of a provided server-side certificate:

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  client:
    tls: insecure
  server:
    tls:
      certFile: /etc/rexray/rexray-server.crt
      keyFile: /etc/rexray/rexray-server.key
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

!!! note "note"
    The above example instructs the client-side TLS configuration to operate in
    _insecure_ mode. This means the client is not attempting to verify the
    certificate provided by the server. This is a security risk and should not
    ever be used in production.

#### Trusted Certs File
This TLS configuration example describes how to instruct the libStorage client
to validate the provided server-side certificate using a custom trusted CA file.
This avoids the perils of insecure TLS while still enabling a privately signed
or snake-oil server-side certificate.

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  client:
    tls:
      trustedCertsFile: $HOME/.rexray/trusted-certs.crt
  server:
    tls:
      certFile: /etc/rexray/rexray-server.crt
      keyFile: /etc/rexray/rexray-server.key
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

#### Require Client-Side Certs
The final TLS example explains how to configure the libStorage server to
to require certificates from clients. This configuration enables the use of
client-side certificates as a means of authorization.

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  client:
    tls:
      certFile: $HOME/.rexray/rexray-client.crt
      keyFile: $HOME/.rexray/rexray-client.key
      trustedCertsFile: $HOME/.rexray/trusted-certs.crt
  server:
    tls:
      certFile: /etc/libstorage/rexray-server.crt
      keyFile: /etc/libstorage/rexray-server.key
      trustedCertsFile: /etc/libstorage/trusted-certs.crt
      clientCertRequired: true
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

A typical scenario that employs the above example would also involve the
server certificate to have a dual purpose as an intermediate signing authority
that has signed the allowed client certificates. Or at the very least the
server certificate would be signed by the same intermediate CA that is used
to sign the client-side certs.

#### Peer Verification
While TLS should never be configured as insecure in production, there is a
compromise that enables an encrypted connection while still providing some
measure of verification of the remote endpoint's identity -- peer verification.

When peer verification mode is enabled, TLS is implicitly configured to operate
as insecure in order to disable server-side certificate verification. This
enables an encrypted transport while delegating the authenticity of the
server's identity to the peer verification process.

The first step to configuring peer verification is to obtain the information
about the peer used to identify it. First, obtain the peer's certificate and
store it locally. This step can be omitted if the remote peer's certificate
is already available locally.

Export the name of the host and port to verify:
```bash
$ export KNOWN_HOST=google.com
$ export KNOWN_PORT=443
```

Connect to that host and save its certificate:

```bash
$ openssl s_client -connect ${KNOWN_HOST}:${KNOWN_PORT} 2>/dev/null </dev/null | \
  sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > ${KNOWN_HOST}.crt
```

Once the remote certificate is available locally it can be used to generate
its identifying information:

```bash
$ printf "${KNOWN_HOST} sha256 %s\n" \
  $(cat ${KNOWN_HOST}.crt | openssl x509 -noout -fingerprint -sha256 | cut -c20-)
```

The above command will emit the following:

```bash
google.com sha256 14:8F:93:BE:EA:AB:68:CE:C8:03:0D:0B:0D:54:C3:59:4C:18:55:5D:2D:7E:4E:8C:68:9E:D4:59:33:3C:68:96
```

The above output is a _known hosts_ entry.

!!! note "note"
    It is also possible to collapse the entire series of above steps into a
    single command:

    <pre>
    <code lang="bash">$ if [ ! -n "${KNOWN_HOST+1}" ]; then \
        KHH=1 && printf "Host? " && read -r KNOWN_HOST; fi; \
      if [ ! -n "${KNOWN_PORT+1}" ]; then \
        KHP=1 && printf "Port? " && read -r KNOWN_PORT; fi; \
      KNOWN_HPRT=${KNOWN_HOST}:${KNOWN_PORT} && \
      KNOWN_CERT=${KNOWN_HOST}.crt && \
      openssl s_client -connect $KNOWN_HPRT 2>/dev/null &lt;/dev/null | \
        sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > $KNOWN_CERT && \
      printf "${KNOWN_HOST} sha256 %s\n" \
        $(cat $KNOWN_CERT |
        openssl x509 -noout -fingerprint -sha256 | cut -c20-) && \
      rm -f $KNOWN_CERT && \
      if [ "$KHH" = "1" ]; then unset KHH && unset KNOWN_HOST; fi; \
      if [ "$KHP" = "1" ]; then unset KHP && unset KNOWN_PORT; fi</code></pre>

      The above command will prompt a user to enter both the `Host` and `Port`
      if the `KNOWN_HOST` and `KNOWN_PORT` environment variables are not
      set to non-empty values.

##### Simple Peer Verification
With the remote peer's identifying information in hand it is possible to
enable peer verification on the client by setting the property
`libstorage.tls` to the remote peer's identifier string:

```yaml
libstorage:
  client:
    tls: google.com sha256 14:8F:93:BE:EA:AB:68:CE:C8:03:0D:0B:0D:54:C3:59:4C:18:55:5D:2D:7E:4E:8C:68:9E:D4:59:33:3C:68:96
```

The above approach does result in the client attempting a TLS connection to
the configured, remote host. However, the peer verification will only be
valid for a single peer.

##### Advanced Peer Verification
While simple peer verification works for a single, remote host, sometimes it
is necessary to enable peer verification for multiple remote hosts. This
configuration requires a _known hosts_ file.

For people that use SSH the concept of a known hosts file should feel familiar.
In fact, libStorage copies the format of SSH's known hosts file entirely. The
file adheres to a line-delimited format:

```bash
ls-svr-01 sha256 15:92:77:BE:6C:90:D3:FB:59:29:9C:51:A7:DB:5C:16:55:BD:B9:9E:E7:7E:C1:9B:30:C3:74:99:21:5F:08:6A
ls-svr-02 sha256 15:92:77:BE:6C:90:D3:FB:59:29:9C:51:A7:DB:5C:16:55:BD:B9:9E:E7:7E:C1:9B:30:C3:74:99:21:5F:08:6C
ls-svr-03 sha256 15:92:77:BE:6C:90:D3:FB:59:29:9C:51:A7:DB:5C:16:55:BD:B9:9E:E7:7E:C1:9B:30:C3:74:99:21:5F:08:6D
```

The known hosts file can be specified via the property
`libstorage.tls.knownHosts`. This is the _system_ known hosts file. If this
property is not explicitly configured then libStorage checks for the file
`$REXRAY_HOME_ETC_TLS/known_hosts`. libStorage also looks for the _user_
known hosts file at `$HOME/.rexray/known_hosts`.

Thus if a known hosts file is present at either of the default system or
user locations, it's possible to take advantage of them with a configuration
similar to the following:

```yaml
libstorage:
  client:
    tls: verifyPeers
```

The above configuration snippet indicates that TLS is enabled with peer
verification. Because no known hosts file is specified the default paths are
checked for any known host files. To enable peer verification with a custom
system known hosts file the following configuration can be used:

```yaml
libstorage:
  client:
    tls:
      verifyPeers: true
      knownHosts:  /tmp/known_hosts
```

The above configuration snippet indicates that TLS is enabled and set to
peer verification mode and that the system known hosts file is located at
`/tmp/known_hosts`.

#### Known Host Conflict
Most people that use SSH have seen an error that begins with the following
text:

```bash
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@ WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED! @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
```

The above warning occurs when a remote host's fingerprint is not what
is stored in the client's known hosts file for that host. libStorage
behaves the exact same way. If the libStorage client is configured
to verify a remote peer's identity and its fingerprint is not what
is stored in the libStorage client's known hosts file, the connection
will fail.

### Authentication
In addition to TLS, the libStorage API includes support for
[JSON Web Tokens](https://jwt.io) (JWT) in order to provide authentication
and authorization.

A JWT is transmitted along with an API call in the standard HTTP `Authorization`
header as an OAuth 2.0 [Bearer](https://tools.ietf.org/html/rfc6750) token. For
example:

```
GET /volumes

Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjI2ODg1NTAsImlhdCI6MTQ5MTIzODk1MCwibmJmIjoxNDkxMjM4OTUwLCJzdWIiOiJha3V0eiJ9.3eAA7AQZUGrwA42H64qKbu8QF_AHpSsJSMR0FALnKj8
```

The above token is split into three discreet parts:

`HEADER`.`PAYLOAD`.`SIGNATURE`

The decoded header is:
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

The decoded payload is:
```json
{
  "exp": 1522688550,
  "iat": 1491238950,
  "nbf": 1491238950,
  "sub": "akutz"
}
```

The signature is the result of signing the concatenation of the header and
signature after Base64 encoding them both.

#### Global Configuration
It's possible to provision access to the libStorage API both globally and per
service. For example, the below configuration restricts all access to the
libStorage API to the bearer token from above:

```yaml
libstorage:
  server:
    auth:
      key: MySuperSecretSigningKey
      alg: HS256
      allow:
      - akutz
```

The above configuration snippet defines a new property,
`libstorage.server.auth` which contains the following child properties:
`key`, `alg`, `allow`, `deny`, and `disabled`.

The property `libstorage.server.auth.key` is the secret key used to verify
the signatures of the tokens included in API calls. If the property's value
is a valid file path then the contents of the file are used as the key. The
value of `libstorage.server.auth.alg` specifies the cryptographic algorithm
used to sign and verify the tokens. It has a default value of `HS256`. Valid
algorithms include:

Algorithm | Strength | Config Value
----------|----------|-------------
[ECDSA](https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm) | 256 | `ES256`
 | 384 | `ES384`
 | 512 | `ES512`
[HMAC](https://en.wikipedia.org/wiki/Hash-based_message_authentication_code) | 256 | `HS256`
 | 384 | `HS384`
 | 512 | `HS512`
[RSA-PSS](http://www.emc.com/emc-plus/rsa-labs/historical/raising-standard-rsa-signatures-rsa-pss.htm) | 256 | `PS256`
 | 384 | `PS384`
 | 512 | `PS512`
[RSA](https://en.wikipedia.org/wiki/RSA_(cryptosystem)) | 256 | `RS256`
 | 384 | `RS384`
 | 512 | `RS512`

Both the properties `libstorage.server.auth.allow` and
`libstorage.server.auth.deny` are arrays of strings.

The values can be either the subject of the token, the entire, encoded JWT, or
the format `tag:encodedJWT`. The prefix `tag:` can be any piece of text in
order to provide a friendly means of identifying the encoded JWT.

The `allow` property indicates which tokens are valid and the `deny` property
is a way to explicitly revoke access.

!!! note "note"

    Please be aware that by virtue of defining the property
    `libstorage.server.auth.allow`, access to the server is restricted to
    requests with valid tokens only and anonymous access is no longer allowed.

There may be occasions when it is necessary to temporarily disable token-based
access restrictions. It's possible to do this without removing the token
configuration by setting the property `libstorage.server.auth.disabled`
to a boolean true value:

```yaml
libstorage:
  server:
    auth:
      disabled: true
      key:      MySuperSecretSigningKey
      allow:
      - akutz
```

#### Service Configuration
The previous section described how to restrict access at the global level.
This section reviews service-level token configurations.

```yaml
libstorage:
  server:
    services:
      ebs-00:
        driver: ebs
        auth:
          key: MySuperSecretSigningKey
          allow:
          - akutz
      ebs-01:
        driver: ebs
```

The above configuration defines a token configuration for the service `ebs-00`
but **not** the service `ebs-01`. That means that API calls to the resource
`/volumes/ebs-00` require a valid token whereas API calls to the resource
`/volumes/ebs-01` do not.

!!! note "note"

    If an API call without a token is submitted that acts on multiple
    resources and one or more of those resources is configured for token-based
    access then the call  will result in an HTTP status of 401 *Unauthorized*.

#### Global vs. Service
It is also possible to configure both global token-based access at the same
time as service token-based access. However, there are some important details
of which to be aware when doing so.

1. When combining global and service token configurations, only the global
token key is respected. Otherwise tokens would always be invalid either at
the global or service scope since a token is signed by a single key.

2. The global `allow` list grants access globally but can be overridden by
including a token in a service's `deny` list.

3. The global `deny` list restricts access globally and cannot be overridden
by including a token in a service's `allow` list. For example, consider the
following configuration snippet:

        libstorage:
          server:
            auth:
              key:   MySuperSecretSigningKey
              allow:
              - akutz
              deny:
              - cduchesne
            services:
              ebs-00:
                driver: ebs
                auth:
                  allow:
                  - cduchesne

    The above example defines a global deny list. That means that even though
    the service `ebs-00` includes `cduchesne` in its own allow list, requests
    to service `ebs-00` with `cduchesne`'s bearer token are denied because
    that token is denied globally.

#### Client Config
Up until now the discussion surrounding security tokens has been centered on
server-side configuration. However, the libStorage client can also be
configured to send tokens as part of its standard API call workflow. For
example:

```
libstorage:
  client:
    auth:
      token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjI2ODg1NTAsImlhdCI6MTQ5MTIzODk1MCwibmJmIjoxNDkxMjM4OTUwLCJzdWIiOiJha3V0eiJ9.3eAA7AQZUGrwA42H64qKbu8QF_AHpSsJSMR0FALnKj8
```

The above client-side configuration snippet defines the property
`libstorage.client.auth.token`, the JWT used for outgoing HTTP calls as
the bearer token.

The value of the `libstorage.client.auth.token` property can also be a valid
file path. The contents of the file will be read from disk and treated as the
encoded token string.

### Embedded Configuration
If libStorage is embedded into another application, such as
[`REX-Ray`](https://github.com/AVENTER-UG/rexray), then that application may
manage its own configuration and supply the embedded libStorage instance
directly with a configuration object. In this scenario, the libStorage
configuration files are ignored in deference to the embedding application.

### Configuration Methods
There are three ways to configure libStorage:

* Command line options
* Environment variables
* Configuration files

The order of the items above is also the order of precedence when considering
options set in multiple locations that may override one another. Values set
via CLI flags have the highest order of precedence, followed by values set by
environment variables, followed, finally, by values set in configuration files.

### Configuration Properties
The section [Configuration Methods](#configuration-methods) mentions there are
three ways to configure libStorage: config files, environment variables, and the
command line. However, this section will illuminate the relationship between the
names of the configuration file properties, environment variables, and CLI
flags.

Below is a simple configuration file that tells the libStorage client where
the libStorage server is hosted:

```yaml
libstorage:
  host: tcp://192.168.0.20:7979
```

The property `libstorage.host` is a string. This value can also be set via
environment variables or the command line, but to do so requires knowing the
names of the environment variables or CLI flags to use. Luckily those are very
easy to figure out just by knowing the property names.

All properties that might appear in the libStorage configuration file
fall under some type of heading. For example, take the default configuration
above.

The rule for environment variables is as follows:

  * Each nested level becomes a part of the environment variable name followed
    by an underscore `_` except for the terminating part.
  * The entire environment variable name is uppercase.

Nested properties follow these rules for CLI flags:

  * The root level's first character is lower-cased with the rest of the root
    level's text left unaltered.
  * The remaining levels' first characters are all upper-cased with the the
    remaining text of that level left unaltered.
  * All levels are then concatenated together.

The following example builds on the previous. In this case we have added logging
directives to the client instance and reference how their transformation in
the table below the example.

```yaml
  libstorage:
    host: tcp://192.168.0.20:7979
    logging:
      level: warn
      stdout:
      stderr:
      httpRequests: false
      httpResponses: false
```

The following table illustrates the transformations:

Property Name | Environment Variable | CLI Flag
--------------|----------------------|-------------
`libstorage.host`    | `LIBSTORAGE_HOST`    | `--libstorageHost`
`libstorage.logging.level`    | `LIBSTORAGE_LOGGING_LEVEL`    | `--libstorageLoggingLevel`
`libstorage.logging.stdout`    | `LIBSTORAGE_LOGGING_STDOUT`    | `--libstorageLoggingStdout`
`libstorage.logging.stderr`    | `LIBSTORAGE_LOGGING_STDERR`    | `--libstorageLoggingStderr`
`libstorage.logging.httpRequests`    | `LIBSTORAGE_LOGGING_HTTPREQUESTS`    | `--libstorageLoggingHttpRequests`
`libstorage.logging.httpResponses`    | `LIBSTORAGE_LOGGING_HTTPRESPONSES`    | `--libstorageLoggingHttpResponses`

### Inherited Properties
Referring to the section on defining
[Multiple Services](./libstorage.md#multiple-services), there is also another way
to define the TLS settings for the external TCP endpoint. The same configuration
can be rewritten and simplified in the process:

```yaml
libstorage:
  integration:
    volume:
      operations:
        create:
          default:
            size: 1 # GB
  server:
    virtualbox:
      endpoint:       http://10.0.2.2:18083
      tls:            false
      controllerName: SATA
    services:
      virtualbox-00:
        driver: virtualbox
        virtualbox:
          volumePath: $HOME/VirtualBox/Volumes-00
      virtualbox-01:
        driver: virtualbox
        virtualbox:
          volumePath: $HOME/VirtualBox/Volumes-01
```

The above example may look different than the previous one, but it's actually
the same with a minor tweak in order to simplify configuration.

While there are still two VirtualBox services defined, `virtualbox-00` and
`virtualbox-01`, neither service contains configuration information about the
VirtualBox driver other than the `volumePath` property. This is because the
change affected above is to take advantage of inherited properties.

When a property is omitted, libStorage traverses the configuration instance
upwards, checking certain, predefined levels known as "scopes" to see if the
property value exists there. All configured services represent a valid
configuration scope as does `libstorage.server`.

Thus when the VirtualBox driver is initialized and it checks for its properties,
while the driver may only find the `volumePath` property defined under the
configured service scope, the property access attempt travels up the
configuration stack until it hits the `libstorage.server` scope where the
remainder of the VirtualBox driver's properties *are* defined.

#### Overriding Inherited Properties
It's also possible to override inherited properties as is demonstrated in the
[Logging configuration example](#logging) above:

```yaml
libstorage:
  logging:
    level: warn
  integration:
    volume:
      operations:
        create:
          default:
            size: 1 # GB
  server:
    logging:
      level: info
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

Note that while the log level is defined at the root of the config, it's also
defined at `libstorage.server.logging.level`. The latter value of `info`
overrides the former value of `warn`. Also please remember that even had the
latter, server-specific value of `info` not been defined, an attempt by to
access the log level by the server would be perfectly valid since the attempt
would traverse up the configuration data until it found the log level defined
at the root of the configuration.

### Tasks Configuration
All operations received by the libStorage API are immediately enqueued into a
Task Service in order to divorce the business objective from the scope of the
HTTP request that delivered it. If a task completes before the HTTP request
times out, the result of the task is written to the HTTP response and sent to
the client. However, if the operation is long-lived and continues to execute
after the original HTTP request has timed out, the goroutine running the
operation will finish regardless.

In the case of such a timeout event, the client receives an HTTP status 408 -
Request Timeout. The HTTP response body also includes the task ID which can
be used to monitor the state of the remote call. The following resource URI can
be used to retrieve information about a task:

```
GET /tasks/${taskID}
```

For systems that experience heavy loads the task system can also be a source of
potential resource issues. Because tasks are kept indefinitely at this point in
time, too many tasks over a long period of time can result in a massive memory
consumption, with reports of up to 50GB and more.

That's why the configuration property `libstorage.server.tasks.logTimeout` is
available to adjust how long a task is logged before it is removed from memory.
The default value is `0` -- that is, do not log the task in memory at all.

While this is in contradiction to the task retrieval example above --
obviously a task cannot be retrieved if it is not retained -- testing and
benchmarks have shown it is too dangerous to enable task retention by default.
Instead tasks are removed immediately upon completion.

The follow configuration example illustrates a libStorage server that keeps
tasks logged for 10 minutes before purging them from memory:

```yaml
libstorage:
  server:
    tasks:
      logTimeout: 10m
```

The `libstorage.server.tasks.logTimeout` property can be set to any value that
is parseable by the Golang
[time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) function. For
example, `1000ms`, `10s`, `5m`, and `1h` are all valid values.

### Driver Configuration
There are three types of drivers:

  1. OS Drivers
  2. Storage Drivers
  3. Integration Drivers

#### OS Drivers
Operating system (OS) drivers enable libStorage to manage storage on
the underlying OS. Currently the following OS drivers are supported:

 Driver | Driver Name
--------|------------
Linux   | linux

The OS driver `linux` is automatically activated when libStorage is running on
the Linux OS.

#### Storage Drivers
Storage drivers enable libStorage to communicate with direct-attached or
remote storage systems. Currently the following storage drivers are supported:

| Provider              | Storage Platform  | <center>[Docker](https://docs.docker.com/engine/extend/plugins_volume/)</center> | <center>Containerized</center> |
|-----------------------|----------------------|:---:|:---:|
| Amazon EC2 | [EBS](./../storage-providers/aws.md#aws-ebs) | ✓ | ✓ |
| | [EFS](./../storage-providers/aws.md#aws-efs) | ✓ | ✓ |
| | [S3FS](./../storage-providers/aws.md#aws-s3fs) | ✓ | ✓ |
| Ceph | [RBD](./../storage-providers/ceph.md#ceph-rbd) | ✓ | ✓ |
| Dell EMC | [Isilon](./../storage-providers/dellemc.md#dell-emc-isilon) | ✓ | ✓ |
| | [ScaleIO](./../storage-providers/dellemc.md#dell-emc-scaleio) | ✓ | ✓ |
| DigitalOcean | [Block Storage](./../storage-providers/digitalocean.md#do-block-storage) | ✓ | ✓ |
| FittedCloud | [EBS Optimizer](./../storage-providers/fittedcloud.md#ebs-optimizer) | ✓ | |
| Google | [GCE Persistent Disk](./../storage-providers/google.md#gce-persistent-disk) | ✓ | ✓ |
| Microsoft | [Azure Unmanaged Disk](./../storage-providers/microsoft.md#azure-ud) | ✓ | ✓ |
| OpenStack | [Cinder](./../storage-providers/openstack.md#cinder) | ✓ | ✓ |
| VirtualBox | [Virtual Media](./../storage-providers/virtualbox.md#virtualbox) | ✓ | |


The `libstorage.server.libstorage.storage.driver` property can be used to
activate a storage drivers. That is not a typo; the `libstorage` key is repeated
beneath `libstorage.server`. This is because configuration property paths are
absolute, and when nested under an architectural component, such as
`libstorage.server`, the entire key path must be replicated.

That said, and this may seem to contradict the last point, the storage driver
property is valid *only* on the server. Well, not really. Internally the
libStorage client uses the same configuration property to denote its own
storage driver. This internal storage driver is actually how the libStorage
client communicates with the libStorage server.

#### Integration Drivers
Integration drivers enable libStorage to integrate with schedulers and other
storage consumers, such as `Docker` or `Mesos`. Currently the following
integration drivers are supported:

 Driver | Driver Name
--------|------------
Linux   | linux

The integration driver `linux` provides necessary functionality to enable
most consuming platforms to work with storage volumes.

### Volume Configuration
This section describes various global configuration options related to an
integration driver's volume operations, such as mounting and unmounting volumes.

#### Volume Properties
The properties listed below are the global properties valid for an integration
driver's volume-related properties.

parameter|description
---------|-----------
`libstorage.integration.volume.operations.mount.preempt`|Forcefully take control of volumes when requested
`libstorage.integration.volume.operations.mount.path`|The default host path for mounting volumes
`libstorage.integration.volume.operations.mount.rootPath`|The path within the volume to return to the integrator (ex. `/data`)
`libstorage.integration.volume.operations.create.disable`|Disable the ability for a volume to be created
`libstorage.integration.volume.operations.remove.disable`|Disable the ability for a volume to be removed

The properties in the next table are the configurable parameters that affect
the default values for volume creation requests.

parameter|description
---------|-----------
`libstorage.integration.volume.operations.create.default.size`|Size in GB
`libstorage.integration.volume.operations.create.default.iops`|IOPS
`libstorage.integration.volume.operations.create.default.type`|Type of Volume or Storage Pool
`libstorage.integration.volume.operations.create.default.fsType`|Type of filesystem for new volumes (ext4/xfs)
`libstorage.integration.volume.operations.create.default.availabilityZone`|Extensible parameter per storage driver

The properties in the next table are the configurable parameters that affect
the default values for volume remove requests.

parameter|description
---------|-----------
`libstorage.integration.volume.operations.remove.force`|Force remove volumes

#### Disable Create
The disable create feature enables you to disallow any volume creation activity.
Any requests will be returned in a successful manner, but the create will not
get passed to the backend storage platform.

```yaml
libstorage:
  integration:
    volume:
      operations:
        create:
          disable: true
```

#### Disable Remove
The disable remove feature enables you to disallow any volume removal activity.
Any requests will be returned in a successful manner, but the remove will not
get passed to the backend storage platform.

```yaml
libstorage:
  integration:
    volume:
      operations:
        remove:
          disable: true
```

#### Force Remove
The force remove feature enables the forced removal of a volume despite
its current contents or state:

```yaml
libstorage:
  integration:
    volume:
      operations:
        remove:
          force: true
```

#### Preemption
There is a capability to preemptively detach any existing attachments to other
instances before attempting a mount.  This will enable use cases for
availability where another instance must be able to take control of a volume
without the current owner instance being involved.  The operation is considered
equivalent to a power off of the existing instance for the device.

Example configuration file follows:

```yaml
libstorage:
  integration:
    volume:
      operations:
        mount:
          preempt: true
```

Driver|Supported
------|---------
Dell EMC Isilon|Not yet
Dell EMC ScaleIO|Yes
VirtualBox|Yes
AWS EBS|Yes
AWS EFS|No
AWS S3FS|No
Ceph RBD|No
GCE PD|Yes
Azure UD|Yes
OpenStack Cinder|Yes

#### Ignore Used Count
By default accounting takes place during operations that are performed
on `Mount`, `Unmount`, and other operations.  This only has impact when running
as a service through the HTTP/JSON interface since the counts are persisted
in memory.  The purpose of respecting the `Used Count` is to ensure that a
volume is not unmounted until the unmount requests have equaled the mount
requests.  

In the `Docker` use case if there are multiple containers sharing a volume
on the same host, the the volume will not be unmounted until the last container
is stopped.  

The following setting should only be used if you wish to *disable* this
functionality.  This would make sense if the accounting is being done from
higher layers and all unmount operations should proceed without control.

```yaml
libstorage:
  integration:
    volume:
      operations:
        unmount:
          ignoreUsedCount: true
```

Currently a reset of the service will cause the counts to be reset.  This
will cause issues if *multiple containers* are sharing a volume.  If you are
sharing volumes, it is recommended that you reset the service along with the
accompanying container runtime (if this setting is false) to ensure they are
synchronized.  

#### Volume Path Cache
In order to optimize `Path` requests, the paths of actively mounted volumes
returned as the result of a `List` request are cached. Subsequent `Path`
requests for unmounted volumes will not dirty the cache. Only once a volume
has been mounted will the cache be marked dirty and the volume's path retrieved
and cached once more.

The following configuration example illustrates the two path cache properties:

```yaml
libstorage:
  integration:
    volume:
      operations:
        path:
          cache:
            enabled: true
            async:   true
```

Volume path caching is enabled and asynchronous by default, so it's possible to
entirely omit the above configuration excerpt from a production deployment, and
the system will still use asynchronous caching. Setting the `async` property to
`false` simply means that the initial population of the cache will be handled
synchronously, slowing down the program's startup time.

#### Volume Root Path
When volumes are mounted there can be an additional path that is specified to
be created and passed as the valid mount point.  This is required for certain
applications that do not want to place data from the root of a mount point.

The default is the `/data` path.  If a value is set by
`linux.integration.volume.operations.mount.rootPath`, then the default will be
overwritten.

```yaml
libstorage:
  integration:
    volume:
      operations:
        mount:
          rootPath: /data
```

### REST Configuration
This section reviews advanced HTTP REST configuration options:

#### Parse POST Options
Normally an incoming POST request with an `opts` field does not copy the
key/value pairs in the `opts` map into the fields that match the request's
schema. For example, take a look at the following request:

```json
{
    "name": "newVolume",
    "iops": 0,
    "size": 5,
    "type": "block",
    "opts": {
        "encrypted": true,
    }
}
```

The above request is used for creating a new volume, and it appears that the
intention is to create the volume as encrypted. However, the `encrypted` field
is part of the free-form `opts` piece of the request. The `opts` map is for
data that is not yet part of the official libStorage API and schema. Certain
drivers may parse data out of the `opts` field, but it *is* driver specific.
The libStorage server does not normally attempt to parse this field's keys
and match it to the request's fields. A proper request would look like this:

```json
{
    "name":      "newVolume",
    "iops":      0,
    "size":      5,
    "type":      "block",
    "encrypted": true,
}
```

In order to make things a little easier on clients, a server can be configured
to treat the first JSON request as equal to the second. This feature is disabled
by default due to the possibility of side-effects. For example, a value in the
`opts` map could unintentionally overwrite the intended value if both keys are
the same.

To enable this feature, the libStorage server's configuration should set the
`libstorage.server.parseRequestOpts` property to true. An example YAML snippet
with this property enabled resembles the following:

```yaml
libstorage:
  server:
    parseRequestOpts: true
```

With the above property set to `true`, values in a request's `opts` map will be
copied to the corresponding key in the request proper.
