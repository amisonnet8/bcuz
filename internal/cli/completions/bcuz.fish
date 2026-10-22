# fish completion for bcuz.
#
# Install it with:  bcuz --completion fish > ~/.config/fish/completions/bcuz.fish
# or try it with:   bcuz --completion fish | source
#
# After the reason, the rest of the line is completed with complete -C, so
# the completion of that command is used.

function __bcuz_complete
    # The words before the one being typed, without "bcuz".
    set -l w (commandline -opc)
    set -e w[1]
    set -l cur (commandline -ct)
    set -l n (count $w)

    if test $n -eq 0
        # The reason is free text: only the options are offered.
        if string match -q -- '-*' "$cur"
            printf '%s\n' -H --history --init -h --help --version --completion --
        end
        return
    end

    set -l start
    if test "$w[1]" = --completion
        test $n -eq 1; and printf '%s\n' bash zsh fish powershell
        return
    else if contains -- "$w[1]" -H --history
        contains -- "$w[-1]" --date --since --until; and return # a date
        for opt in - --date --since --until
            contains -- $opt $w[2..-1]; or echo $opt
        end
        return
    else if test "$w[1]" = --
        set start 3 # bcuz -- <reason> <command>
    else if string match -q -- '-*' "$w[1]"
        return
    else
        set start 2 # bcuz <reason> <command>
    end

    test $n -lt (math $start - 1); and return # the reason
    set -l rest
    if test $n -ge $start
        set rest (string escape -- $w[$start..-1])
    end
    complete -C (string join ' ' -- $rest "$cur")
end

complete -c bcuz -f -a '(__bcuz_complete)'
