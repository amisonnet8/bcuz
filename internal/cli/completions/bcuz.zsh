#compdef bcuz
# zsh completion for bcuz.
#
# Install it with:  bcuz --completion zsh > "${fpath[1]}/_bcuz"
# or try it with:   source <(bcuz --completion zsh)
#
# After the reason, the words are shifted so that the command comes first,
# and the completion of that command is left to _normal.
_bcuz() {
    # The words before the one being typed, without "bcuz".
    local -a w list
    w=("${(@)words[2,CURRENT-1]}")
    local cur=${words[CURRENT]} n=${#w} start opt

    if (( n == 0 )); then
        # The reason is free text: only the options are offered.
        [[ $cur == -* ]] && compadd -- -H --history --init -h --help --version --completion --
        return
    fi

    case ${w[1]} in
    (--completion)
        (( n == 1 )) && compadd -- bash zsh fish powershell
        return
        ;;
    (-H|--history)
        case ${w[n]} in
        (--date|--since|--until) return ;; # a date
        esac
        for opt in - --date --since --until; do
            (( ${w[(Ie)$opt]} )) || list+=("$opt")
        done
        (( ${#list} )) && compadd -- "${list[@]}"
        return
        ;;
    (--) start=4 ;; # bcuz -- <reason> <command>
    (-*) return ;;
    (*) start=3 ;; # bcuz <reason> <command>
    esac

    (( CURRENT < start )) && return # the reason
    shift $(( start - 1 )) words
    (( CURRENT -= start - 1 ))
    _normal
}

if [ "$funcstack[1]" = "_bcuz" ]; then
    _bcuz "$@"
else
    compdef _bcuz bcuz
fi
