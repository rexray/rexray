### Example libStorage Config with TLS
```
libstorage:
  tls:
    serverName: libstorage-server
    clientCertRequired: true
    trustedCertsFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-ca.crt
  service: vfs
  logging:
    httpRequests: true
    httpResponses: true
  client:
    libstorage:
      tls:
        certFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-client.crt
        keyFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-client.key
  server:
    libstorage:
      tls:
        certFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-server.crt
        keyFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-server.key
    services:
      vfs:
        libstorage:
          storage:
            driver: vfs
      mock:
        libstorage:
          storage:
          driver: mock
```

### Example REX-Ray Config with TLS
```
rexray:
  modules:
    default-docker:
      libstorage:
        tls:
          serverName: libstorage-server
          clientCertRequired: true
          trustedCertsFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-ca.crt
        service: vfs
        logging:
          httpRequests: true
          httpResponses: true
        client:
          libstorage:
            tls:
              certFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-client.crt
              keyFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-client.key
        server:
          libstorage:
            tls:
              certFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-server.crt
              keyFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-server.key
          services:
            vfs:
              libstorage:
                storage:
                  driver: vfs
            mock:
              libstorage:
                storage:
                  driver: mock
```

It's also possible to disable TLS without removing all the keys. Under the `tls` key (at any of the known scopes), place `disabled: true`. For example, here's the libStorage config with all of the TLS settings, but the server has TLS disabled:

### Example libStorage Config with TLS Disabled
```
libstorage:
  tls:
    serverName: libstorage-server
    clientCertRequired: true
    trustedCertsFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-ca.crt
  service: vfs
  logging:
    httpRequests: true
    httpResponses: true
  client:
    libstorage:
      tls:
        certFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-client.crt
        keyFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-client.key
  server:
    libstorage:
      tls:
        disabled: true
        certFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-server.crt
        keyFile: /Users/akutz/Projects/go/src/github.com/emccode/libstorage/.tls/libstorage-server.key
    services:
      vfs:
        libstorage:
          storage:
            driver: vfs
      mock:
        libstorage:
          storage:
          driver: mock
```
