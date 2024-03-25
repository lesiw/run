# pb: project builder

Run commands relative to the root of the git repository.

## Installation

### curl

```sh
curl -L lesiw.io/pb | sh
```

### go install

```sh
go install lesiw.io/pb@latest
```

## Usage

```
Usage of pb:

    pb COMMAND [ARGS...]

  -V    print version
  -i    install completion scripts
  -l    list all commands
  -r    print git root
  -u mapping
        chowns files based on a given mapping (uid:gid::uid:gid)
  -v    verbose
```

## Configuration

* `PBPATH`: Defaults to `./bin`. Set to `-` to disable.

## Completion

Install bash/zsh completion:

```sh
sudo "$(which pb)" -i
```

After running the command, follow the printed instructions.
