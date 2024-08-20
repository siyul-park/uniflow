# 🛠️ Debugging

A debugger allows you to track the execution flow of your application in real-time from the command line. You can pause execution at specific points to analyze the state of the application in detail.

## Getting Started

To start the debugger, use the `start` command with the `--debug` flag. This command will activate the debugger, providing an interface in the command line to control the execution of your application.

```sh
./dist/uniflow start --debug
```

## Key Commands

While the debugger is running, you can use a variety of commands to efficiently debug your application. Below are the key commands and their usage.

### Quit

Ends the debugging session. To terminate the session and return to the command line, use the following command:

```sh
(debug) quit
```

Alternatively, you can use the `q` command to exit the session.

### Break

Sets a breakpoint to pause code execution at a specific point. Breakpoints allow you to stop the execution and analyze the state of the application at that moment.

```sh
(debug) break               # Set a breakpoint at all symbols
(debug) break <symbol>      # Set a breakpoint at a specific symbol
(debug) break <symbol> <port>  # Set a breakpoint at a specific port of a symbol
```

The `b` command can also be used to achieve the same result.

### Continue

Resumes execution from the current breakpoint. This command will continue running the program until the next breakpoint is reached.

```sh
(debug) continue
```

The `c` command can also be used to resume execution.

### Delete

Deletes a breakpoint. Each breakpoint has a unique ID, which can be used to specify which breakpoint to delete.

```sh
(debug) delete  # Delete the current breakpoint
(debug) delete <breakpoint>  # Delete a specific breakpoint
```

You can also use the `d` command to perform the same action.

### Breakpoints

Lists all currently set breakpoints. This command allows you to view the location and status of each breakpoint.

```sh
(debug) breakpoints
```

The `bps` command will produce the same result.

### Breakpoint

Displays detailed information about a specific breakpoint. Use the breakpoint's ID to check its status.

```sh
(debug) breakpoint # View details of the current breakpoint
(debug) breakpoint <breakpoint>  # View details of a specific breakpoint
```

The `bp` command can be used as an alternative.

### Symbols

Lists all available symbols. Symbols refer to the nodes currently running in the application. Use this command to view a list of all symbols.

```sh
(debug) symbols
```

You can also use the `sbs` command to achieve the same result.

### Symbol

Displays detailed information about a specific symbol. Use the symbol's ID or name to view its status and related information.

```sh
(debug) symbol # View details of the current symbol
(debug) symbol <symbol> # View details of a specific symbol
```

The `sb` command can also be used to obtain the same information.

### Processes

Lists all processes currently running. This command lets you see the status of active processes in the system.

```sh
(debug) processes
```

The `procs` command will also display the same information.

### Process

Displays detailed information about a specific process. Enter the process's ID to view its status and related details.

```sh
(debug) process # View details of the current process
(debug) process <process> # View details of a specific process
```

This command can also be executed using `proc`.

### Frame

Displays detailed information about the current frame. A frame represents a specific execution state of the code, and this command allows you to inspect the current frame in detail.

```sh
(debug) frame
```

The `fm` command can be used as an alternative to achieve the same action.
