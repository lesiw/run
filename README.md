# gx (git exec)

Run commands relative to the root of the git repository.

## Installation

### curl

```sh
curl -L lesiw.io/gx | sh
```

### go install

```sh
go install lesiw.io/gx@latest
```

## Usage

```
Usage of gx:

    gx COMMAND [ARGS...]

  -i    install completion scripts
  -l    list all commands
  -r    print git root
  -u mapping
        chowns files based on a given mapping uid:gid::uid:gid
  -v    print version
```

## Configuration

* `GXPATH`: Defaults to `./bin`. Set to `-` to disable.

## Completion

Install bash/zsh completion:

```sh
sudo "$(which gx)" -i
```

After running the command, follow the printed instructions.
