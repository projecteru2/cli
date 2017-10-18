Cli
=====

cli for Eru.

Modify resources for eru pods / nodes, manipulate containers and images.

### Usage

* Use `cli -h` to show commands and subcommands.
* Currently supported commands are:
	* `container` subcommands:
		* `container get {id} ...`
		* `container remove {id} ...`
		* `container realloc --cpu {cpu} --mem {mem}`
		* `container deploy`
	* `pod` subcommands:
		* `pod list`
		* `pod add`
		* `pod nodes {podname}`
		* `pod networks {podname}`
	* `node` subcommands:
		* `node get {nodename}`
		* `node remove {nodename}`
		* `node available {nodename} --available`
	* `image` subcommands:
		* `image build`
	* `lambda`

### Develop

Commands' source code in `commands` dir, you can define your own commands inside. Use `make test` to test and `make build` to build. If you want to modify and build in local, you can use `make deps` to generate vendor dirs.

