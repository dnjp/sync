# Zync

Zync is a utility for backing up your files and folders to [IPFS](https://ipfs.io/). In it's current state, Zync is sort of like a primitive version of [Dropbox](https://www.dropbox.com/home) that can be used on the command line. Files and directories managed with `zync` will be continuously backed up to IPFS when they are changed. The list of all of your actively managed files is also backed up to IPFS, providing you a single [CID](https://docs.ipfs.io/concepts/content-addressing/) that can be used to restore all managed files to their original location. 

## How does this work?

There are two components that make up Zync - the daemon (`zyncd`) and the command line client (`zync`). Once installed, `zyncd` will be launched by the daemon manager for your operating system - [launchd](https://en.wikipedia.org/wiki/Launchd) on MacOS and [systemd](https://en.wikipedia.org/wiki/Systemd) on Linux. `zync`, the command line client, is the tool you use to add and remove files as well as list those that are already managed.

## Building

Zync depends on having recent versions of [Go](https://go.dev/learn/) and
[Protobuf](https://grpc.io/docs/protoc-installation/) installed. The linked
guides will take you through the process of getting them installed on your
system.

Once the dependencies are installed, you can install Zync by cloning this
repository and running `sudo make
install`:

```
$ git clone https://github.com/dnjp/zync.git
$ cd ./zync
$ sudo make install
$ which zync
/usr/local/bin/zync
$ which zyncd
/usr/local/bin/zyncd
```

