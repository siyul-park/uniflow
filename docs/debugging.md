# 🛠️ Debugging

A debugger lets you track and control an application's execution in real-time from the command line, allowing you to pause execution at specific points for detailed inspection.

## Getting Started

To start the debugger, use the `start` command with the `--debug` flag. This activates the debugger and provides an interface for managing the application's execution from the command line.

```sh
./dist/uniflow start --debug
```

## Commands

When the debugger is active, you can use the following commands for various debugging tasks:

### Quit

Ends the debugging session. Use this command to exit the debugger.

```sh
(debug) quit
```

You can also type `q` to quit.

### Break

Sets a breakpoint to pause execution at a specific point in the code. Breakpoints let you stop the program at chosen locations for analysis.

```sh
(debug) break               # Set a breakpoint at all symbols
(debug) break <symbol>      # Set a breakpoint at a specific symbol
(debug) break <symbol> <port>  # Set a breakpoint at a specific symbol and port
```

The `b` command can also be used for this.

### Continue

Resumes execution from the current breakpoint. This command continues running the program until the next breakpoint is hit.

```sh
(debug) continue
```

The `c` command also resumes execution.

### Delete

Removes a breakpoint. Each breakpoint has a unique ID, which you can use to specify which one to delete.

```sh
(debug) delete  # Delete the current breakpoint
(debug) delete <breakpoint>  # Delete a specific breakpoint
```

You can also use the `d` command to delete breakpoints.

### Breakpoints

Lists all the active breakpoints. This command shows the locations and statuses of all breakpoints.

```sh
(debug) breakpoints
```

You can use `bps` to get the same list.

### Breakpoint

Displays details about a specific breakpoint. Use the breakpoint's ID to view its status.

```sh
(debug) breakpoint # Show details of the current breakpoint
(debug) breakpoint <breakpoint>  # Show details of a specific breakpoint
```

The `bp` command also provides this information.

### Symbols

Lists all the available symbols. Symbols represent nodes in the runtime. This command shows a list of all symbols.

```sh
(debug) symbols
```

You can use `sbs` to see the same list.

### Symbol

Shows details about a specific symbol. Enter the ID or name of the symbol to view its status.

```sh
(debug) symbol # Show details of the current symbol
(debug) symbol <symbol> # Show details of a specific symbol
```

The `sb` command provides the same details.

### Frame

Displays details of the current frame. Frames represent specific execution states. Use this command to inspect the state of the current frame.

```sh
(debug) frame
```

You can also use `fm` to get this information.
