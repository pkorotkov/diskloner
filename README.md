# Diskloner

This is a handy console app for cloning disks and partitions. For now, it supports only Linux-based OS.

## Things to do before use

The app currently requires enough permissions to own `/var/run/diskloner` directory. The simplest way to do it is here:
```
sudo mkdir -p /var/run/diskloner
sudo chown -R <user-name>:<user-group> /var/run/diskloner
```

## Building from source

### Prerequisites

1. Recent version of [Go](https://golang.org/dl/) installed.

### Go-getting the project

To get the latest stable version of `diskloner` execute the below command:
```
go get -u -v -t github.com/pkorotkov/diskloner
```
