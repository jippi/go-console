
<div align="center">
  <a href="https://github.com/reeflective/console">
    <img alt="" src="" width="600">
  </a>
  <br> <h1> Console </h1>

  <p>  Closed-loop application library for Cobra commands  </p>
</div>


<!-- Badges -->
<p align="center">
  <a href="https://github.com/reeflective/console/actions/workflows/go.yml">
    <img src="https://github.com/reeflective/console/actions/workflows/go.yml/badge.svg?branch=main"
      alt="Github Actions (workflows)" />
  </a>

  <a href="https://github.com/reeflective/console">
    <img src="https://img.shields.io/github/go-mod/go-version/reeflective/console.svg"
      alt="Go module version" />
  </a>

  <a href="https://godoc.org/reeflective/go/console">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg"
      alt="GoDoc reference" />
  </a>

  <a href="https://goreportcard.com/report/github.com/reeflective/console">
    <img src="https://goreportcard.com/badge/github.com/reeflective/console"
      alt="Go Report Card" />
  </a>

  <a href="https://codecov.io/gh/reeflective/console">
    <img src="https://codecov.io/gh/reeflective/console/branch/main/graph/badge.svg"
      alt="codecov" />
  </a>

  <a href="https://opensource.org/licenses/BSD-3-Clause">
    <img src="https://img.shields.io/badge/License-BSD_3--Clause-blue.svg"
      alt="License: BSD-3" />
  </a>
</p>

Console is an all-in-one console application library built on top of a [readline](https://github.com/reeflective/readline) shell and using [Cobra](https://github.com/spf13/cobra) commands. 
It aims to provide users with a modern interface at at minimal cost while allowing them to focus on developing 
their commands and application core: the console will then transparently interface with these commands, and provide
the various features below almost for free.


## Features

### Menus & Commands 
- Declare & use multiple menus with their own command tree, prompt engines and special handlers.
- Bind cobra commands to provide the core functionality (see documentation for binding usage).
- Virtually all cobra settings can be modified, set and used freely, like in normal CLI workflows.
- Ability to bind handlers to special interrupt errors (eg. CtrlC/CtrlD), per menu.

### Shell interface
- Shell interface is powered by a [readline](https://github.com/reeflective/readline) instance.
- All features of readline are supported in the console. It also allows the console to give:
- Configurable bind keymaps, with live reload and sane defaults, and system-wide configuration.
- Out-of-the-box, advanced completions for commands, flags, positional and flag arguments.
- Provided by readline and [carapace](https://github.com/rsteube/carapace): automatic usage & validation command/flags/args hints.
- Syntax highlighting for commands (might be extended in the future).

### Other features 
- Support for an arbitrary number of history sources, per menu.
- Support for [oh-my-posh](https://github.com/JanDeDobbeleer/oh-my-posh) prompts, per menu and with custom configuration files for each.
- Also with oh-my-posh, ability to write and bind application-specific prompt segments.

<!-- ![readme-main-gif](https://github.com/maxlandon/gonsole/blob/assets/readme-main.gif) -->


## Documentation Contents

You can install and use the [example application console](https://github.com/reeflective/console/tree/main/example). This example application 
will give you a taste of the behavior and supported features. Additionally, the following 
documentation is available in the [wiki](https://github.com/reeflective/console/wiki):

* [Getting started](https://github.com/reeflective/console/Getting-Started) 
* [Shell readline](https://github.com/reeflective/console/wiki/Readline-Shell)
* [Menus](https://github.com/reeflective/console/wiki/Menus)
* [Prompts](https://github.com/reeflective/console/wiki/Prompts)
* [Binding commands](https://github.com/reeflective/console/Binding-Commands)
* [History Sources](https://github.com/reeflective/console/wiki/History-Sources)
* [Default commands](https://github.com/maxlandon/gonsole/wiki/Default-Commands)
* [Logging](https://github.com/reeflective/console/wiki/Logging)


## Status


