# CNT - container build tool

[![GoDoc](https://godoc.org/blablacar/cnt?status.png)](https://godoc.org/github.com/blablacar/cnt) [![Build Status](https://travis-ci.org/blablacar/cnt.svg?branch=master)](https://travis-ci.org/blablacar/cnt)

**CNT** is a command line utility designed to build and to configure at runtime App Containers Images ([ACI](https://github.com/appc/spec/blob/master/spec/aci.md)) and App Container Pods ([POD](https://github.com/appc/spec/blob/master/spec/pods.md)) based on convention over configuration.

CNT allows you to build generic container images for a service and to configure them at runtime. Therefore you can use the same image for different environments, clusters, or nodes by overriding the appropriate attributes when launching the container.


_cnt is actively used at blablacar to build more than an hundred different aci and pod to [run all platforms](http://blablatech.com/blog/why-and-how-blablacar-went-full-containers)._

## Build the ACI once, configure your app at runtime.

CNT provides various resources to build and configure an ACI :

  - scripts at runlevels (build, prestart...)
  - templates and attributes
  - static files
  - images from (base filesystem to start from)
  - images dependencies


**Scripts** are executed at the image build, before your container is started and more. See [runlevels](#runlevels) for more information.

**Templates** and **attributes** are the way CNT deals with environment-specific configurations. **Templates** are stored in the image and resolved at runtime ; **attributes** are inherited from different contexts (aci -> pod -> environement). 

**Static files** are copied to same path in the container.

**Images from** is the base filesystem to start building from.
  
**Image dependencies** are used as defined in [APPC spec](https://github.com/appc/spec/blob/master/spec/aci.md#dependency-matching).



![demo](https://raw.githubusercontent.com/blablacar/cnt/gh-pages/aci-dummy.gif)

## Comparison with alternatives

### CNT vs Dockerfiles

A Dockerfile is purely configuration, describing the steps to build the container.
It does not provide scripts levels, ending with very long bash scripting for the run option in the dockerfile.
It does not handle configuration, nor build time nor at runtime. So users usually use sed in the bash script to replace parts of configuration.

### CNT vs acbuild

acbuild is a command line tools to build ACIs. It is more flexible than Dockerfiles as it can be wrapped by other tools such as Makefiles but like Dockerfiles it doesn't provide a standard way of configuring the images.


## Commands

```bash
$ cnt init          # init a sample project
$ cnt build         # build the image
$ cnt clean         # clean the build
$ cnt install       # store target image to rkt local store
$ cnt push          # push target image to remote storage
$ cnt test          # test the final image
```

## Cnt configuration file

CNT global conf is a yaml file located at `~/.config/cnt/config.yml`. Home is the home of starting user (The caller user if running with sudo)
It is used to indicate the target work directory where CNT will create the ACI and the push endpoint informations. Both are optional.

content :
```yml
targetWorkDir: /tmp/target          # if you want to use another directory for all builds
push:
  type: maven
  url: https://localhost/nexus
  username: admin
  password: admin
```

# Building an ACI

## Initializing a new project

Run the following commands to initialize a new project :

```bash
$ mkdir aci-myapp
$ cd aci-myapp
$ cnt init
```

It will generate the following file tree :

```text
.
|-- attributes
|   `-- attributes.yml                 # Attributes file for confd
|-- cnt-manifest.yml                   # Manifest
|-- confd
|   |-- conf.d
|   |   `-- templated.toml             # Confd template resource config
|   `-- templates
|       `-- templated.tmpl             # Confd source template
|-- files
|   `-- dummy                          # Files to be copied in the target rootfs
|-- runlevels
|   |-- build
|   |   `-- 10.install.sh              # Scripts to be run when building
|   |-- build-late
|   |   `-- 10.setup.sh                # Scripts to be run when building after the copy of files
|   |-- build-setup
|   |   `-- 10.setup.sh                # Scripts to be run directly on source host before building
|   |-- inherit-build-early
|   |   `-- 10.inherit-build-early.sh  # Scripts stored in ACI and used when building FROM this image
|   |-- inherit-build-late
|   |   `-- 10.inherit-build-early.sh  # Scripts stored in ACI and used when building FROM this image
|   |-- prestart-early
|   |   `-- 10.prestart-early.sh       # Scripts to be run when starting ACI before confd templating
|   `-- prestart-late
|       `-- 10.prestart-late.sh        # Scripts to be run when starting ACI after confd templating
`-- tests
    |-- dummy.bats                     # Bats tests for this ACI
    `-- wait.sh
```

This project is already valid which means that you can build it and it will result in a runnable ACI. (CNT always adds busybox to the ACI). But you probably want to customize it at this point.

## Customizing

### The manifest

The CNT manifest looks like a light ACI manifest. CNT will take this manifest and convert it to the format defined in the APPC spec.

Example of a cnt-manifest.yml :
```yaml
from: example.com/base:1
name: example.com/myapp:0.1
aci:
  app:
    exec:
      - /bin/myapp
      - -c
      - /etc/myapp/myapp.cfg
    mountPoints:
      - name: myapp-data
        path: /var/lib/myapp
        readOnly: false
```

The **from** points to an ACI that will be taken as the base for the ACI we are building. The rootfs of this ACI will be copied before executing the build scripts. Typically you can use there an ACI of your favorite distro.
The **name**, well, is the name of the ACI you are building.
Under the **aci** key, you can add every key that is defined in the APPC spec such as :
  - **exec** which contains the absolute path to the executable your want to run at the start of the ACI and its args.
  - **mountPoints** even though you can do it on the command line with recent versions of RKT.
  - **isolators**...

### The build scripts

The scripts in ```runlevels/build``` dir are executed during the build to install in the ACI everything you need. For instance if you base ACI in the FROM field of the manifest is a debootstab from Debian, a build script could look like :

```bash
#!/bin/bash
apt-get update
apt-get install -y myapp
```

### The templates

You can create templates in your ACI. For that we use [confd](https://github.com/kelseyhightower/confd), so you should at the documentation on there if you're not familiar with it.
For our example :

confd/conf.d/myapp.cfg.toml
```ini
[template]
src = "myapp.cfg.tmpl"
dest = "/etc/myapp/myapp.cfg"
keys = ["/data"]
```

confd/templates/myapp.cfg.tmpl
```
{{$data := json (getv "/data")}}
{{ if $data.myapp.setting1 }}
setting1: {{$data.myapp.setting1}}
{{ end }}
```

Note that the first line is compulsory as this is the way to get all the attributes in the $data variable.

### The attributes

All the YAML files in the directory **attributes** are read by CNT. The first node of the YAML has to be "default" as it can be overridden in a POD or with a json in the env variable CONFD_OVERRIDE in the cmd line.

attributes/myapp.yml
```yaml
---
default:
  myapp:
    setting1: value1
    setting2: 42
```


### The prestart

CNT uses the "pre-start" eventHandler of the ACI to customize the ACI rootfs before the run depending on the instance or the environment.
It resolves at that time the templates so it has all the context needed to do that.
You can also run custom scripts before (prestart-early) or after (prestart-late) this template resolution. This is useful if you want to initialize a mountpoint with some data before running your app for instance.

runlevels/prestart-late/init.sh
```bash
#!/bin/bash
set -e
/usr/bin/myapp-init
```

Building a POD
=============

#Standard FileTree for POD

```bash
├── aci-elasticsearch               # Directory that match the pod app shortname (or name)
│   ├── attributes
│   │   └── attributes.yml          # Attributes file for confd in this ACI
│   ├── files                       # Files to be inserted into this ACI
│   ...  
├── cnt-pod-manifest.yml            # Pod Manifest

```

####
File templating will be resolved on container start using [confd](https://github.com/kelseyhightower/confd) with the backend env.


# caveats

- [rkt](https://github.com/coreos/rkt) in path
- systemd-nspawn to launch 'build runlevels' scripts
- being root is required to construct the filesystem


# Inspiration

Tool to build [APPC](https://github.com/appc/spec) ACI and POD in a mixup of Chef, Dockerfile and Packer logic.
