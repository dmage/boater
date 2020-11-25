# Boater

A Docker Registry HTTP API client.

# Installation

```
go get -u github.com/dmage/boater
```

# Usage

## Get an image manifest

```console
$ boater get-manifest -a ubuntu | head -c120
{"manifests":[{"digest":"sha256:1d7b639619bdca2d008eca2d5293e3c43ff84cbee597ff76de3b7a7de3e84956","mediaType":"applicati
```

## Mimic a client that supports only schema 2 manifests

```console
$ boater get-manifest ubuntu --accept-schema2
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 3352,
      "digest": "sha256:d70eaf7277eada08fca944de400e7e4dd97b1262c06ed2b1011500caa4decaf1"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 28558714,
         "digest": "sha256:6a5697faee43339ef8e33e3839060252392ad99325a48f7c9d7e93c22db4d4cf"
      },
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 847,
         "digest": "sha256:ba13d3bc422b493440f97a8f148d245e1999cb616cb05876edc3ef29e79852f2"
      },
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 162,
         "digest": "sha256:a254829d9e55168306fd80a49e02eb015551facee9c444d9dce3b26d19238b82"
      }
   ]
}
```

## Get an image blob

```console
$ boater get-blob ubuntu sha256:7b9b13f7b9c086adfb6be4d2d264f90f16b4d1d5b3ab9f955caa728c3675c8a2 | head -c120
{"architecture":"amd64","config":{"Hostname":"abff9e7e2995","Domainname":"","User":"","AttachStdin":false,"AttachStdout"
```

## View all HTTP requests

```console
$ boater get-manifest ubuntu -v 2>&1 | grep HTTP
2017/06/14 00:32:10 > GET https://index.docker.io/v2/ HTTP/1.1
2017/06/14 00:32:11 < HTTP/1.1 401 Unauthorized
2017/06/14 00:32:11 > GET https://auth.docker.io/token?scope=repository%3Alibrary%2Fubuntu%3Apull&service=registry.docker.io HTTP/1.1
2017/06/14 00:32:11 < HTTP/1.1 200 OK
2017/06/14 00:32:11 > GET https://index.docker.io/v2/library/ubuntu/manifests/latest HTTP/1.1
2017/06/14 00:32:12 < HTTP/1.1 200 OK
```

## Copy a schema 2 image from one repository to another

```console
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

```console
REPO=my/repo USER=user PASSWORD=password
for file in ./ubuntu/blobs/*; do
    boater put-blob -u "$USER" -p "$PASSWORD" "$REPO" "$file"
done
boater put-manifest -u "$USER" -p "$PASSWORD" "$REPO" ./ubuntu/manifest \
    --content-type="application/vnd.docker.distribution.manifest.v2+json"
```

### Alternatives

  * [reg](https://github.com/genuinetools/reg)
