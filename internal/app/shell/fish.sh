complete -c {{ $.Executable }} -f


function {{ $.Executable }}-completion
  set -f commands ""
{{- range $idx, $value := $.Options }}
  {{- if gt $idx 0 }}
  set -f commands " $commands"
  {{ end }}
  if {{ $value.Conditional }}
    set -f commands "{{ $value.Key }}$commands"
  end
{{- end }}
  complete -c {{ $.Executable }} -n "not __fish_seen_subcommand_from $commands" -a "$commands"
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.HelpCommand }}; and test (count (commandline -opc)) -lt 3" -a "{{ $.HelpAdvancedCommand }}"
  if {{ $.Conditionals.Not.ReadOnly }}
    if {{ $.Conditionals.Not.AskMode }}
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.InsertCommand }} {{ $.MultiLineCommand }} {{ $.RemoveCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList }})"
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.MoveCommand }}; and test (count (commandline -opc)) -lt 4" -a "({{ $.DoList }})"
    end
  end
  if {{ $.Conditionals.Not.NoTOTP }}
    set -f totps ""
{{- range $idx, $value := $.TOTPSubCommands }}
  {{- if gt $idx 0 }}
    set -f totps " $totps"
  {{ end }}
    if {{ $value.Conditional }}
      set -f totps "{{ $value.Key }}$totps"
    end
{{- end }}
    complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.TOTPCommand }}; and not __fish_seen_subcommand_from $totps" -a "$totps"
    if {{ $.Conditionals.Not.AskMode }}
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.TOTPCommand }}; and __fish_seen_subcommand_from $totps; and test (count (commandline -opc)) -lt 4" -a "({{ $.DoTOTPList }})"
    end
  end
  if {{ $.Conditionals.Not.NoClip }} 
    if {{ $.Conditionals.Not.AskMode }}
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.ClipCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList}})"
    end
  end
  if {{ $.Conditionals.Not.AskMode }}
    complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.ShowCommand }} {{ $.JSONCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList}})"
  end
end

{{ $.Executable }}-completion
