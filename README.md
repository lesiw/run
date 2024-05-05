# run: Contextual commands

Run commands relative to the root of the git repository.

## Installation

### curl

```sh
curl lesiw.io/run | sh
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

* `RUNPATH`: Defaults to `.`. Unlike `PATH`, it will search the given
  directories' `.run` directories for executables.

## Completion

Install bash/zsh completion:

```sh
sudo "$(which run)" --install-completions
```

After running the command, follow the printed instructions.
