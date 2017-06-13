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
boater get-manifest ubuntu --accept-schema2 >./ubuntu.manifest
mkdir -p ./ubuntu.blobs/
for digest in $(jq -r '.config.digest, .layers[].digest' ./ubuntu.manifest); do
    boater get-blob ubuntu "$digest" >"./ubuntu.blobs/$digest"
done

R=my/repo U=user P=password
for f in ./ubuntu.blobs/*; do
    boater put-blob -u "$U" -p "$P" "$R" "$f"
done
boater put-manifest -u "$U" -p "$P" "$R" ./ubuntu.manifest --content-type="application/vnd.docker.distribution.manifest.v2+json"
```
