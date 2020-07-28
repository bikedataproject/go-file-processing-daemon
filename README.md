# go-file-processing-daemon

![Docker Image CI](https://github.com/bikedataproject/go-file-processing-daemon/workflows/Docker%20Image%20CI/badge.svg)![Docker Image CD](https://github.com/bikedataproject/go-file-processing-daemon/workflows/Docker%20Image%20CD/badge.svg)

This repository contains a service to process files that have been uploaded to the server. It will scan the mounted volume each minute for new files, and process their contents. 

Currently allowed filetypes are:

- FIT
- GPX
- Create an issue (feature request) for new filetypes, or feel free to create a Pull Request

To ensure data privacy for the uploader of a file, the file is deleted from the volume right after it has been processed succesfully. No sensitive data from the files (especially FIT files) is being extracted except for location history and timestamps. 

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
