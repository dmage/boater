# Boater

A Docker Registry HTTP API client.

# Installation

```
go get -u github.com/dmage/boater
```

# Usage

```
boater get-manifest ubuntu
boater get-manifest ubuntu --accept-schema2
boater get-blob ubuntu sha256:7b9b13f7b9c086adfb6be4d2d264f90f16b4d1d5b3ab9f955caa728c3675c8a2
```

```
$ boater get-manifest ubuntu -v 2>&1 | grep HTTP
2017/06/14 00:32:10 > GET https://index.docker.io/v2/ HTTP/1.1
2017/06/14 00:32:11 < HTTP/1.1 401 Unauthorized
2017/06/14 00:32:11 > GET https://auth.docker.io/token?scope=repository%3Alibrary%2Fubuntu%3Apull&service=registry.docker.io HTTP/1.1
2017/06/14 00:32:11 < HTTP/1.1 200 OK
2017/06/14 00:32:11 > GET https://index.docker.io/v2/library/ubuntu/manifests/latest HTTP/1.1
2017/06/14 00:32:12 < HTTP/1.1 200 OK
```

```
$ mkdir -p ./ubuntu/blobs
$ boater get-manifest ubuntu --accept-schema2 >./ubuntu/manifest
$ for digest in $(jq -r '.config.digest, .layers[].digest' ./ubuntu/manifest); do
>     boater get-blob ubuntu "$digest" >"./ubuntu/blobs/$digest"
> done
$ tree ./ubuntu
./ubuntu
|-- blobs
|   |-- sha256:1f88dc826b144c661a8d1d08561e1ff3711f527042955505e9f3e563bdb2281f
|   |-- sha256:2b61829b0db5f4033ff48cbf3495271c8410c76e6396b56f15a79c3f7b5b7845
|   |-- sha256:6960dc1aba1816652969986284410927a5d942bf8042e077a3ebc8d1c58bb432
|   |-- sha256:73b3859b1e43f3ff32f10055951a568a9ad5ab6dc4ab61818b117b6912088f3d
|   |-- sha256:7b9b13f7b9c086adfb6be4d2d264f90f16b4d1d5b3ab9f955caa728c3675c8a2
|   `-- sha256:bd97b43c27e332fc4e00edf827bbc26369ad375187ce6eee91c616ad275884b1
`-- manifest

1 directory, 7 files
```

```
REPO=my/repo USER=user PASSWORD=password
for file in ./ubuntu/blobs/*; do
    boater put-blob -u "$USER" -p "$PASSWORD" "$REPO" "$file"
done
boater put-manifest -u "$USER" -p "$PASSWORD" "$REPO" ./ubuntu/manifest \
    --content-type="application/vnd.docker.distribution.manifest.v2+json"
```

### Alternatives

  * [reg](https://github.com/genuinetools/reg)
