#compdef pb

_pb_tasks() {
    local -a tasks
    IFS=$'\n'
    tasks=($(pb -l 2>/dev/null))
    _describe 'tasks' tasks
}

_arguments \
    '-i[Install autocomplete scripts.]' \
    '-r[Print root.]' \
    '-l[List tasks.]' \
    '*:task:_pb_tasks'
