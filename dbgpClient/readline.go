package main

import (
	"fmt"
	"github.com/chzyer/readline"     // MIT
	. "github.com/logrusorgru/aurora" // WTFPL
	"os/user"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("breakpoint_get -d"),
	readline.PcItem("breakpoint_list"),
	readline.PcItem("breakpoint_remove -d"),
	readline.PcItem("breakpoint_get -d"),
	readline.PcItem("breakpoint_set",
		readline.PcItem("-t line",
			readline.PcItem("-f"),
			readline.PcItem("-n"),
		),
		readline.PcItem("-t conditional",
			readline.PcItem("-f"),
			readline.PcItem("-n"),
			readline.PcItem("--"),
		),
		readline.PcItem("-t call",
			readline.PcItem("-a"),
			readline.PcItem("-m"),
		),
		readline.PcItem("-t return",
			readline.PcItem("-a"),
			readline.PcItem("-m"),
		),
		readline.PcItem("-t exception",
			readline.PcItem("-x"),
		),
		readline.PcItem("-t watch"),
		readline.PcItem("-h"),
		readline.PcItem("-o >="),
		readline.PcItem("-o =="),
		readline.PcItem("-o %"),
		readline.PcItem("-s enabled"),
		readline.PcItem("-s disabled"),
	),
	readline.PcItem("breakpoint_update -d",
		readline.PcItem("-n"),
		readline.PcItem("-h"),
		readline.PcItem("-o >="),
		readline.PcItem("-o =="),
		readline.PcItem("-o %"),
		readline.PcItem("-s enabled"),
		readline.PcItem("-s disabled"),
	),

	readline.PcItem("context_get",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
	),
	readline.PcItem("context_names"),

	readline.PcItem("eval",
		readline.PcItem("-p"),
		readline.PcItem("--"),
	),
	readline.PcItem("feature_get -n",
		readline.PcItem("breakpoint_languages"),
		readline.PcItem("breakpoint_types"),
		readline.PcItem("data_encoding"),
		readline.PcItem("encoding"),
		readline.PcItem("extended_properties"),
		readline.PcItem("language_name"),
		readline.PcItem("language_supports_threads"),
		readline.PcItem("language_version"),
		readline.PcItem("max_children"),
		readline.PcItem("max_data"),
		readline.PcItem("max_depth"),
		readline.PcItem("notify_ok"),
		readline.PcItem("protocol_version"),
		readline.PcItem("resolved_breakpoints"),
		readline.PcItem("show_hidden"),
		readline.PcItem("supported_encodings"),
		readline.PcItem("supports_async"),
		readline.PcItem("supports_postmortem"),
	),
	readline.PcItem("feature_set -n",
		readline.PcItem("encoding -v"),
		readline.PcItem("extended_properties -v"),
		readline.PcItem("max_children -v"),
		readline.PcItem("max_data -v"),
		readline.PcItem("max_depth -v"),
		readline.PcItem("notify_ok -v"),
		readline.PcItem("resolved_breakpoints -v"),
		readline.PcItem("show_hidden -v"),
	),

	readline.PcItem("typemap_get"),
	readline.PcItem("property_get",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
		readline.PcItem("-m"),
		readline.PcItem("-n"),
		readline.PcItem("-p"),
	),
	readline.PcItem("property_set",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
		readline.PcItem("-n"),
		readline.PcItem("-p"),
		readline.PcItem("--"),
	),
	readline.PcItem("property_value",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
		readline.PcItem("-n"),
		readline.PcItem("-p"),
	),

	readline.PcItem("source",
		readline.PcItem("-f"),
		readline.PcItem("-b"),
		readline.PcItem("-e"),
	),
	readline.PcItem("stack_depth"),
	readline.PcItem("stack_get",
		readline.PcItem("-d"),
	),
	readline.PcItem("status"),

	readline.PcItem("stderr"),
	readline.PcItem("stdout -c",
		readline.PcItem("0"),
		readline.PcItem("1"),
	),

	readline.PcItem("run"),
	readline.PcItem("step_into"),
	readline.PcItem("step_out"),
	readline.PcItem("step_over"),

	readline.PcItem("stop"),
	readline.PcItem("detach"),

	readline.PcItem("help"),
)

func initReadline() *readline.Instance {
	usr, _ := user.Current()
	dir := usr.HomeDir

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s", Bold("(cmd) ")),
		Stdout:          output,
		HistoryFile:     dir + "/.xdebug-debugclient.hist",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
	})
	if err != nil {
		panic(err)
	}

	return rl
}
