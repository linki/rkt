# Trying out rkt 

This guide provides a short introduction to trying out rkt. 
For a more in-depth guide of a full end-to-end workflow of building an application and running it using rkt, check out the [getting-started-guide](getting-started-guide.md)

## Using rkt on Linux

`rkt` consists of a single self-contained CLI, and is currently supported on amd64 Linux.
A modern kernel is required but there should be no other system dependencies.
We recommend booting up a fresh virtual machine to test out rkt.

To download the `rkt` binary, simply grab the latest release directly from GitHub:

```
wget https://github.com/coreos/rkt/releases/download/v0.16.0/rkt-v0.16.0.tar.gz
tar xzvf rkt-v0.16.0.tar.gz
cd rkt-v0.16.0
./rkt help
```

## Trying out rkt using Vagrant

For Mac (and other Vagrant) users we have set up a `Vagrantfile`: clone this repository and make sure you have [Vagrant](https://www.vagrantup.com/) 1.5.x or greater installed.
`vagrant up` starts up a Linux box and installs via some scripts `rkt` and `actool`.
With a subsequent `vagrant ssh` you are ready to go:

```
git clone https://github.com/coreos/rkt
cd rkt
vagrant up
vagrant ssh
```

Keep in mind while running through the examples that right now `rkt` needs to be run as root for most operations.

## rkt basics

### Building App Container Images (ACIs)

rkt's native image format is ACI, defined in the [App Container spec](app-container.md).
To build ACIs, a simple way to get started is by using [`acbuild`](https://github.com/appc/acbuild).
Another good resource is the [appc build repository](https://github.com/appc/build-repository) which has resources for building ACIs from a number of popular projects and languages.
There are also tools for converting [Docker images to ACIs](https://github.com/appc/docker2aci) (although note that rkt can [also run Docker images natively](running-docker-images.md) directly from Docker repositories by using this library internally).

The example below uses a pre-built ACI for [etcd](https://github.com/coreos/etcd) (this was built by the [build-aci script](https://github.com/coreos/etcd/blob/master/scripts/build-aci)).

### Downloading an App Container Image (ACI)

rkt uses content addressable storage (CAS) for storing an ACI on disk.
In this example, the image is downloaded and added to the CAS.
Downloading an image before running it is not strictly necessary (if it is not present, rkt will automatically retrieve it), but useful to illustrate how rkt works.

Since rkt verifies signatures by default, you will need to first [trust](signing-and-verification-guide.md#establishing-trust) the [CoreOS public key](https://coreos.com/dist/pubkeys/aci-pubkeys.gpg) used to sign the image, using `rkt trust`:

```
$ sudo rkt trust --prefix=coreos.com/etcd
Prefix: "coreos.com/etcd"
Key: "https://coreos.com/dist/pubkeys/aci-pubkeys.gpg"
GPG key fingerprint is: 8B86 DE38 890D DB72 9186  7B02 5210 BD88 8818 2190
  CoreOS ACI Builder <release@coreos.com>
Are you sure you want to trust this key (yes/no)? yes
Trusting "https://coreos.com/dist/pubkeys/aci-pubkeys.gpg" for prefix "coreos.com/etcd".
Added key for prefix "coreos.com/etcd" at "/etc/rkt/trustedkeys/prefix.d/coreos.com/etcd/8b86de38890ddb7291867b025210bd8888182190"
```

For more information, see the [detailed, step-by-step guide for the signing procedure](signing-and-verification-guide.md).

Now that we've trusted the CoreOS public key, we can fetch the ACI using `rkt fetch`:

```
$ sudo rkt fetch coreos.com/etcd:v2.0.4
rkt: searching for app image coreos.com/etcd:v2.0.4
rkt: fetching image from https://github.com/coreos/etcd/releases/download/v2.0.4/etcd-v2.0.4-linux-amd64.aci
Downloading aci: [==========================================   ] 3.47 MB/3.7 MB
Downloading signature from https://github.com/coreos/etcd/releases/download/v2.0.0/etcd-v2.0.4-linux-amd64.aci.asc
rkt: signature verified:
  CoreOS ACI Builder <release@coreos.com>
sha512-1eba37d9b344b33d272181e176da111e
```

Sometimes you will want to download an image from a private repository.
This usually involves passing usernames and passwords or other kinds of credentials to the server.
rkt currently supports authentication via configuration files.
You can find configuration file format description (with examples!) in [configuration documentation](configuration.md).

For the curious, we can see the files written to disk in rkt's CAS:

```
$ find /var/lib/rkt/cas/blob/
/var/lib/rkt/cas/blob/
/var/lib/rkt/cas/blob/sha512
/var/lib/rkt/cas/blob/sha512/1e
/var/lib/rkt/cas/blob/sha512/1e/sha512-1eba37d9b344b33d272181e176da111ef2fdd4958b88ba4071e56db9ac07cf62
```

Per the [App Container Specification](https://github.com/appc/spec/blob/master/spec/aci.md#image-archives), the SHA-512 hash is of the tarball and can be reproduced with other tools:

```
$ wget https://github.com/coreos/etcd/releases/download/v2.0.4/etcd-v2.0.4-linux-amd64.aci
...
$ gzip -dc etcd-v2.0.4-linux-amd64.aci > etcd-v2.0.4-linux-amd64.tar
$ sha512sum etcd-v2.0.4-linux-amd64.tar
1eba37d9b344b33d272181e176da111ef2fdd4958b88ba4071e56db9ac07cf62cce3daaee03ebd92dfbb596fe7879938374c671ae768cd927bab7b16c5e432e8  etcd-v2.0.4-linux-amd64.tar
```

### Launching an ACI

After it has been retrieved and stored locally, an ACI can be run by pointing `rkt run` at either the original image reference (in this case, "coreos.com/etcd:v2.0.4"), the full URL of the ACI, or the ACI hash.
Hence, the following three examples are equivalent:

```
# Example of running via ACI name:version
$ sudo rkt run coreos.com/etcd:v2.0.4
...
Press ^] three times to kill container
```

```
# Example of running via ACI hash
$ sudo rkt run sha512-1eba37d9b344b33d272181e176da111e
...
Press ^] three times to kill container
```

```
# Example of running via ACI URL
$ sudo rkt run https://github.com/coreos/etcd/releases/download/v2.0.4/etcd-v2.0.4-linux-amd64.aci
...
Press ^] three times to kill container
```

In the latter case, `rkt` will do the appropriate ETag checking on the URL to make sure it has the most up to date version of the image.

Note that the escape character ```^]``` is generated by ```Ctrl-]``` on a US keyboard.
The required key combination will differ on other keyboard layouts.
For example, the Swedish keyboard layout uses ```Ctrl-å``` on OS X and ```Ctrl-^``` on Windows to generate the ```^]``` escape character.
