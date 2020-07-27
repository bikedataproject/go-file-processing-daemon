# go-file-processing-daemon

![Docker Image CI](https://github.com/bikedataproject/go-file-processing-daemon/workflows/Docker%20Image%20CI/badge.svg)

This repository contains a service to process files that have been uploaded to the server.

## Required parameters

```sh
export CONFIG_FILEDIR="files"
export CONFIG_POSTGRESHOST="localhost"
export CONFIG_POSTGRESPORT="5432"
export CONFIG_POSTGRESPASSWORD="MyPostgresPassword"
export CONFIG_POSTGRESUSER="postgres"
export CONFIG_POSTGRESDB="bikedata"
export CONFIG_POSTGRESREQUIRESSL="require"
```
