# traefik-http-shaping

HTTP Shaping is a middleware plugin for [traefik](https://github.com/traefik/traefik) allowing to track IN/OUT traffic of an HTTP endpoint and limit network usage. Once any of the limit is reached, the HTTP request returns a too many requests status code.

## Adding the middleware

In traefik.yml, using the dev method, add:

    experimental:
      localPlugins:
        httpShaping:
          moduleName: github.com/martinlevesque/http_shaping

and copy the plugin into your local folder:

    mkdir -p plugins-local/src/github.com/martinlevesque/http_shaping
    git clone https://github.com/martinlevesque/http_shaping.git plugins-local/src/github.com/martinlevesque/http_shaping

## Configurations

    http:
      routers:
        my-router:
          rule: host(`localhost`)
          service: service-foo
          entryPoints:
            - web
          middlewares:
            - my-plugin

      services:
       service-foo:
          loadBalancer:
            servers:
              - url: http://127.0.0.1:5000

      middlewares:
        my-plugin:
          plugin:
            httpShaping:
              InTrafficLimit: 200MiB
              OutTrafficLimit: 200MiB
              LoopInterval: 60
              ConsiderLimits: true

These configs mean that 200 MiB per 60 seconds are allowed (IN and OUT).