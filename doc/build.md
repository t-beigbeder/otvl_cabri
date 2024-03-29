# Building Cabri DSS application

## Building from source

Instructions provided below must be run on a Linux system.

Install a golang 1.21+ distribution, instructions are available
[here](https://go.dev/install). 

Install the following golang support tool:

    $ go install golang.org/x/tools/cmd/goimports@latest

Install a git distribution and clone the source distribution from gitlab:

    $ git clone https://github.com/t-beigbeder/otvl_cabri

Build the executable:

    $ cd otvl_cabri/
    $ dev_tools/build_go_fast.sh 
    2022/12/20 17:34:55,000 | INFO | build_go_fast.sh: starting
    2022/12/20 17:35:00,000 | INFO | build_go_fast.sh: ended

You will get the Linux AMD64 executable in `gocode/build/cabri`
and the Windows AMD64 one in `gocode/build/cabri.exe`. 
Copy them wherever you want.

Golang tooling enables to target other supported platforms as well, just adapt the script to your needs.
