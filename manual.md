# ERU Cli Manual

# TOC

- [Some Terms](#some-terms)
- [Global Options](#global-options)
- [Sub Commands](#sub-commands)
  - [Core Sub Commands](#core-sub-commands)
    - [info](#info)
  - [Image Sub Commands](#image-sub-commands)
    - [build](#build)
    - [cache](#cache)
    - [remove](#remove)
  - [Lambda Sub Commands](#lambda-sub-commands)
  - [Node Sub Commands](#node-sub-commands)
    - [add](#add)
    - [remove](#remove-1)
    - [get](#get)
    - [set](#set)
    - [workloads, containers](#workloads--containers)
    - [up](#up)
    - [down](#down)
    - [set-status](#set-status)
    - [watch-status](#watch-status)
    - [resource](#resource)
  - [Pod Sub Commands](#pod-sub-commands)
    - [add](#add-1)
    - [list](#list)
    - [remove](#remove-2)
    - [resource](#resource-1)
    - [capacity](#capacity)
    - [nodes](#nodes)
    - [networks](#networks)
  - [Status Sub Commands](#status-sub-commands)
  - [Workload / Container Sub Commands](#workload---container-sub-commands)
    - [get](#get-1)
    - [logs](#logs)
    - [get-status](#get-status)
    - [set-status](#set-status-1)
    - [list](#list-1)
    - [stop](#stop)
    - [start](#start)
    - [restart](#restart)
    - [remove](#remove-3)
    - [copy](#copy)
    - [send](#send)
    - [dissociate](#dissociate)
    - [realloc](#realloc)
    - [exec](#exec)
    - [deploy](#deploy)
    - [replace](#replace)

## Some Terms

Sometimes it confuses users when an option is given as a flag, or a bool value. So we make some agreements here:

- When you see `flag`, it means you only need to give this option for a `true` value, or don't give for a `false` value.

  - Assume `eru-cli` supports a flag option `--flag-option`, then:
  - `eru-cli --flag-option` means `--flag-option` is `true`.
  - `eru-cli` means `--flag-option` is `false`.

- When you see `use this option`, it means you give a value to this option, or give this option when it's a flag.

  - Assume `eru-cli` supports `--use-this-option`, then:
  - `eru-cli --use-this-option tonic` means the value of `--use-this-option` is `tonic`.
  - `eru-cli --use-this-option muroq` means the value of `--use-this-option` is `muroq`.
  - `eru-cli` means the value of `--use-this-option` is the default value, if it has a default value.
  - For flag case, refer to the former one.

- When you see `define this option`, it means you set a value for this option, either via this cli option, or via environment variables.

Also pay attention to the order of components, it should be:

```
eru-cli [global options] command [command options] [arguments...]
```

`command` can also be a group of `sub commands`.

## Global Options

Global options are used to define the global settings for eru-cli command. Once defined, options are affected to all sub commands.

We currently have these options:

- `--debug`, `-d`

  - This is a flag.
  - If given, enables debug mode, debug mode prints more detailed information.
  - Default value is empty, which means you don't use this option, this will disable debug mode.

- `--eru`, `-e`

  - This option defines which eru-core you want to use.
  - Default value is `localhost:5001`
  - You can also set environment variable `ERU` to define this option.
  - Note: eru-cli will try to connect to this eru-core instance, but if there're multiple eru-core instances, service discovery will decide which one is the one to use, it may not be the one you defined in this option.

- `--username`, `-u`

  - This option defines the username of the target eru-core (as defined in `--eru` / `-e`).
  - If your eru-core instance doesn't require a username, don't use this option.
  - You can also set environment variable `ERU_USERNAME` to define this option.

- `--password`, `-p`

  - This option defines the password of the target eru-core (as defined in `--eru`/ `-e`).
  - If your eru-core instance doesn't require a password, dont' use this option.
  - You can also set environment variable `ERU_PASSWORD` to define this option.

- `--output`, `-o`

  - This option defines the output format of eru-cli.
  - Possible values are `json`, `yaml`, or don't use this option.
  - Format `json` will print result in JSON format.
  - Format `yaml` will print result in Yaml format.
  - The default value is empty, which means you don't use this option, then the result will be printed as a table.
  - Table format will only print some user friendly information, for details, `json` / `yaml` format is suggested.
  - You can also set environment variable `ERU_OUTPUT_FORMAT` to define this option.

- `--help`, `-h`

  - When this option is used, eru-cli will print help message and exit.
  - This option is also supported by all the sub commands. Hence when introducing sub commands, we will skip this `--help` / `-h` option.

- `--version`, `-v`
  - When this option is used, eru-cli will print version and exit.

## Sub Commands

### Core Sub Commands

Core sub commands are started with `core` command. The format should be `eru-cli core [sub command] [command options] [arguments...]`

These sub commands are supported:

- `info`

#### info

This command doesn't require any arguments. It shows the information of eru-core.

An example is:

```
$ eru-cli core info
┌────────────────┬──────────────────────────────────────────┐
│ NAME           │ DESCRIPTION                              │
├────────────────┼──────────────────────────────────────────┤
│ Version        │ v21.05.24                                │
│ Git hash       │ 7f382e257d1ff55ed9308782c13a948f18dac07b │
│ Built          │ 2021-05-24T07:12:18                      │
│ Golang version │ go1.16.4                                 │
│ OS/Arch        │ linux/amd64                              │
│ Identifier     │ 749e1a1cc00fa5d2cd9dbec61ae0856c6431856b │
└────────────────┴──────────────────────────────────────────┘
```

An example with JSON format is:

```
$ eru-cli --output json core info
{
  "version": "v21.05.24",
  "revison": "7f382e257d1ff55ed9308782c13a948f18dac07b",
  "build_at": "2021-05-24T07:12:18",
  "golang_version": "go1.16.4",
  "os_arch": "linux/amd64",
  "identifier": "749e1a1cc00fa5d2cd9dbec61ae0856c6431856b"
}
```

### Image Sub Commands

Image sub commands are started with `image` command. The format should be `eru-cli image [sub command] [command options] [arguments...]`.

The image sub commands are only supported by docker engine, systemd and virtual machine are not supported.

These sub commands are supported:

- `build`
- `cache`
- `remove`

#### build

This command will build an image from a specification.

The format is `eru-cli image build [command options] <buildargument>`.

Command options are:

- `--name`

  - Defines the name of the image.
  - MUST be given

- `--tag`

  - Defines the tag of the image.
  - MUST be given

- `--raw`

  - This is a flag.
  - If this option is defined, `eru-cli image build --name ${name} --tag ${tag} path/to/content` acts like `docker build -t ${name}:${tag} path/to/content`.
  - You need to make sure the content in `path/to/content` is exactly like when you are using `docker build`, for example, there is a default Dockerfile.
  - This flag conflicts with `--exist`.

- `--exist`

  - This is a flag.
  - If this option is defined, `eru-cli image build --name ${name} --tag ${tag} ${containerID}` acts like `docker commit ${containerID} ${name}:${tag}`.
  - This flag conflicts with `--raw`.
  - This is useful when you want to create an image from existing ERU container.

- `--user` and `--uid`

  - Defines the default user and uid of the image, equivalent to `USER` in Dockerfile.
  - Default value of user is `root`.
  - Default value of uid is `1`.

- `--stop-signal`
  - Defines the signal sent to the workload for docker.
  - Equivalent to `STOPSIGNAL` in Dockerfile.
  - You can set the signal number or the signal name, `--stop-signal SIGKILL` and `--stop-signal 9` both work.

Command arguments are:

- `<buildargument>`

  - This argument has multiple meanings.
  - If `--exist` is defined, this argument refers to a working workload ID, note that workload type can only be docker container currently.
  - If `--raw` is defined, this argument refers to a path that contains all the files for `docker build` command.
  - If neither `--exist` nor `--raw` is defined, this argument refers to either a URL contains the specification file, or a path to the specification file.
  - An example of specification file is:

    ```
    stages:
      - test
      - build

    builds:
      test:
        base: tonic/ubuntu:phistage
        commands:
          - echo tested
      build:
        base: tonic/ubuntu:phistage
        repo: git@github.com:tonicbupt/phistage.git
        version: "HEAD"
        commands:
          - echo built
          - echo test > testfile
        cache:
          "testfile": "/usr/local/testfile
    ```

#### cache

This command will cache an image to all the nodes or pods specified.

The format is `eru-cli image cache [command options] <cacheargument...>`.

The `<cacheargument...>` refers to the images to pull, can be given multiple times, like `eru-cli image cache --podname pod img1 img2 img3`.

Command options are:

- `--nodename`

  - This option can be defined multiple times, like `--nodename n1 --nodename n2`
  - Defines the name of the node.
  - If you only want to pre pull an image on some nodes, use this option.

- `--podname`

  - Defines the name of the pod.
  - If you want to pre pull an image on all nodes of a pod, use this option.

- `--concurrent`

  - Defines how many workers to pull image concurrently.
  - Default value is `10`.

#### remove

This command will remove an image from all the nodes or pods specified.

The format is `eru-cli image remove [command options] <removeargument...>`.

The `<removeargument...>` refers to the images to remove, can be given multiple times, like `eru-cli image remove --podname pod img1 img2 img3`.

Command options are:

- `--nodename`

  - This option can be defined multiple times, like `--nodename n1 --nodename n2`
  - Defines the name of the node.
  - If you only want to remove an image from some nodes, use this option.

- `--podname`

  - Defines the name of the pod.
  - If you want to remove an image from all nodes of a pod, use this option.

- `--concurrent`

  - Defines how many workers to remove image concurrently.
  - Default value is `10`.

- `--prune`

  - This is a flag.
  - If this flag is defined, will also remove unused images on the nodes.

### Lambda Sub Commands

Lambda sub commands are started with `lambda` command, and only contains one command: `eru-cli lambda`. The format should be `eru-cli lambda [command options] <commands...>`

`<commands...>` refers to the command to execute, for example, `sh -c "ls ~ && echo done"`.

Command options are:

- `--name`

  - This option defines the name of this lambda.
  - MUST be given.

- `--network`

  - This option defines the network for this lambda workload.
  - The format can be a single string representing the network name, like `--network calico` or `--network host`,
  - Or can be a string contains a ":" to specify the network name and IPv4 IP address, like `--network calico:10.233.12.41`.
  - If this value is not defined, an empty string will be used, the network will depend on the settings of eru-core engine.

- `--pod`, `--node`

  - This option defines the pod or the node for this lambda to run.
  - Note that if pod and node are both defined, will take node for priority. For example if you define `--pod podname --node nodename` and `nodename` doesn't belong to `podname`, `nodename` will still be used while `podname` is ignored.

- `--env`

  - This option defines the environment variable for this lambda runtime.
  - This option can be defined multiple times, like `--env A=1 --env B=2 --env C=3`.

- `--working-dir`, `--working_dir`

  - This option defines the working directory for this lambda runtime.
  - The default value is `/`.
  - `--working_dir` is only a compat, it looks so ugly...

- `--image`

  - This option defines the image for this lambda runtime.
  - The default value is `alpine:latest`.

- `--count`

  - This option defines how many workloads to run this lambda concurrently.
  - The default value is `1`, in most cases, you only need one lambda runtime.

- `--stdin`

  - This is a flag.
  - If this flag is defined, stdin for this lambda runtime is open, you can interact with this lambda workload.

- `--user`

  - This option defines the user for this lambda runtime.
  - The default value is `root`.

- `--file`

  - This option defines the files for lambda runtime.
  - This option can be defined multiple times, like `--file localpath1:remotepath1 --file path/to/local:path/to/remote`.
  - Files will be sent to lambda runtime before the execution of commands.

- `--async`, `--async-timeout`

  - `--async` is a flag, if it's defined, eru-cli will not wait for the termination of lambda runtime.
  - If `--async` is defined, you can use `--async-timeout` to define the timeout of this lambda runtime, default value is `30`, which means 30 seconds.
  - The lambda runtime will terminate after `--async-timeout` seconds.

- `--privileged`, `-p`
  - This is a flag.
  - If this option is defined, the extended privileges are given to this lambda runtime.
  - Default value is `false`, which means if you don't use this option, no more extended privileges are given.

An example is:

```
root@tonic-eru-test:~# eru-cli lambda --name test-lambda --pod muroq --image tonic/ubuntu:phistage sh -c "ls -alh ~ && echo done && cat ~/.profile"
INFO[2021-06-17 14:50:53] [WorkloadID] 6272652822eff3ec3892ae6cb6c263b465e6802498c77c1cd38b168f4b425787
total 16K
drwx------ 2 root root 4.0K Apr 29 07:08 .
drwxr-xr-x 1 root root 4.0K Jun 17 06:50 ..
-rw-r--r-- 1 root root 3.1K Oct 22  2015 .bashrc
-rw-r--r-- 1 root root  148 Aug 17  2015 .profile
done
# ~/.profile: executed by Bourne-compatible login shells.

if [ "$BASH" ]; then
  if [ -f ~/.bashrc ]; then
    . ~/.bashrc
  fi
fi

mesg n || true
```

### Node Sub Commands

Node sub commands are started with `node` command. The format should be `eru-cli node [sub command] [command options] [arguments...]`.

These sub commands are supported:

- `add`
- `remove`
- `get`
- `set`
- `workloads`, `containers`
- `up`
- `down`
- `set-status`
- `watch-status`
- `resource`

#### add

This command will add a node to ERU system.

The format is `eru-cli node add [command options] <podname>`.

`<podname>` refers to the name of the pod, this value must be defined to specify which pod this node belongs to.

Command options are:

- `--nodename`

  - Defines the name of the node.
  - The default value is the hostname of the server (`$HOSTNAME` in environment variables).
  - You can use a customized name to over write this.

- `--endpoint`

  - Defines the endpoint of the node engine.
  - For docker engine, this should be like `tcp://10.233.10.43:2376`.
  - For systemd engine, this should be like `systemd://10.233.10.44:22`.
  - For virtual machines controlled by yavirtd, this should be like `virt-grpc://10.233.10.45:9697`.

- `--ca`, `--cert`, `--key`

  - Defines the certificates for docker server.
  - If you are not using docker engine, or your docker engine doesn't require certificates, don't use these options.

- `--cpu`

  - Defines how many cores this node has.
  - Usually this can be detected automatically, don't set this manually if you have no idea what you are doing.

- `--share`

  - Defines the share of cores.
  - Default value is `100`, usually default value is good, no need to change it.

- `--memory`

  - Defines how much memory this node has.
  - Usually this can be detected automatically, don't set this manually if you have no idea what you are doing.
  - Units are supported, you can use `--memory 128G` or `--memory 1024M` to set the value.

- `--label`

  - Defines the labels to this node.
  - This option can be defined multiple times, like `--label rack=rack1 --label cluster=cluster3`.

- `--volumes`

  - Defines the volumes of this node.
  - This option can be defined multiple times, like `--volumes /data0:500G --volumes /data1:8T`.

An example is:

```
root@tonic-eru-test:~# eru-cli node add --nodename test7 --endpoint tcp://127.0.0.1:2376 muroq
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test7 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘
```

#### remove

This command will remove a node from ERU system.

The format is `eru-cli node remove [command options] <nodename>`.

`<nodename>` refers to the name of the node to remove.

An example is:

```
root@tonic-eru-test:~# eru-cli node remove test7
INFO[2021-06-17 15:18:14] [RemoveNode] success
```

#### get

This command will get a node from ERU system, and print its information.

The format is `eru-cli node get [command options] <nodename>`.

`<nodename>` refers to the name of the node.

An example is:

```
root@tonic-eru-test:~# eru-cli node get test0
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test0 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘
```

An example of JSON output format is:

```
root@tonic-eru-test:~# eru-cli --output json node get test0
[
  {
    "name": "test0",
    "endpoint": "tcp://127.0.0.1:2376",
    "podname": "muroq",
    "cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "memory": 13026197504,
    "available": true,
    "init_memory": 13026197504,
    "init_cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "info": "{\"ID\":\"CZ4T:7ED7:LKTH:DRUX:MWK3:5NHS:67TT:SYPE:O2WD:TN7T:YSL5:LQG7\",\"NCPU\":4,\"MemTotal\":16282746880,\"StorageTotal\":0}"
  }
]
```

#### set

This command will modify some attributes of a node.

The format is `eru-cli node set [command options] <nodename>`.

`<nodename>` refers to the name of the node to modify.

Command options are:

- `--delta`

  - This is a flag.
  - If this flag is defined, `--memory`, `--storage`, `--volume` are using delta mode, the values of these options are relative to the current values.
  - If this flag is not defined, the values of `--memory`, `--storage`, `--volume` are absolute values, regardless of the current values.
  - This flag is very **IMPORTANT** as it can change the behavior of `eru-cli node set`.

- `--mark-workloads-down`

  - This is a flag.
  - If this flag is defined, all workloads of this node will be marked as unavailable.
  - It's only a mark, the workloads may still work well.

- `--memory`

  - Defines memory of this node.
  - Units are supported, you can use `--memory 128G` or `--memory 1024M` to set the value.
  - For example, if the current memory is `128G`:

    - `--memory 128G` will make it still `128G`;
    - `--memory 256G` will make it `256G`;
    - `--memory 0` will set it to `0`;
    - `--memory 128G --delta` will make it `256G`;
    - `--memory 0 --delta` will make it still `128G`;
    - `--memory -128G --delta` will set it to `0`;

- `--storage`

  - Defines storage of this node.
  - Units are supported, you can use `--storage 512G` or `--storage 8T` to set the value.
  - For example, if the current storage is `1T`:

    - `--storage 1T` will make it still `1T`;
    - `--storage 2T` will make it `2T`;
    - `--storage 0` will set it to `0`;
    - `--storage 1T --delta` will make it `2T`;
    - `--storage 0 --delta` will make it still `1T`;
    - `--storage -1T --delta` will set it to `0`;

- `--volumes`

  - Defines the volumes of this node.
  - The format is `PATH0:SIZE0,PATH1:SIZE1,...`, for example `--volume /data0:100G,/data1:200G`.
  - Units are supported, you can use `--volume /data0:512G` or `--volume /data1:8T` to set the value.
  - For example, if the current volume is `/data0:500G` and `/data1:300G`:

    - `--volume /data0:500G,/data1:300G,/data2:100G` will make it `/data0:500G`, `/data1:300G`, `/data2:100G`;
    - `--volume /data0:800G,/data1:200G,/data2:100G` will make it `/data0:800G`, `/data1:200G`, `/data2:100G`;
    - `--volume /data0:0,/data1:0,/data2:0` will make it `/data0:0`, `/data1:0`, `/data2:0`;
    - `--volume /data0:500G,/data1:300G,/data2:100G --delta` will make it `/data0:1000G`, `/data1:600G`, `/data2:100G`;
    - `--volume /data0:0,/data1:0,/data2:0 --delta` will make it `/data0:500G`, `/data1:300G`, `/data2:0`;
    - `--volume /data0:-500G,/data1:-300G,/data2:100G --delta` will make it `/data0:0`, `/data1:0`, `/data2:100G`;

- `--cpu`

  - Defines the cores and shares of this node.
  - The format is `CPUNO0:CPUSHARE0,CPUNO1:CPUSHARE1,CPUNO2:CPUSHARE2,...`, for example `--cpu 0:100,1:50,2:50,3:100`

- `--label`

  - Defines the labels to this node.
  - This option can be defined multiple times, like `--label rack=rack1 --label cluster=cluster3`.

An example is:

```
root@tonic-eru-test:~# eru-cli node get test0
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test0 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘

root@tonic-eru-test:~# eru-cli node set --storage 100G test0
INFO[2021-06-17 15:52:47] [SetNode] set node test0 success

root@tonic-eru-test:~# eru-cli node get test0
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬────────────────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE                │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼────────────────────────┤
│ test0 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 107374182400 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴────────────────────────┘

root@tonic-eru-test:~# eru-cli node set --storage -100G --delta test0
INFO[2021-06-17 15:52:59] [SetNode] set node test0 success

root@tonic-eru-test:~# eru-cli node get test0
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test0 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘
```

#### workloads, containers

This command will list workloads of a node.

The format is `eru-cli node workloads [command options] <nodename>`.

`<nodename>` refers to the name of the node.

`containers` is a legacy replacement of `workloads`, currently it's still available to use.

Command options are:

- `--label`

  - Defines the labels of the workloads.
  - This option can be defined multiple times, like `--label runtime=ubuntu --label ERU=1`.

An example is:

```
root@tonic-eru-test:~# eru-cli node workloads test0
┌──────────────────────────────────────────────────────────────────┬───────────────────────────┬──────────────────────────┬──────────────────────┐
│ NAME/ID/POD/NODE                                                 │ STATUS                    │ VOLUME                   │ NETWORKS             │
├──────────────────────────────────────────────────────────────────┼───────────────────────────┼──────────────────────────┼──────────────────────┤
│ test_http_etdWiu                                                 │ CPUQuotaRequest: 1.000000 │ VolumesRequest: []       │ host: 127.0.0.1:8000 │
│ 9ef5a7c414c74aa39187fb669a202c905b8ecb30915b5bc9d6199e583927beb5 │ CPUQuotaLimit: 1.000000   │ VolumesLimit: []         │                      │
│ muroq                                                            │ CPUMap: map[]             │ VolumePlanRequest: map[] │                      │
│ test0                                                            │ MemoryRequest: 536870912  │ VolumePlanLimit: map[]   │                      │
│                                                                  │ MemoryLimit: 536870912    │                          │                      │
│                                                                  │ StorageRequest: 0         │                          │                      │
│                                                                  │ StorageLimit: 0           │                          │                      │
│                                                                  │ Privileged: false         │                          │                      │
└──────────────────────────────────────────────────────────────────┴───────────────────────────┴──────────────────────────┴──────────────────────┘
```

#### up

This command will label a node as UP, only nodes in state UP can be used to deploy workloads.

The format is `eru-cli node up [command options] <nodename>`.

`<nodename>` refers to the name of the node.

An example is:

```
root@tonic-eru-test:~# eru-cli --output json node get test0
[
  {
    "name": "test0",
    "endpoint": "tcp://127.0.0.1:2376",
    "podname": "muroq",
    "cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "memory": 13026197504,
    "init_memory": 13026197504,
    "init_cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "info": "{\"ID\":\"CZ4T:7ED7:LKTH:DRUX:MWK3:5NHS:67TT:SYPE:O2WD:TN7T:YSL5:LQG7\",\"NCPU\":4,\"MemTotal\":16282746880,\"StorageTotal\":0}"
  }
]

root@tonic-eru-test:~# eru-cli node up test0
INFO[2021-06-17 16:03:23] [SetNode] node test0 up

root@tonic-eru-test:~# eru-cli --output json node get test0
[
  {
    "name": "test0",
    "endpoint": "tcp://127.0.0.1:2376",
    "podname": "muroq",
    "cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "memory": 13026197504,
    "available": true,
    "init_memory": 13026197504,
    "init_cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "info": "{\"ID\":\"CZ4T:7ED7:LKTH:DRUX:MWK3:5NHS:67TT:SYPE:O2WD:TN7T:YSL5:LQG7\",\"NCPU\":4,\"MemTotal\":16282746880,\"StorageTotal\":0}"
  }
]
```

Note the `"available"` key in the JSON formatted result.

#### down

This command will label a node as DOWN, the nodes in state DOWN will not be chosen to deploy workloads.

The format is `eru-cli node down [command options] <nodename>`.

`<nodename>` refers to the name of the node.

Command options are:

- `--check`

  - This is a flag.
  - If this flag is defined, `eru-cli node down` will check if the node is really down before set node to DOWN, if the node is healthy, will not set it to DOWN.

- `--check-timeout`

  - Defines the timeout for checking the node.

An example is:

```
root@tonic-eru-test:~# eru-cli --output json node get test0
[
  {
    "name": "test0",
    "endpoint": "tcp://127.0.0.1:2376",
    "podname": "muroq",
    "cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "memory": 13026197504,
    "available": true,
    "init_memory": 13026197504,
    "init_cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "info": "{\"ID\":\"CZ4T:7ED7:LKTH:DRUX:MWK3:5NHS:67TT:SYPE:O2WD:TN7T:YSL5:LQG7\",\"NCPU\":4,\"MemTotal\":16282746880,\"StorageTotal\":0}"
  }
]

root@tonic-eru-test:~# eru-cli node down --check test0
WARN[2021-06-17 16:12:05] [SetNode] node is not down

root@tonic-eru-test:~# eru-cli node down test0
INFO[2021-06-17 16:12:11] [SetNode] node test0 down

root@tonic-eru-test:~# eru-cli --output json node get test0
[
  {
    "name": "test0",
    "endpoint": "tcp://127.0.0.1:2376",
    "podname": "muroq",
    "cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "memory": 13026197504,
    "init_memory": 13026197504,
    "init_cpu": {
      "0": 100,
      "1": 100,
      "2": 100,
      "3": 100
    },
    "info": "{\"ID\":\"CZ4T:7ED7:LKTH:DRUX:MWK3:5NHS:67TT:SYPE:O2WD:TN7T:YSL5:LQG7\",\"NCPU\":4,\"MemTotal\":16282746880,\"StorageTotal\":0}"
  }
]
```

Note the `"available"` key in the JSON formatted result.

This command can help to remove some nodes from scheduler temporarily.

#### set-status

This command will set the status of a node. It is used for heartbeat, in a side way. Usually used for debugging.

The format is `eru-cli node set-status [command options] <nodename>`.

`<nodename>` refers to the name of the node.

Command options are:

- `--ttl`

  - Defines the TTL of the status data.
  - Default value is `180`, which means the status data will be alive for 180 seconds, it not refreshed.

- `--interval`

  - Defines the interval of the set action.
  - Default value is `0`, which means don't set the status periodically.
  - If this option is defined, like `--interval 30`, eru-cli will set the status every 30 seconds.

#### watch-status

This command will build a connection to eru-core and wait for node events.

The format is `eru-cli node watch-status`.

An example is:

```
root@tonic-eru-test:~# eru-cli node watch-status
INFO[2021-06-17 16:23:36] [WatchNodeStatus] Node test2 on pod muroq, alive: true
INFO[2021-06-17 16:23:39] [WatchNodeStatus] Node test1 on pod muroq, alive: true
INFO[2021-06-17 16:23:40] [WatchNodeStatus] Node test0 on pod muroq, alive: true
INFO[2021-06-17 16:23:51] [WatchNodeStatus] Node test2 on pod muroq, alive: true
```

#### resource

This command will show resource of a node, or try to fix the resource inconsistency.

The format is `eru-cli node resource [command options] <nodename>`.

`<nodename>` refers to the name of the node.

Command options are:

- `--fix`

  - This is a flag.
  - If this flag is defined, eru-cli will try to fix the resource inconsistency as well as print the current resource.

An example is:

```
root@tonic-eru-test:~# eru-cli node resource test0
┌───────┬───────┬────────┬─────────┬────────┬───────┐
│ NAME  │ CPU   │ MEMORY │ STORAGE │ VOLUME │ DIFFS │
├───────┼───────┼────────┼─────────┼────────┼───────┤
│ test0 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
└───────┴───────┴────────┴─────────┴────────┴───────┘
```

This command can sometimes help to fix the resource inconsistency of nodes, it's very useful.

### Pod Sub Commands

Pod sub commands are started with `pod` command. The format should be `eru-cli pod [sub command] [command options] [arguments...]`.

These sub commands are supported:

- `add`
- `list`
- `remove`
- `resource`
- `capacity`
- `nodes`
- `networks`

#### add

This command will add a pod to ERU system.

The format is `eru-cli pod add [command options] <podname>`.

`<podname>` refers to the name of the pod.

Command options are:

- `--desc`

  - Defines the description of the pod.
  - If not defined, an empty description will be used.

An example is:

```
root@tonic-eru-test:~# eru-cli pod add --desc test-pod2 muroq2
┌────────┬─────────────┐
│ NAME   │ DESCRIPTION │
├────────┼─────────────┤
│ muroq2 │ test-pod2   │
└────────┴─────────────┘
```

#### list

This command will list all pods in ERU system.

The format is `eru-cli pod list`.

An example is:

```
root@tonic-eru-test:~# eru-cli pod list
┌────────┬─────────────┐
│ NAME   │ DESCRIPTION │
├────────┼─────────────┤
│ muroq  │             │
│ muroq1 │             │
│ muroq2 │ test-pod2   │
└────────┴─────────────┘
```

#### remove

This command will remove a pod, only an empty pod without nodes can be removed.

The format is `eru-cli pod remove <podname>`.

`<podname>` refers to the name of the pod.

An example is:

```
root@tonic-eru-test:~# eru-cli pod remove muroq2
INFO[2021-06-17 16:37:08] [RemovePod] success, name: muroq2

root@tonic-eru-test:~# eru-cli pod list
┌────────┬─────────────┐
│ NAME   │ DESCRIPTION │
├────────┼─────────────┤
│ muroq  │             │
│ muroq1 │             │
└────────┴─────────────┘

root@tonic-eru-test:~# eru-cli pod remove muroq
rpc error: code = Code(1022) desc = pod has nodes: pod muroq still has 3 nodes, delete them first
```

#### resource

This command will show resource of nodes within this pod.

The format is `eru-cli pod resource [command options] <podname>`.

`<podname>` refers to the name of the pod.

Command options are:

- `--filter`

  - Defines the filter to filter resource value.
  - The filtered resource can be `cpu`, `memory`, `storage`, `volume`.
  - The comparator can be `<`, `<=`, `>`, `>=`, `==`.
  - The compare target can be like `40%` or `0.4`.
  - The default value is `all`, which means no filter is used, will show resource of all nodes.

An example is:

```
root@tonic-eru-test:~# eru-cli pod resource muroq
┌───────┬───────┬────────┬─────────┬────────┬───────┐
│ NAME  │ CPU   │ MEMORY │ STORAGE │ VOLUME │ DIFFS │
├───────┼───────┼────────┼─────────┼────────┼───────┤
│ test0 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
│ test1 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
│ test2 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
└───────┴───────┴────────┴─────────┴────────┴───────┘

root@tonic-eru-test:~# eru-cli pod resource --filter cpu>0.5 muroq

root@tonic-eru-test:~# eru-cli pod resource --filter cpu==0 muroq
┌───────┬───────┬────────┬─────────┬────────┬───────┐
│ NAME  │ CPU   │ MEMORY │ STORAGE │ VOLUME │ DIFFS │
├───────┼───────┼────────┼─────────┼────────┼───────┤
│ test0 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
│ test1 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
│ test2 │ 0.00% │ 0.00%  │ 0.00%   │ 0.00%  │       │
└───────┴───────┴────────┴─────────┴────────┴───────┘
```

#### capacity

This command will calculate the capacity of this pod. Capacity means how many workloads can be deployed.

The format is `eru-cli pod capacity [command options] <podname>`.

`<podname>` refers to the name of the pod.

Command options are:

- `--cpu`, `-c`

  - Defines the CPU the target workload requires.
  - If the target workload doesn't require an exclusive CPU, define this option to `0` like `--cpu 0`.

- `--memory`, `-m`, `--mem`

  - Defines the memory the target workload requires.
  - Units are supported, you can use `--memory 128G` or `--memory 1024M` to set the value.

- `--storage`, `-s`

  - Defines the storage the target workload requires.
  - Units are supported, you can use `--storage 500G` or `--storage 1T` to set the value.

- `--cpu-bind`

  - This is a flag.
  - If this flag is defined, the CPU will be bound to workloads.

- `--nodename`, `--node`, `-n`

  - Defines the nodes to join calculation.
  - This option can be defined multiple times, like `--nodename node1 --nodename node2`.

An example is:

```
root@tonic-eru-test:~# eru-cli pod capacity --cpu 0 --memory 1G --storage 0G muroq
Total: 36
┌───────┬──────────┐
│ NODE  │ CAPACITY │
├───────┼──────────┤
│ test2 │ 12       │
│ test1 │ 12       │
│ test0 │ 12       │
└───────┴──────────┘

root@tonic-eru-test:~# eru-cli pod capacity --cpu 0 --memory 2G --storage 0G muroq
Total: 18
┌───────┬──────────┐
│ NODE  │ CAPACITY │
├───────┼──────────┤
│ test0 │ 6        │
│ test2 │ 6        │
│ test1 │ 6        │
└───────┴──────────┘

root@tonic-eru-test:~# eru-cli pod capacity --cpu 0 --memory 4G --storage 0G muroq
Total: 9
┌───────┬──────────┐
│ NODE  │ CAPACITY │
├───────┼──────────┤
│ test2 │ 3        │
│ test1 │ 3        │
│ test0 │ 3        │
└───────┴──────────┘

root@tonic-eru-test:~# eru-cli pod capacity --cpu 0 --memory 4G --storage 0G --nodename test0 muroq
Total: 3
┌───────┬──────────┐
│ NODE  │ CAPACITY │
├───────┼──────────┤
│ test0 │ 3        │
└───────┴──────────┘
```

#### nodes

This command will list all nodes within this pod, or by some filters.

The format is `eru-cli pod nodes [command options] <podname>`.

`<podname>` refers to the name of the pod.

Command options are:

- `--filter`, `-f`

  - Defines the filter for filtering the nodes.
  - The value of this option can be defined as `up`, `down`, or `all`, default value is `all`.
  - `--filter up` will only show those nodes with state UP.
  - `--filter down` will only show those nodes with state DOWN.
  - `--filter all` will show all nodes regardless of the state.

- `--label`

  - Defines the labels to filter.
  - This option can be defined multiple times, like `--label rack=rack1 --label cluster=cluster3`.

An example is:

```
root@tonic-eru-test:~# eru-cli pod nodes --filter up muroq
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test1 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test2 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘

root@tonic-eru-test:~# eru-cli pod nodes --filter down muroq
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test0 │ tcp://127.0.0.1:2376 │ DOWN   │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘

root@tonic-eru-test:~# eru-cli pod nodes --filter all muroq
┌───────┬──────────────────────┬────────┬──────────┬───────────────────────┬─────────────┬─────────────┐
│ NAME  │ ENDPOINT             │ STATUS │ CPU      │ MEMORY                │ VOLUME      │ STORAGE     │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test0 │ tcp://127.0.0.1:2376 │ DOWN   │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test1 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
├───────┼──────────────────────┼────────┼──────────┼───────────────────────┼─────────────┼─────────────┤
│ test2 │ tcp://127.0.0.1:2376 │ UP     │ 0.00 / 4 │ 0 / 13026197504 bytes │ 0 / 0 bytes │ 0 / 0 bytes │
└───────┴──────────────────────┴────────┴──────────┴───────────────────────┴─────────────┴─────────────┘
```

#### networks

This command will list all networks supported by this pod.

The format is `eru-cli pod networks [command options] <podname>`.

`<podname>` refers to the name of the pod.

Command options are:

- `--driver`

  - Defines the network driver to filter.
  - If this option is not defined, will show all networks.

An example is:

```
root@tonic-eru-test:~# eru-cli pod networks muroq
┌────────┬───────────────┐
│ NAME   │ NETWORK       │
├────────┼───────────────┤
│ none   │               │
│ bridge │ 172.17.0.0/16 │
│ host   │               │
└────────┴───────────────┘

root@tonic-eru-test:~# eru-cli pod networks --driver calico muroq
┌──────┬─────────┐
│ NAME │ NETWORK │
├──────┼─────────┤
└──────┴─────────┘
```

### Status Sub Commands

Status sub commands are started with `status` command, and only contains one command: `eru-cli status`. The format should be `eru-cli status [command options] <appname>`.

This command will show the events of the corresponding workloads. Usually used for debugging.

`<appname>` refers to the appname of workloads.

Command options are:

- `--entry`

  - Defines the entrypoint of the workloads.
  - If this option is not defined, will not filter events by entrypoint.

- `--node`

  - Defines the node of the workloads.
  - If this option is not defined, will not filter events by node.

- `--label`

  - Defines the labels to filter.
  - This option can be defined multiple times, like `--label rack=rack1 --label cluster=cluster3`.

An example is:

```
root@tonic-eru-test:~# eru-cli status --entry http test
WARN[2021-06-17 17:31:30] [5b8129e] test_http_RfKuXJ on test0 is unhealthy
INFO[2021-06-17 17:31:40] [5b8129e] test_http_RfKuXJ back to life
INFO[2021-06-17 17:31:40] [5b8129e] published at host bind 127.0.0.1:8000
WARN[2021-06-17 17:32:04] [5b8129e] test_http_RfKuXJ on test0 is stopped
WARN[2021-06-17 17:32:09] 5b8129e deleted
```

### Workload / Container Sub Commands

Workload / container sub commands are started with `workload` command. The format should be `eru-cli workload [sub command] [command options] [arguments...]`.

For legacy reasons, `container` sub commands are still supported, just use `container` instead of `workload`, but will be removed in the future version.

These sub commands are supported:

- `get`
- `logs`
- `get-status`
- `set-status`
- `list`
- `stop`
- `start`
- `restart`
- `remove`
- `copy`
- `send`
- `dissociate`
- `realloc`
- `exec`
- `replace`
- `deploy`

#### get

This command will get a workload and print its information.

The format is `eru-cli workload get [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

An example is:

```
root@tonic-eru-test:~# eru-cli workload get 90a99da902b1a92d3062ea17d01f38773e0091c93815bad038396c4e047da074 31252a3c678f65bfa9185c53b98f5cb063b39f4bf134e7508fe7e17050c3dfb5
┌──────────────────────────────────────────────────────────────────┬───────────────────────────┬──────────────────────────┬──────────────────────┐
│ NAME/ID/POD/NODE                                                 │ STATUS                    │ VOLUME                   │ NETWORKS             │
├──────────────────────────────────────────────────────────────────┼───────────────────────────┼──────────────────────────┼──────────────────────┤
│ test_http_KuiJNm                                                 │ CPUQuotaRequest: 1.000000 │ VolumesRequest: []       │ host: 127.0.0.1:8000 │
│ 90a99da902b1a92d3062ea17d01f38773e0091c93815bad038396c4e047da074 │ CPUQuotaLimit: 1.000000   │ VolumesLimit: []         │                      │
│ muroq                                                            │ CPUMap: map[]             │ VolumePlanRequest: map[] │                      │
│ test0                                                            │ MemoryRequest: 536870912  │ VolumePlanLimit: map[]   │                      │
│                                                                  │ MemoryLimit: 536870912    │                          │                      │
│                                                                  │ StorageRequest: 0         │                          │                      │
│                                                                  │ StorageLimit: 0           │                          │                      │
│                                                                  │ Privileged: false         │                          │                      │
├──────────────────────────────────────────────────────────────────┼───────────────────────────┼──────────────────────────┼──────────────────────┤
│ test_http_CpxCtZ                                                 │ CPUQuotaRequest: 1.000000 │ VolumesRequest: []       │                      │
│ 31252a3c678f65bfa9185c53b98f5cb063b39f4bf134e7508fe7e17050c3dfb5 │ CPUQuotaLimit: 1.000000   │ VolumesLimit: []         │                      │
│ muroq                                                            │ CPUMap: map[]             │ VolumePlanRequest: map[] │                      │
│ test1                                                            │ MemoryRequest: 536870912  │ VolumePlanLimit: map[]   │                      │
│                                                                  │ MemoryLimit: 536870912    │                          │                      │
│                                                                  │ StorageRequest: 0         │                          │                      │
│                                                                  │ StorageLimit: 0           │                          │                      │
│                                                                  │ Privileged: false         │                          │                      │
└──────────────────────────────────────────────────────────────────┴───────────────────────────┴──────────────────────────┴──────────────────────┘
```

Note that the IDs should not be truncated result.

#### logs

This command will print log stream of a workload.

The format is `eru-cli workload logs [command options] <workloadID>`.

`<workloadID>` refers to the ID of the workload.

Command options are:

- `--tail`

  - Defines the number of lines to show from the end, works just like `tail -n`.
  - If this option is not defined, will show all logs from the beginning.

- `--since`

  - Defines the timestamp for the log to show.
  - Works just like `docker logs --since`.
  - The value can be an absolute timestamp (e.g. `2013-01-02T13:23:37`) or relative time period (e.g. `42m` for 42 minutes).

- `--until`

  - Defines the timestamp for the log to show.
  - Works just like `docker logs --until`.
  - The value can be an absolute timestamp (e.g. `2013-01-02T13:23:37`) or relative time period (e.g. `42m` for 42 minutes).

- `--follow`, `-f`

  - This is a flag.
  - If this flag is defined, will follow the log output.

An example is:

```
root@tonic-eru-test:~# eru-cli workload logs --follow 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30
INFO[2021-06-17 18:10:00] [GetWorkloadLog] PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.
INFO[2021-06-17 18:10:00] [GetWorkloadLog] 64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.070 ms
INFO[2021-06-17 18:10:00] [GetWorkloadLog] 64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.046 ms
INFO[2021-06-17 18:10:00] [GetWorkloadLog] 64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=0.032 ms
INFO[2021-06-17 18:10:00] [GetWorkloadLog] 64 bytes from 127.0.0.1: icmp_seq=4 ttl=64 time=0.043 ms
INFO[2021-06-17 18:10:00] [GetWorkloadLog] 64 bytes from 127.0.0.1: icmp_seq=5 ttl=64 time=0.059 ms
```

#### get-status

This command will print status of workloads, from this we can know if the workload is running or is healthy.

The format is `eru-cli workload get-status [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

An example is:

```
root@tonic-eru-test:~# eru-cli workload get-status 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30
┌──────────────────────────────────────────────────────────────────┬───────────────┬─────────────────┬──────────────────────────────────────────────────────┐
│ ID                                                               │ STATUS        │ NETWORKS        │ EXTENSIONS                                           │
├──────────────────────────────────────────────────────────────────┼───────────────┼─────────────────┼──────────────────────────────────────────────────────┤
│ 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 │ Running: true │ host: 127.0.0.1 │ ERU: 1                                               │
│                                                                  │ Healthy: true │                 │ ERU_META: {"Publish":null,"HealthCheck":null}        │
│                                                                  │               │                 │ eru.coreid: 8b22a85572704a13ec7c61a9ceea14b4729a74f3 │
│                                                                  │               │                 │ eru.nodename: test0                                  │
└──────────────────────────────────────────────────────────────────┴───────────────┴─────────────────┴──────────────────────────────────────────────────────┘
```

#### set-status

This command will set status of workloads, usually used for debugging.

The format is `eru-cli workload set-status [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Command options are:

- `--running`

  - Defines the running status of workloads.
  - `--running=true` means set the running status to `true`, `--running=false` means set the running status to `false`.

- `--healthy`

  - Defines the healthy status of workloads.
  - `--healthy=true` means set the healthy status to `true`, `--healthy=false` means set the healthy status to `false`.

- `--ttl`

  - Defines the TTL of this status value.

- `--network`

  - Defines the network for this workload.
  - The format can be a single string representing the network name, like `--network calico` or `--network host`,
  - Or can be a string contains a ":" to specify the network name and IPv4 IP address, like `--network calico:10.233.12.41`.

- `--extension`

  - Defines the `extension` field of workload.

An example is:

```
root@tonic-eru-test:~# eru-cli workload set-status --running=false --healthy=false 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30
┌──────────────────────────────────────────────────────────────────┬────────────────┬──────────┬────────────┐
│ ID                                                               │ STATUS         │ NETWORKS │ EXTENSIONS │
├──────────────────────────────────────────────────────────────────┼────────────────┼──────────┼────────────┤
│ 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 │ Running: false │          │            │
│                                                                  │ Healthy: false │          │            │
└──────────────────────────────────────────────────────────────────┴────────────────┴──────────┴────────────┘
```

Note that this command is usually used for debugging, dont' use it unless you know what you are doing.

#### list

This command will list workloads using some filters.

The format is `eru-cli workload list [command options] <appname>`.

`<appname>` refers to the appname of workloads.

Command options are:

- `--entry`

  - Defines the entrypoint of workloads.
  - If defined, this will be used to filter the workloads.

- `--nodename`

  - Defines the node of workloads.
  - If defined, this will be used to filter the workloads.

- `--label`

  - Defines the labels to filter.
  - This option can be defined multiple times, like `--label rack=rack1 --label cluster=cluster3`.
  - If defined, labels will be used to filter workloads, act like `AND` operator.

- `--limit`

  - Defines the number of results returned.
  - If not defined, will show all results.

An example is:

```
root@tonic-eru-test:~# eru-cli workload list test
┌──────────────────────────────────────────────────────────────────┬───────────────────────────┬──────────────────────────┬─────────────────┐
│ NAME/ID/POD/NODE                                                 │ STATUS                    │ VOLUME                   │ NETWORKS        │
├──────────────────────────────────────────────────────────────────┼───────────────────────────┼──────────────────────────┼─────────────────┤
│ test_ping_SYClfp                                                 │ CPUQuotaRequest: 1.000000 │ VolumesRequest: []       │ host: 127.0.0.1 │
│ 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 │ CPUQuotaLimit: 1.000000   │ VolumesLimit: []         │                 │
│ muroq                                                            │ CPUMap: map[]             │ VolumePlanRequest: map[] │                 │
│ test0                                                            │ MemoryRequest: 536870912  │ VolumePlanLimit: map[]   │                 │
│                                                                  │ MemoryLimit: 536870912    │                          │                 │
│                                                                  │ StorageRequest: 0         │                          │                 │
│                                                                  │ StorageLimit: 0           │                          │                 │
│                                                                  │ Privileged: false         │                          │                 │
└──────────────────────────────────────────────────────────────────┴───────────────────────────┴──────────────────────────┴─────────────────┘

root@tonic-eru-test:~# eru-cli workload list --entry not-existing-entry test
┌──────────────────┬────────┬────────┬──────────┐
│ NAME/ID/POD/NODE │ STATUS │ VOLUME │ NETWORKS │
├──────────────────┼────────┼────────┼──────────┤
└──────────────────┴────────┴────────┴──────────┘
```

#### stop

This command will stop workloads.

Stopping the workloads of docker engine will first send a SIGTERM, then wait for 10 seconds, send a SIGKILL.

Stopping the workloads of virtual machines controlled by yavirtd will do graceful shutdown.

The format is `eru-cli workload stop [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Command options are:

- `--force`

  - This is a flag.
  - If this option is defined, stopping docker workloads will send SIGTERM then immediately SIGKILL, stopping yavirtd workloads will shutdown immediately.
  - No matter this option is defined or not, the `BEFORE_STOP` hooks will be executed anyway.

#### start

This command will start workloads.

Starting the workloads will first start all the workloads, then execute the `AFTER_START` hooks.

The format is `eru-cli workload start [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Command options are:

- `--force`

  - This is a flag.
  - If this option is defined, any errors from `AFTER_START` hooks will be ignored, otherwise errors will be returned.

#### restart

This command will restart workloads.

Restart is actually a combination of stop and start.

The format is `eru-cli workload start [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Command options are:

- `--force`

  - This is a flag.
  - Refer to `stop` and `start` for the effect of this flag, this will affect both of them.

#### remove

This command will remove workloads. Usually only stopped workloads can be removed.

The format is `eru-cli workload remove [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Command options are:

- `--force`

  - This is a flag.
  - If this option is defined, the workloads will be removed regardless of their status.

- `--step`

  - Defines the concurrent step.
  - The default value is `1`, which means the workloads will be removed one by one.

#### copy

This command will copy files from workload to local environment.

The format is `eru-cli workload copy [command options] workloadID:path1,path2,...`.

`workloadID` refers to the ID of the workload.

`path1,path2,path3,...` refers to the paths to be copied, separated by `,`.

The copied files are packed into a tarball for docker engine.

Command options are:

- `--dir`

  - Defines the directory to store the copied files.
  - Default value is `/tmp`.

An example is:

```
root@tonic-eru-test:~/copy# eru-cli workload copy --dir `pwd` 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30:/wow,/tbc

root@tonic-eru-test:~/copy# ls
47ae978-tbc-2021.06.17.19.25.24.tar  47ae978-wow-2021.06.17.19.25.24.tar

root@tonic-eru-test:~/copy# tar xf 47ae978-tbc-2021.06.17.19.25.24.tar

root@tonic-eru-test:~/copy# tar xf 47ae978-wow-2021.06.17.19.25.24.tar

root@tonic-eru-test:~/copy# ls
47ae978-tbc-2021.06.17.19.25.24.tar  47ae978-wow-2021.06.17.19.25.24.tar  tbc  wow

root@tonic-eru-test:~/copy# cat wow
zandalar forever
```

#### send

This command will send files from local environment to workloads.

The format is `eru-cli workload send [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

`path1,path2,path3,...` refers to the paths to be copied, separated by `,`.

The copied files are packed into a tarball for docker engine.

Command options are:

- `--file`

  - Defines the file to send.
  - Format is `SRC_PATH:DEST_PATH`, can be defined multiple times, like `--file localfile1:/path/in/workload/file1 --file localfile2:/path/in/workload/file2`

An example is:

```
root@tonic-eru-test:~/copy# echo "fae mage did 7 flame strikes in combustion!" >> SL

root@tonic-eru-test:~/copy# eru-cli workload send --file ./SL:/SLfile 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30
INFO[2021-06-17 19:38:57] [Send] Send /SLfile to 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 success

root@tonic-eru-test:~/copy# docker exec 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 cat /SLfile
fae mage did 7 flame strikes in combustion!
```

Note: The directory of the file path will not be created automatically, hence the path in the workload MUST exist, otherwise you will get an error. For example, if you specify `--file localfile:/a/nonexisting/path/file`, the `/a/nonexisting/path` should be created before.

#### dissociate

This command will dissociate workloads from ERU system.

The format is `eru-cli workload send [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Dissociating the workloads will release the bindings of workloads and ERU system. Sometimes you may find that ERU claims it has the record of a workload, but the workload doesn't exist on any nodes, now dissociating this workload will return its resources back to ERU system without removing it. Later you can remove it or keep it for debugging. Since the actual workload has gone away, we usually remove it.

Also sometimes you may want to release a workload from ERU system, you can use this command too.

#### realloc

This command will realloc resources of workloads.

The format is `eru-cli workload realloc [command options] workloadID(s)`.

`workloadID(s)` refers to the IDs of the workloads.

Note: all the values in options are **delta**, the final result will be `current value + delta value`.

Command options are:

- `--cpu-request`, `--cpu-limit`

  - Defines the CPU request and limit **delta** value.
  - Default value is `0`, means don't change the CPU request / limit value.
  - For example, currently the workload has occupied 2 cores:

    - `--cpu-request 1` will reallocate 1 more core to this workload, it currently occupied 3 cores.
    - `--cpu-request -1` will reallocate to 1 less core to this workload, it currently occupied 1 core.
    - `--cpu-request 0` will not reallocate cores to this workload, it still occupied 2 cores.

- `--cpu`

  - This is a shortcut for `--cpu-request` and `--cpu-limit`
  - If you define `--cpu 1`, it's equivalent to `--cpu-request 1 --cpu-limit 1`

- `--cpu-bind`

  - This is a flag.
  - If this flag is defined, workloads will bind fixed CPUs.
  - If workload has already bound to fixed CPUs, this flag takes no effect.

- `--cpu-unbind`

  - This is a flag.
  - If this flag is defined, workloads will release the bound fixed CPUs.
  - If workload has not bound to any fixed CPUs, this flag takes no effect.

- `--memory-request`, `--memory-limit`

  - Defines the memory request and limit **delta** value.
  - Default value is `0`, means don't change the memory request / limit value.
  - Units are supported, you can use `--memory-request 128G` or `--memory-limit 1024M` to set the value.
  - It works just like `--cpu-request` and `--cpu-limit`

- `--memory`

  - This is a shortcut for `--memory-request` and `--memory-limit`
  - If you define `--memory 10G`, it's equivalent to `--memory-request 10G --memory-limit 10G`

- `--storage-request`, `--storage-limit`

  - Defines the storage request and limit **delta** value.
  - Default value is `0`, means don't change the storage request / limit value.
  - Units are supported, you can use `--storage-request 128G` or `--storage-limit 1024M` to set the value.
  - It works just like `--cpu-request` and `--cpu-limit`

- `--storage`

  - This is a shortcut for `--storage-request` and `--storage-limit`
  - If you define `--storage 10G`, it's equivalent to `--storage-request 10G --storage-limit 10G`

- `--volumes-request`, `--volumes-limit`

  - Defines the volumes request and limit **delta** value.
  - Default value is `0`, means don't change the volumes request / limit value.
  - Format is `AUTO:{path}:{mode}:{size}` or `{workloadpath}:{hostpath}`

    - `AUTO:/data:rw:-1G` means automatically choose a volume from host and bind it with `/data` in workload, the mode is `rw` which means read and write, the size is decreased by 1G, for how this AUTO thing works, refer to eru-core documents.
    - `/tmp:/tmp` means still keep `/tmp` to `/tmp` in workload.

  - Units are supported, you can use `--volumes-request AUTO:/data:rw:-10G` or `--volumes-limit AUTO:/data:rw:-10240M` to set the value.
  - It will increase or decrease the size of the volumes.

#### exec

This command can execute commands in a workload.

The format is `eru-cli workload exec [command options] workloadID <commands...>`.

`workloadID(s)` refers to the IDs of the workloads.

`<commands...>` refers to the commands to be executed.

Command options are:

- `--interactive`, `-i`

  - This is a flag.
  - If this flag is defined, you can interact from stdin with the workload, otherwise you can only see the output of your defined commands but can not send new commands to the workload.

- `--env`, `-e`

  - Defines the environment variables for this execution.
  - This option can be defined multiple times, like `--env A=1 --env B=2`.

- `--workdir`, `-w`

  - Defines the current working directory for this execution.
  - Default value is `/`.

An example is:

```
root@tonic-eru-test:~/copy# eru-cli workload exec --env EV1=TONIC --env EV2=SONIC 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 env
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=tonic-eru-test
APP_NAME=test
ERU_POD=muroq
ERU_NODE_NAME=test0
ERU_WORKLOAD_SEQ=0
ERU_MEMORY=536870912
ERU_STORAGE=0
ERU_NODE_IP=127.0.0.1
EV1=TONIC
EV2=SONIC
HOME=/root
```

#### deploy

This command can deploy workloads from a specification file.

The format is `eru-cli workload deploy [command options] <specification-file>`.

`<specification-file>` refers to the path of local specification file, or a remote URL of specification file.

Command options are:

- `--dry-run`

  - This is a flag.
  - If this flag is defined, eru-cli will not really deploy the workloads, but only shows the deployment plan instead.

- `--pod`

  - Defines which pod to deploy.

- `--node`

  - Defines which nodes to deploy.
  - This option can be defined multiple times, like `--node node1 --node node2`.
  - When this option is defined, `--pod` is ignored, which means the nodes specified in `--node` don't have to belong to the pod `--pod` specified.

- `--nodelabel`

  - Defines the labels to filter nodes.
  - This option can be defined multiple times, like `--nodelabel key1=value1 --nodelabel key2=value2`.
  - If this option is defined, only the nodes with all the labels defined are selected.

- `--entry`

  - Defines the entrypoint.
  - Entrypoint must exist in the specification file.

- `--image`

  - Defines which image to use.

- `--count`

  - Defines how many workloads to deploy.
  - Default value is `1`.

- `--network`

  - Defines the network of workload.
  - The format can be a single string representing the network name, like `--network calico` or `--network host`,
  - Or can be a string contains a ":" to specify the network name and IPv4 IP address, like `--network calico:10.233.12.41`.

- `--cpu-request`, `--cpu-limit`, `--cpu`

  - Defines the CPU request and limit value.
  - Default value is `0`, means don't limit CPU request, and share CPU with other workloads.
  - `--cpu` is a shortcut for `--cpu-request` and `--cpu-limit`, `--cpu 2` is equivalent to `--cpu-request 2 --cpu-limit 2`.

- `--cpu-bind`

  - This is a flag.
  - If this flag is defined, workload will bind to fixed CPUs exclusively.

- `--memory-request`, `--memory-limit`, `--memory`

  - Defines the memory request and limit value.
  - Default value is `0`, means don't limit memory request, and share memory with other workloads (which may lead to OOM killed).
  - `--memory` is a shortcut for `--memory-request` and `--memory-limit`, `--memory 2G` is equivalent to `--memory-request 2G --memory-limit 2G`.
  - Units are supported, you can use `--memory-request 128G` or `--memory-limit 1024M` to set the value.

- `--storage-request`, `--storage-limit`, `--storage`

  - Defines the storage request and limit value.
  - Default value is `0`, means don't limit storage request, and share storage with other workloads.
  - `--storage` is a shortcut for `--storage-request` and `--storage-limit`, `--storage 100G` is equivalent to `--storage-request 100G --storage-limit 100G`.
  - Units are supported, you can use `--storage-request 100G` or `--storage-limit 1T` to set the value.

- `--env`

  - Defines the environment variables for this workload.
  - This option can be defined multiple times, like `--env A=1 --env B=2`.

- `--file`

  - Defines the file to send to workloads.
  - Format is `SRC_PATH:DEST_PATH`, can be defined multiple times, like `--file localfile1:/path/in/workload/file1 --file localfile2:/path/in/workload/file2`

- `--user`

  - Defines the user to run the process in workloads.

- `--deploy-strategy`

  - Defines the deploy strategy.
  - Possible choices are `AUTO`, `FILL`, `EACH`, `GLOBAL`.
  - Default value is `AUTO`.
  - For details of each strategy, refer to eru-core documents.

- `--auto-replace`

  - This is a flag.
  - If this flag is defined, eru-core will create workloads if there're none, or replace the workloads of specified `appname`, `entrypoint` with new `image`, and `entrypoint`, etc.

Command arguments are:

- `<specification-file>`

  - This can be a local path to the file, or a remove URL to the file.
  - An example of a specification file is like:

    ```
    appname: "test"
    entrypoints:
      http:
        cmd: "python3 -m http.server"
        restart: always
        publish:
          - "8000"
        healthcheck:
          tcp_ports:
            - "8000"
      ping:
        cmd: "ping 127.0.0.1"
    ```

An example is:

```
root@tonic-eru-test:~# eru-cli workload deploy --pod muroq --entry ping --image tonic/ubuntu:phistage2 spec.yaml
INFO[2021-06-17 18:07:41] [Deploy] Success 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 test_ping_SYClfp test0 1 1 map[] 536870912 536870912 map[] map[]

root@tonic-eru-test:~# eru-cli workload get 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30
┌──────────────────────────────────────────────────────────────────┬───────────────────────────┬──────────────────────────┬─────────────────┐
│ NAME/ID/POD/NODE                                                 │ STATUS                    │ VOLUME                   │ NETWORKS        │
├──────────────────────────────────────────────────────────────────┼───────────────────────────┼──────────────────────────┼─────────────────┤
│ test_ping_SYClfp                                                 │ CPUQuotaRequest: 1.000000 │ VolumesRequest: []       │ host: 127.0.0.1 │
│ 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 │ CPUQuotaLimit: 1.000000   │ VolumesLimit: []         │                 │
│ muroq                                                            │ CPUMap: map[]             │ VolumePlanRequest: map[] │                 │
│ test0                                                            │ MemoryRequest: 536870912  │ VolumePlanLimit: map[]   │                 │
│                                                                  │ MemoryLimit: 536870912    │                          │                 │
│                                                                  │ StorageRequest: 0         │                          │                 │
│                                                                  │ StorageLimit: 0           │                          │                 │
│                                                                  │ Privileged: false         │                          │                 │
└──────────────────────────────────────────────────────────────────┴───────────────────────────┴──────────────────────────┴─────────────────┘
```

For more details of deploying workloads, refer to eru-core documents.

#### replace

This command can replace workloads with a specification file.

The format is `eru-cli workload replace [command options] <specification-file>`.

`<specification-file>` refers to the path of local specification file, or a remote URL of specification file.

Replace will replace all workloads filtered by `--pod`, `--node`, `--entry` and `appname` in specification file, with new `--image`, `--count`, `--env`, `--user`... etc.

Command options are:

- `--pod`

  - Defines which pod to filter.

- `--node`

  - Defines which nodes to filter.
  - This option can be defined multiple times, like `--node node1 --node node2`.
  - When this option is defined, only the workloads on these nodes and also in pod specified by `--pod` will be selected to replace.

- `--entry`

  - Defines the entrypoint to filter.
  - Entrypoint must exist in the specification file.

- `--image`

  - Defines new image to use.

- `--count`

  - Defines how many workloads to replace at the sametime.
  - Default value is `1`, which means workloads are replaced one by one.

- `--network`

  - Defines the network of workload.
  - The format can be a single string representing the network name, like `--network calico` or `--network host`,
  - Or can be a string contains a ":" to specify the network name and IPv4 IP address, like `--network calico:10.233.12.41`.

- `--network-inherit`

  - This is a flag.
  - If this flag is defined, `--network` will be ignored, the original networks are used in new workloads.
  - This is actually the design target for `replace`.

- `--env`

  - Defines the environment variables for this workload.
  - This option can be defined multiple times, like `--env A=1 --env B=2`.

- `--file`

  - Defines the file to send to workloads.
  - Format is `SRC_PATH:DEST_PATH`, can be defined multiple times, like `--file localfile1:/path/in/workload/file1 --file localfile2:/path/in/workload/file2`

- `--user`

  - Defines the user to run the process in workloads.

- `--label`

  - Defines the labels to filter.
  - This option can be defined multiple times, like `--label rack=rack1 --label cluster=cluster3`.
  - If defined, labels will be used to filter workloads, only the workloads with all labels will be selected to replace.

- `--copy`

  - Defines the files copied from original workload to new workload.
  - Format is `SRC_PATH:DEST_PATH`, this option can be defined multiple times, like `--copy /old/path/file1:/new/path/file1 --copy /old/path/file2:/new/path/file2`.

An example is:

```
root@tonic-eru-test:~# eru-cli workload replace --pod muroq --entry ping --network-inherit --image tonic/ubuntu:phistage2 --env V1=TONIC --env V2=AARON --copy /tbc:/tbccopy --copy /wow:/wowcopy spec.yaml
INFO[2021-06-17 21:05:35] [Replace] Replace 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30
INFO[2021-06-17 21:05:35] [Replace] Hook workload 47ae97833e3042c57763206901b348c1956e53928e44007952d9c5b4f958db30 removed
INFO[2021-06-17 21:05:35] [Replace] New workload test_ping_qrdoMW, cpu map[], quotaRequest 1, quotaLimit 1, memRequest 536870912, memLimit 536870912

root@tonic-eru-test:~# eru-cli workload list
┌──────────────────────────────────────────────────────────────────┬───────────────────────────┬──────────────────────────┬─────────────────┐
│ NAME/ID/POD/NODE                                                 │ STATUS                    │ VOLUME                   │ NETWORKS        │
├──────────────────────────────────────────────────────────────────┼───────────────────────────┼──────────────────────────┼─────────────────┤
│ test_ping_qrdoMW                                                 │ CPUQuotaRequest: 1.000000 │ VolumesRequest: []       │ host: 127.0.0.1 │
│ 4280528b8785955df3d3ca52b1e5dbda36548df18c66745f6b0b1d648ac98729 │ CPUQuotaLimit: 1.000000   │ VolumesLimit: []         │                 │
│ muroq                                                            │ CPUMap: map[]             │ VolumePlanRequest: map[] │                 │
│ test0                                                            │ MemoryRequest: 536870912  │ VolumePlanLimit: map[]   │                 │
│                                                                  │ MemoryLimit: 536870912    │                          │                 │
│                                                                  │ StorageRequest: 0         │                          │                 │
│                                                                  │ StorageLimit: 0           │                          │                 │
│                                                                  │ Privileged: false         │                          │                 │
└──────────────────────────────────────────────────────────────────┴───────────────────────────┴──────────────────────────┴─────────────────┘

root@tonic-eru-test:~# docker exec -it 4280528b8785955df3d3ca52b1e5dbda36548df18c66745f6b0b1d648ac98729 /bin/bash
root@tonic-eru-test:/# ls
bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tbccopy  tmp  usr  var  wowcopy
root@tonic-eru-test:/# cat wowcopy
zandalar forever

root@tonic-eru-test:~# eru-cli workload exec 4280528b8785955df3d3ca52b1e5dbda36548df18c66745f6b0b1d648ac98729 env
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=tonic-eru-test
V1=TONIC
V2=AARON
APP_NAME=test
ERU_POD=muroq
ERU_NODE_NAME=test0
ERU_WORKLOAD_SEQ=0
ERU_MEMORY=536870912
ERU_STORAGE=0
ERU_NODE_IP=127.0.0.1
HOME=/root
```

Note that:

- Workload ID has beend changed since it's a new workload.
- `wow` is copied to `wowcopy`, `tbc` is copied to `tbccopy`.
- New environment variables are set in new workloads (`V1=TONIC`, `V2=AARON`).
