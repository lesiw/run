#compdef run

_run_tasks() {
    local -a tasks
    IFS=$'\n'
    tasks=($(run -l 2>/dev/null))
    _describe 'tasks' tasks
}

_arguments \
    '-i[Install autocomplete scripts.]' \
    '-r[Print root.]' \
    '-l[List tasks.]' \
    '*:task:_run_tasks'
