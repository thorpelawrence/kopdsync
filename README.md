# kopdsync

OPDS browser & KOReader sync server

```shell
docker run \
    --rm \
    -p 8080:8080 \
    -v ./books:/books:ro \
    -v ./data:/data \
    ghcr.io/thorpelawrence/kopdsync \
    -books /books \
    -db /data/progress.db \
    -registrations true
```
