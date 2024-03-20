#compdef gx

_gx_tasks() {
    local -a tasks
    IFS=$'\n'
    tasks=($(gx -l 2>/dev/null))
    _describe 'tasks' tasks
}

_arguments \
    '-i[Install autocomplete scripts.]' \
    '-r[Print root.]' \
    '-l[List tasks.]' \
    '*:task:_gx_tasks'
