# Kontrol

## Development

Use the following command to start KK in dev mode (hot-reload):

```bash
./dev-start-kk.sh --apply-directly
```

## Updating the API from the public repo

```bash
# update the api the latest hash in the branch
go get github.com/kurtosis-tech/kardinal/libs/cli-kontrol-api
go get github.com/kurtosis-tech/kardinal/libs/manager-kontrol-api
# update the go mod file and nix toml file
../scripts/go-tidy-all.sh
```
