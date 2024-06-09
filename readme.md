# swiss-knife middleware

This plugin allows you to protect routes with an API key specified in a header. If the user does not provide a valid key the middleware will return a 403.

You can protect routes using `X-API-KEY:$key` or `Authorization: Bearer $key` headers. Both these header names are configurable and can be toggled on/off as needed.

Valid keys are specified in a list. When a user visits a protected route and provides one of these headers, the key is looked up. If it is found in your valid keys the middleware succeeds. If the key is not found, or an incorrect header is provided, a 403 is returned to the user.

## Config
§
### Static file

Add to your Traefik static configuration

#### yaml

```yaml
experimental:
  plugins:
    swiss-knife-middleware:
      moduleName: "github.com/quangnguyen/swiss-knife"
      version: "v0.0.1"
```

#### toml

```toml
[experimental.plugins.swiss-knife]
  moduleName = "github.com/quangnguyen/swiss-knife"
  version = "v0.0.1"
```

### CLI

Add to your startup args:

```sh
--experimental.plugins.swiss-knife.modulename=github.com/quangnguyen/swiss-knife
--experimental.plugins.swiss-knife.version=v0.0.1
```

### Dynamic configuration

Configure the plugin

### yaml

```yaml
http:
  middlewares:
    verify-api-key:
      plugin:
        swiss-knife:
          authenticationHeader: true
          authenticationHeaderName: X-API-KEY
          bearerHeader: true
          bearerHeaderName: Authorization
          removeHeadersOnSuccess: true
          enableLog: true
          keys:
            - some-api-key
```

### toml

```toml
[http]
  [http.middlewares]
    [http.middlewares.verify-api-key]
      [http.middlewares.verify-api-key.plugin]
        [http.middlewares.verify-api-key.plugin.swiss-knife]
          authenticationHeader = true
          authenticationHeaderName = "X-API-KEY"
          bearerHeader = true
          bearerHeaderName = "Authorization"
          removeHeadersOnSuccess = true
          enableLog = true
          keys = ["some-api-key"]
```

### K8s CRD

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: verify-api-key
spec:
  plugin:
    swiss-knife:
      authenticationHeader: true
      authenticationHeaderName: X-API-KEY
      bearerHeader: true
      bearerHeaderName: Authorization
      removeHeadersOnSuccess: true
      enableLog: true
      keys:
        - some-api-key
```

## Usage

Use in your `IngressRoute` to protect routes.

An example using a K8s `IngressRoute`:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: my-route
spec:
  entryPoints:
    - web
  routes:
    - kind: Rule
      match: PathPrefix(`/protected-route`)
      middlewares:
        - name: verify-api-key
      services:
        - kind: Service
          name: service-name
          port: 8000
```

## Plugin options

| option                     | default           | type     | description                                                | optional   |
|:---------------------------|:------------------|:---------|:-----------------------------------------------------------|:-----------|
| `authenticationHeader`     | `true`            | bool     | Use an authentication header to pass a valid key.          | ⚠️         |
| `authenticationHeaderName` | `"X-API-KEY"`     | string   | The name of the authentication header.                     | ✅          |
| `bearerHeader`             | `true`            | bool     | Use an authorization header to pass a bearer token (key).  | ⚠️         |
| `bearerHeaderName`         | `"Authorization"` | string   | The name of the authorization bearer header.               | ✅          |
| `removeHeadersOnSuccess`   | `true`            | bool     | If true will remove the header on success.                 | ✅          |
| `keys`                     | `[]`              | []string | A list of valid keys that can be passed using the headers. | ❌          |
| `enableLog`                | `false`           | bool     | Log request                                                | ✅          |

⚠️ - Is optional but at least one of `authenticationHeader` or `bearerHeader` must be set to `true`.

❌ - Required.

✅ - Is optional and will use the default values if not set.
