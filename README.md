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

## Configuration

* `GX_PATH`: Defaults to `./bin`. Set to `-` to disable.

## Completion

Install bash/zsh completion:

```sh
sudo "$(which gx)" -i
```

After running the command, follow the printed instructions.
