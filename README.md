# run: Contextual commands

Run commands relative to the root of the git repository.

## Installation

### curl

```sh
curl -L lesiw.io/run | sh
```

### go install

```sh
go install lesiw.io/run@latest
```

## Usage

```
Usage of run:

    run COMMAND [ARGS...]

  -V    print version
  -i    install completion scripts
  -l    list all commands
  -r    print git root
  -u mapping
        chowns files based on a given mapping (uid:gid::uid:gid)
  -v    verbose
```

## Configuration

* `RUNPATH`: Defaults to `./bin`. Set to `-` to disable.

## Completion

Install bash/zsh completion:

```sh
sudo "$(which run)" -i
```

After running the command, follow the printed instructions.
