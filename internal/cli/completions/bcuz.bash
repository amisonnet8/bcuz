# bash completion for bcuz.
#
# Install it with:  bcuz --completion bash > ~/.local/share/bash-completion/completions/bcuz
# or try it with:   source <(bcuz --completion bash)
#
# After the reason, the command and its arguments are left to the completion
# of that command (_command_offset of bash-completion). Without
# bash-completion, command names and files are offered.

# _bcuz_reply puts what compgen offers for its arguments in COMPREPLY. It reads
# the lines one by one, as mapfile is not in bash 3.2 (the bash of macOS).
_bcuz_reply() {
    local line
    while IFS= read -r line; do
        COMPREPLY+=("$line")
    done < <(compgen "$@")
}

_bcuz() {
    local cur=${COMP_WORDS[COMP_CWORD]} start opt
    # The words before the one being typed, without "bcuz".
    local -a words=("${COMP_WORDS[@]:1:COMP_CWORD-1}") list=()
    local n=${#words[@]}
    COMPREPLY=()

    if ((n == 0)); then
        # The reason is free text: only the options are offered.
        if [[ $cur == -* ]]; then
            _bcuz_reply -W '-H --history --init -h --help --version --completion --' -- "$cur"
        fi
        return 0
    fi

    case ${words[0]} in
    --completion)
        if ((n == 1)); then
            _bcuz_reply -W 'bash zsh fish powershell' -- "$cur"
        fi
        return 0
        ;;
    -H | --history)
        case ${words[n - 1]} in
        --date | --since | --until) return 0 ;; # a date
        esac
        for opt in - --date --since --until; do
            [[ " ${words[*]:1} " == *" $opt "* ]] || list+=("$opt")
        done
        _bcuz_reply -W "${list[*]}" -- "$cur"
        return 0
        ;;
    --) start=3 ;; # bcuz -- <reason> <command>
    -*) return 0 ;;
    *) start=2 ;; # bcuz <reason> <command>
    esac

    ((COMP_CWORD < start)) && return 0 # the reason
    if declare -F _command_offset >/dev/null; then
        _command_offset "$start"
    elif ((COMP_CWORD == start)); then
        _bcuz_reply -c -- "$cur"
    else
        compopt -o filenames 2>/dev/null
        _bcuz_reply -f -- "$cur"
    fi
    return 0
}

complete -F _bcuz bcuz
