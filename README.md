# reader-file-server

This is a file server for a (private) reader application.
It allows you to read all files and also write to some of them.

## Why?

This is a backend implementation for a private book reader application.

## Environment Variables

- `ROOT`: The root directory of the file server. Default: `.`
- `CORS_ORIGINS`: The origins that are allowed to access the file server. Default: `*`
- `ACCESS_TOKEN`: The access token that is required to write to the file server.

### Access Token

Access token must be set using the `authorization` header with the `Token` auth scheme.

