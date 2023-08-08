package readline

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/reeflective/readline"
	"github.com/reeflective/readline/inputrc"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

// Bind returns a command named `bind`, for manipulating readline keymaps and bindings.
func Bind(shell *readline.Shell) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Display or modify readline key bindings",
		Long: `Manipulate readline keymaps and bindings.

Changing binds:
Note that the keymap name is optional, and if omitted, the default keymap is used.
The default keymap is 'vi' only if 'set editing-mode vi' is found in inputrc , and 
unless the -m option is used to set a different keymap.
Also, note that the bind [seq] [command] slightly differs from the original bash 'bind' command.

Exporting binds:
- Since all applications always look up to the same file for a given user,
  the export command does not allow to write and modify this file itself.
- Also, since saving the entire list of options and bindings in a different
  file for each application would also defeat the purpose of .inputrc.`,
		Example: `Changing binds:
    bind "\C-x\C-r": re-read-init-file          # C-x C-r to reload the inputrc file, in the default keymap.
    bind -m vi-insert "\C-l" clear-screen       # C-l to clear-screen in vi-insert mode
    bind -m menu-complete '\C-n' menu-complete  # C-n to cycle through choices in the completion keymap.

Exporting binds:
   bind --binds-rc --lib --changed # Only changed options/binds to stdout applying to all apps using this lib
   bind --app OtherApp -c          # Changed options, applying to an app other than our current shell one`,
	}

	// Flags
	cmd.Flags().StringP("keymap", "m", "", "Specify the keymap")
	cmd.Flags().BoolP("list", "l", false, "List names of functions")
	cmd.Flags().BoolP("binds", "P", false, "List function names and bindings")
	cmd.Flags().BoolP("binds-rc", "p", false, "List functions and bindings in a form that can be reused as input")
	cmd.Flags().BoolP("macros", "S", false, "List key sequences that invoke macros and their values")
	cmd.Flags().BoolP("macros-rc", "s", false, "List key sequences that invoke macros and their values in a form that can be reused as input")
	cmd.Flags().BoolP("vars", "V", false, "List variables names and values")
	cmd.Flags().BoolP("vars-rc", "v", false, "List variables names and values in a form that can be reused as input")
	cmd.Flags().StringP("query", "q", "", "Query about which keys invoke the named function")
	cmd.Flags().StringP("unbind", "u", "", "Unbind all keys which are bound to the named function")
	cmd.Flags().StringP("remove", "r", "", "Remove the bindings for KEYSEQ")
	cmd.Flags().StringP("file", "f", "", "Read key bindings from FILENAME")
	cmd.Flags().StringP("app", "A", "", "Optional application name (if empty/not used, the current app)")
	cmd.Flags().BoolP("changed", "c", false, "Only export options modified since app start: maybe not needed, since no use for it")
	cmd.Flags().BoolP("lib", "L", false, "Like 'app', but export options/binds for all apps using this specific library")

	// Run implementation
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Set keymap of interest
		keymap, _ := cmd.Flags().GetString("keymap")
		if keymap == "" {
			keymap = string(shell.Keymap.Main())
		}

		// If some flags were triggered, we don't require
		// positional args and don't have to always print usage.
		var executed bool

		// All flags and args that are "exiting the command
		// after run" are listed and evaluated first.

		// Function names
		if cmd.Flags().Changed("list") {
			for name := range shell.Keymap.Commands() {
				fmt.Println(name)
			}

			return nil
		}

		// 2 - Query binds for function
		if cmd.Flags().Changed("query") {
			bindsQuery(shell, cmd, keymap)
			return nil
		}

		// From this point on, some flags don't exit after printing
		// their respective listings, since we can combine and output
		// various types of stuff at once, for configs or display.
		//
		// We can even read a file for binds, remove some of them,
		// and display all or specific sections of our config in
		// a single call, with multiple flags of all sorts.

		// 1 - Apply any changes we want from a file first.
		if cmd.Flags().Changed("file") {
			if err := readFileConfig(shell, cmd, keymap); err != nil {
				return err
			}
		}

		// Remove anything we might have been asked to.
		if cmd.Flags().Changed("unbind") {
			executed = true
			unbindKeys(shell, cmd, keymap)
		}

		if cmd.Flags().Changed("remove") {
			executed = true
			removeCommands(shell, cmd, keymap)
		}

		// Then print out the sections of interest.
		// Start with Options
		if cmd.Flags().Changed("vars") {
			executed = true
			varsQuery(shell, cmd, keymap)
		}

		if cmd.Flags().Changed("vars-rc") {
			executed = true
			varsQueryRC(shell, cmd, keymap)
		}

		// Sequences to function names
		if cmd.Flags().Changed("binds") {
			executed = true
			fmt.Println()
			fmt.Printf("=== Binds (%s)===\n", shell.Keymap.Main())
			fmt.Println()

			shell.Keymap.PrintBinds(keymap, false)

			return nil
		}

		if cmd.Flags().Changed("binds-rc") {
			executed = true
			fmt.Println()
			fmt.Println("# Command binds (autogenerated from reeflective/readline)")

			shell.Keymap.PrintBinds(keymap, true)
		}

		// Macros
		if cmd.Flags().Changed("macros") {
			executed = true
			macrosQuery(shell, cmd, keymap)
		}

		if cmd.Flags().Changed("macros-rc") {
			executed = true
			macrosQueryRC(shell, cmd, keymap)
		}

		// The command has performed an action, so any binding
		// with positional arguments is not considered or evaluated.
		if executed {
			return nil
		}

		// Bind actions.
		// Some keymaps are aliases of others, so use either
		// all equivalents or fallback to the relevant keymap.
		if len(args) < 2 {
			return errors.New("Usage: bind [-m keymap] [keyseq] [command]")
		}

		// The key sequence is an escaped string, so unescape it.
		seq := inputrc.Unescape(args[0])

		var found bool

		for command := range shell.Keymap.Commands() {
			if command == args[1] {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Unknown command: %s", args[1])
		}

		// If the keymap doesn't exist, create it.
		if shell.Config.Binds[keymap] == nil {
			shell.Config.Binds[keymap] = make(map[string]inputrc.Bind)
		}

		// Adjust some keymaps (aliases of each other).
		bindkey := func(keymap string) {
			shell.Config.Binds[keymap][seq] = inputrc.Bind{Action: args[1]}
		}

		// (Bind the key sequence to the command)
		applyToKeymap(keymap, bindkey)

		return nil
	}

	// *** Completions ***
	comps := carapace.Gen(cmd)
	flagComps := make(carapace.ActionMap)

	// Flags
	flagComps["keymap"] = completeKeymaps(shell, cmd)
	flagComps["query"] = completeCommands(shell, cmd)
	flagComps["unbind"] = completeCommands(shell, cmd)
	flagComps["remove"] = completeBindSequences(shell, cmd)
	flagComps["file"] = carapace.ActionFiles()

	comps.FlagCompletion(flagComps)

	// Positional arguments
	comps.PositionalCompletion(
		carapace.ActionValues().Usage("sequence"),
		completeCommands(shell, cmd),
	)

	return cmd
}

func completeKeymaps(sh *readline.Shell, cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		results := make([]string, 0)

		for name := range sh.Config.Binds {
			results = append(results, name)
		}

		return carapace.ActionValues(results...).Tag("keymaps").Usage("keymap")
	})
}

func completeBindSequences(sh *readline.Shell, cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(ctx carapace.Context) carapace.Action {
		// Get the keymap.
		var keymap string

		if cmd.Flags().Changed("keymap") {
			keymap, _ = cmd.Flags().GetString("keymap")
		}

		if keymap == "" {
			keymap = string(sh.Keymap.Main())
		}

		// Get the binds.
		binds := sh.Config.Binds[keymap]
		if binds == nil {
			return carapace.ActionValues().Usage("sequence")
		}

		// Make a list of all sequences bound to each command, with descriptions.
		cmdBinds := make([]string, 0)
		insertBinds := make([]string, 0)

		for key, bind := range binds {
			if bind.Action == "self-insert" {
				insertBinds = append(insertBinds, "\""+inputrc.Escape(key)+"\"")
			} else {
				cmdBinds = append(cmdBinds, "\""+inputrc.Escape(key)+"\"")
				cmdBinds = append(cmdBinds, bind.Action)
			}
		}

		return carapace.Batch(
			carapace.ActionValues(insertBinds...).Tag(fmt.Sprintf("self-insert binds (%s)", keymap)).Usage("sequence"),
			carapace.ActionValuesDescribed(cmdBinds...).Tag(fmt.Sprintf("non-insert binds (%s)", keymap)).Usage("sequence"),
		).ToA()
	})
}

func completeCommands(sh *readline.Shell, cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		results := make([]string, 0)

		for name := range sh.Keymap.Commands() {
			results = append(results, name)
		}

		return carapace.ActionValues(results...).Tag("commands").Usage("command")
	})
}

func applyToKeymap(keymap string, bind func(keymap string)) {
	switch keymap {
	case "emacs", "emacs-standard":
		for _, km := range []string{"emacs", "emacs-standard"} {
			bind(km)
		}
	case "emacs-ctlx":
		for _, km := range []string{"emacs-ctlx", "emacs-standard", "emacs"} {
			bind(km)
		}
	case "emacs-meta":
		for _, km := range []string{"emacs-meta", "emacs-standard", "emacs"} {
			bind(km)
		}
	case "vi", "vi-move", "vi-command":
		for _, km := range []string{"vi", "vi-move", "vi-command"} {
			bind(km)
		}
	default:
		bind(keymap)
	}
}

func bindsQuery(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	binds := sh.Config.Binds[keymap]
	if binds == nil {
		return
	}

	command, _ := cmd.Flags().GetString("query")

	// Make a list of all sequences bound to each command.
	cmdBinds := make([]string, 0)

	for key, bind := range binds {
		if bind.Action != command {
			continue
		}

		cmdBinds = append(cmdBinds, inputrc.Escape(key))
	}

	sort.Strings(cmdBinds)

	switch {
	case len(cmdBinds) == 0:
	case len(cmdBinds) > 5:
		var firstBinds []string

		for i := 0; i < 5; i++ {
			firstBinds = append(firstBinds, "\""+cmdBinds[i]+"\"")
		}

		bindsStr := strings.Join(firstBinds, ", ")
		fmt.Printf("%s can be found on %s ...\n", command, bindsStr)

	default:
		var firstBinds []string

		for _, bind := range cmdBinds {
			firstBinds = append(firstBinds, "\""+bind+"\"")
		}

		bindsStr := strings.Join(firstBinds, ", ")
		fmt.Printf("%s can be found on %s\n", command, bindsStr)
	}
}

func varsQuery(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	var variables []string

	for variable := range sh.Config.Vars {
		variables = append(variables, variable)
	}

	sort.Strings(variables)

	fmt.Println()
	fmt.Println("=== Options ===")
	fmt.Println()

	for _, variable := range variables {
		value := sh.Config.Vars[variable]
		fmt.Printf("%s is set to `%v'\n", variable, value)
	}
}

func varsQueryRC(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	var variables []string

	for variable := range sh.Config.Vars {
		variables = append(variables, variable)
	}

	sort.Strings(variables)

	fmt.Println()
	fmt.Println("# Options (autogenerated from reeflective/readline)")

	for _, variable := range variables {
		value := sh.Config.Vars[variable]
		fmt.Printf("set %s %v\n", variable, value)
	}
}

func macrosQuery(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	binds := sh.Config.Binds[keymap]
	if len(binds) == 0 {
		return
	}

	var macroBinds []string

	for keys, bind := range binds {
		if bind.Macro {
			macroBinds = append(macroBinds, inputrc.Escape(keys))
		}
	}

	if len(macroBinds) == 0 {
		return
	}

	sort.Strings(macroBinds)

	fmt.Println()
	fmt.Printf("=== Macros (%s)===\n", sh.Keymap.Main())
	fmt.Println()

	for _, key := range macroBinds {
		action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
		fmt.Printf("%s outputs %s\n", key, action)
	}
}

func macrosQueryRC(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	binds := sh.Config.Binds[keymap]
	if len(binds) == 0 {
		return
	}
	var macroBinds []string

	for keys, bind := range binds {
		if bind.Macro {
			macroBinds = append(macroBinds, inputrc.Escape(keys))
		}
	}

	sort.Strings(macroBinds)

	fmt.Println()
	fmt.Println("# Macros (autogenerated from reeflective/readline)")

	for _, key := range macroBinds {
		action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
		fmt.Printf("\"%s\": \"%s\"\n", key, action)
	}
}

func unbindKeys(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	command, _ := cmd.Flags().GetString("unbind")

	unbind := func(keymap string) {
		binds := sh.Config.Binds[keymap]
		if binds == nil {
			return
		}

		cmdBinds := make([]string, 0)

		for key, bind := range binds {
			if bind.Action != command {
				continue
			}

			cmdBinds = append(cmdBinds, key)
		}

		for _, key := range cmdBinds {
			delete(binds, key)
		}
	}

	applyToKeymap(keymap, unbind)
}

func removeCommands(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	seq, _ := cmd.Flags().GetString("remove")

	removeBind := func(keymap string) {
		binds := sh.Config.Binds[keymap]
		if binds == nil {
			return
		}

		cmdBinds := make([]string, 0)

		for key := range binds {
			if key != seq {
				continue
			}

			cmdBinds = append(cmdBinds, key)
		}

		for _, key := range cmdBinds {
			delete(binds, key)
		}
	}

	applyToKeymap(keymap, removeBind)
}

func readFileConfig(sh *readline.Shell, cmd *cobra.Command, keymap string) error {
	fileF, _ := cmd.Flags().GetString("file")

	file, err := os.Stat(fileF)
	if err != nil {
		return err
	}

	if err = inputrc.ParseFile(file.Name(), sh.Config, sh.Opts...); err != nil {
		return err
	}

	fmt.Printf("Read and parsed %s\n", file.Name())

	return nil
}
