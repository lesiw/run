#!/usr/bin/env bash

__pb_completion () {
    case "${COMP_WORDS[COMP_CWORD]}" in
        -*) suggestions="-i -r -l"
            ;;
        *)
            suggestions="$(pb -l)"
            ;;
    esac
    [ -z "$suggestions" ] && return 0
    COMPREPLY=()
    while read -r suggestion
    do
        COMPREPLY+=("$suggestion")
    done < <(compgen -W "$suggestions" -- "${COMP_WORDS[COMP_CWORD]}")
}

complete -F __pb_completion pb
