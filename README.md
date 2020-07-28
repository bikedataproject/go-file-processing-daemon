# go-file-processing-daemon

<p align="center">
  <a href="https://github.com/bikedataproject/go-file-processing-daemon">
    <img src="https://avatars3.githubusercontent.com/u/64870976?s=200&v=4" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">Go File Processing Daemon</h3>

  <p align="center">
    This repository goal is to process files that have been uploaded to the server.
    <br />
    <a href="https://github.com/bikedataproject/go-file-processing-daemon/issues">Report Bug</a>
    ·
    <a href="https://github.com/bikedataproject/go-file-processing-daemon/issues">Request Feature</a>
  </p>
</p>

![Docker Image CI](https://github.com/bikedataproject/go-file-processing-daemon/workflows/Docker%20Image%20CI/badge.svg)![Docker Image CD](https://github.com/bikedataproject/go-file-processing-daemon/workflows/Docker%20Image%20CD/badge.svg)

This repository contains a service to process files that have been uploaded to the server. It will scan the mounted volume each minute for new files, and process their contents. 

Currently allowed filetypes are:

- [FIT](https://wiki.openstreetmap.org/wiki/FIT#:~:text=FIT%20or%20Flexible%20and%20Interoperable,including%20the%20Edge%20and%20Forerunner.&text=In%20the%20GUI%20it%20is,(FIT)%20Activity%20filefit%22.)
- [GPX](https://wiki.openstreetmap.org/wiki/GPX)
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
