#!/usr/bin/env bash

__run_completion () {
    case "${COMP_WORDS[COMP_CWORD]}" in
        -*) suggestions="-i -r -l"
            ;;
        *)
            suggestions="$(run -l)"
            ;;
    esac
    [ -z "$suggestions" ] && return 0
    COMPREPLY=()
    while read -r suggestion
    do
        COMPREPLY+=("$suggestion")
    done < <(compgen -W "$suggestions" -- "${COMP_WORDS[COMP_CWORD]}")
}

complete -F __run_completion run
