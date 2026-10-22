# PowerShell completion for bcuz.
#
# Install it with:  bcuz --completion powershell >> $PROFILE
# or try it with:   bcuz --completion powershell | Out-String | Invoke-Expression
#
# After the reason, the rest of the line is completed with CompleteInput, so
# the completion of that command is used.
Register-ArgumentCompleter -Native -CommandName bcuz -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # The elements before the one being typed, without "bcuz", and their words.
    $before = @($commandAst.CommandElements | Select-Object -Skip 1 |
        Where-Object { $_.Extent.EndOffset -lt $cursorPosition })
    $w = @($before | ForEach-Object {
            if ($_ -is [System.Management.Automation.Language.StringConstantExpressionAst]) { $_.Value } else { $_.Extent.Text }
        })
    $n = $w.Count

    $offer = {
        param([string[]]$values)
        foreach ($value in $values) {
            if ($value.StartsWith($wordToComplete, [System.StringComparison]::Ordinal)) {
                [System.Management.Automation.CompletionResult]::new($value, $value, 'ParameterValue', $value)
            }
        }
    }

    if ($n -eq 0) {
        # The reason is free text: only the options are offered.
        if ($wordToComplete.StartsWith('-')) {
            & $offer @('-H', '--history', '--init', '-h', '--help', '--version', '--completion', '--')
        }
        return
    }

    $start = 1 # bcuz <reason> <command>
    if ($w[0] -ceq '--completion') {
        if ($n -eq 1) { & $offer @('bash', 'zsh', 'fish', 'powershell') }
        return
    } elseif ($w[0] -ceq '-H' -or $w[0] -ceq '--history') {
        if (@('--date', '--since', '--until') -ccontains $w[$n - 1]) { return } # a date
        $used = if ($n -gt 1) { $w[1..($n - 1)] } else { @() }
        & $offer @(@('-', '--date', '--since', '--until') | Where-Object { $used -cnotcontains $_ })
        return
    } elseif ($w[0] -ceq '--') {
        $start = 2 # bcuz -- <reason> <command>
    } elseif ($w[0].StartsWith('-')) {
        return
    }

    if ($n -lt $start) { return } # the reason
    if ($n -eq $start) {
        $rest = $wordToComplete
    } else {
        $from = $before[$start].Extent.StartOffset
        $rest = $commandAst.Extent.Text.Substring($from - $commandAst.Extent.StartOffset, $cursorPosition - $from)
    }
    [System.Management.Automation.CommandCompletion]::CompleteInput($rest, $rest.Length, $null).CompletionMatches
}
