# Ingress Autoswagger
A small Go application provides Swagger UI for all services listed in the environment variable. 
Typically used with Kubernetes where this app listens the root and each microservice exposed with Ingress on subdirectories.

## How it works
Assume, you have three microservices `cart`, `delivery`, and `payment` deployed on the same host.
To make this work, each application should expose [Open API JSON](https://swagger.io/specification/) on `/{version}/api-docs`. 
For example:

* `/cart/v3/api-docs`
* `/delivery/v3/api-docs`
* `/payment/v3/api-docs`

Then, run this application with environment variable `SERVICES=["cart", "delivery", "payment"]"` and expose to `/`.
The application finds the right version of the specification for each service and periodically checks the liveness of the applications.

![Main window screen](https://github.com/adeo/ingress-autoswagger/raw/master/docs/main_window.png)

## Supported environment variables

* **SERVICES** *`required`* array of services to look up
* **VERSIONS**  *`default: ["v2", "v3"]`* array of versions of specifications used in microservices
* **REFRESH_CRON** *`default: @every 1m`* schedule for check liveness of applications

## Usage

### With helm (stored inside LMRU, needs to be builded for external setups)

```bash
helm repo add lmru https://art.lmru.tech/helm
helm upgrade --install --namespace \
 --set services={plaster-calculator,product-binder} --set hostname=$hostname --set version=3.2 \
 $namespace $release-name lmru/ingress-autoswagger
```

### With docker (stored inside LMRU, needs to be builded for external setups)

```bash
docker run -it -e SERVICES="[\"plaster-calculator\",\"product-binder\"]" -e VERSIONS="[\"v2\",\"v3\"]" docker-devops.art.lmru.tech/bricks/ingress-autoswagger:3.1
```

### Without docker

```bash
SERVICES="[\"plaster-calculator\",\"product-binder\"]" VERSIONS="[\"v2\",\"v3\"]" go run ingress-autoswagger.go 
```

After run you can open http://localhost:3000 in browser.

## Development & Build

0. The tool written in simple Go language, so one that you need it to have installed Go.
1. Install dependencies
go get -u github.com/gobuffalo/packr/packr
2. Build with packr (syntax the same with typical 'go build' command)
packr build .

## Maintainers

Dmitrii Sugrobov @voborgus

Nikita Medvedev @MisterRnobe

Stanislav Myachenkov @smyachenkov
