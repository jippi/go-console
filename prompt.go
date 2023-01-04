package console

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/reeflective/readline"
)

// Prompt wraps an oh-my-posh prompt engine, so as to be able
// to be configured/enhanced and used the same way oh-my-posh is.
// Some methods have been added for ordering the application to
// to recompute prompts, print logs in sync with them, etc.
type Prompt struct {
	*engine.Engine
	console *Console
}

// LoadConfig loads the prompt JSON configuration at the specified path.
// It returns an error if the file is not found, or could not be read.
func (p *Prompt) LoadConfig(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	flags := &platform.Flags{
		Shell:  shell.PLAIN,
		Config: path,
	}

	p.Engine = engine.New(flags)

	return nil
}

// bind reassigns the prompt printing functions to the shell helpers.
func (p *Prompt) bind(shell *readline.Instance) {
	shell.Prompt.Primary(p.PrintPrimary)
	shell.Prompt.Right(p.PrintRPrompt)

	secondary := func() string {
		return p.PrintExtraPrompt(engine.Secondary)
	}
	shell.Prompt.Secondary(secondary)

	transient := func() string {
		return p.PrintExtraPrompt(engine.Transient)
	}
	shell.Prompt.Transient(transient)

	shell.Prompt.Tooltip(p.PrintTooltip)
}

// LogTransient prints a string message (a log, or more broadly, an
// asynchronous event) without bothering the user, and by "pushing"
// the prompt below the message.
//
// If this function is called while a command is running, the console
// will simply print the log below the current line, and will not print
// the prompt. In any other case this function will work normally.
func (c *Console) LogTransient(msg string, args ...interface{}) {
	if c.isExecuting {
		fmt.Printf(msg, args...)
	} else {
		c.shell.LogTransient(msg, args...)
	}
}

// Log - A simple function to print a message and redisplay the prompt below it.
// As with LogTransient, if this function is called while a command is running,
// the console will simply print the log below the current line, and will not
// print the prompt. In any other case this function will work normally.
func (c *Console) Log(msg string, args ...interface{}) {
	if c.isExecuting {
		fmt.Printf(msg, args...)
	} else {
		c.shell.Log(msg, args...)
	}
}

// makes a prompt engine with default/builtin configuration.
func newDefaultEngine() *engine.Engine {
	flags := &platform.Flags{
		Shell: shell.PLAIN,
	}

	return engine.New(flags)
}

func (c *Console) checkPrompts() {
	for _, menu := range c.menus {
		if menu.prompt.Engine == nil {
			menu.prompt.Engine = newDefaultEngine()
		}
	}
}
