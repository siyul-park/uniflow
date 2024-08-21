# 🛠️ Debugging

Using the debugger, you can track the execution flow of your application in real-time from the command line, pausing at specific points to analyze the state in detail.

## Getting Started

To start the debugger, use the `start` command with the `--debug` flag. This will activate the debugger, allowing you to precisely control the application's execution via the command-line interface.

```sh
./dist/uniflow start --debug
```

## Key Commands

When the debugger is running, you can use various commands to efficiently carry out debugging tasks. Below are some key commands and their usage during the debugging process.

### Quit

Use this command to end the debugging session. To exit the session and return to the command line, enter the following:

```sh
(debug) quit
```

Alternatively, you can use the `q` command to exit the session.

### Break

Set a breakpoint to stop code execution at a specific point. This allows you to analyze the state of the program mid-execution and identify issues.

```sh
(debug) break               # Set a breakpoint on all symbols
(debug) break <symbol>      # Set a breakpoint on a specific symbol
(debug) break <symbol> <port>  # Set a breakpoint on a specific port of a symbol
```

You can also use the `b` command for the same purpose.

### Continue

Resume execution from the current breakpoint. This command allows the program to run until it hits the next breakpoint.

```sh
(debug) continue
```

This command can also be executed using the `c` command.

### Delete

Remove a set breakpoint. Each breakpoint has a unique ID, which can be used to delete a specific breakpoint.

```sh
(debug) delete  # Delete the current breakpoint
(debug) delete <breakpoint>  # Delete a specific breakpoint
```

This command can also be executed using the `d` command.

### Breakpoints

List all currently set breakpoints. This command allows you to view the location and status of each breakpoint.

```sh
(debug) breakpoints
```

You can also use the `bps` command to achieve the same result.

### Breakpoint

Display detailed information about a specific breakpoint. Use the breakpoint's ID to check its status and other details.

```sh
(debug) breakpoint  # View details of the current breakpoint
(debug) breakpoint <breakpoint>  # View details of a specific breakpoint
```

This command can also be used as `bp`.

### Symbols

List all available symbols. Symbols represent the nodes currently being executed at runtime. This command helps you view the complete list of symbols.

```sh
(debug) symbols
```

This command can also be executed using `sbs`.

### Symbol

Display detailed information about a specific symbol. Enter the symbol's ID or name to view its status and related information.

```sh
(debug) symbol  # View details of the current symbol
(debug) symbol <symbol>  # View details of a specific symbol
```

This command can also be executed using `sb`.

### Processes

List all currently running processes. This command allows you to check the status of the processes active in the system.

```sh
(debug) processes
```

You can also use the `procs` command for the same purpose.

### Process

Display detailed information about a specific process. Enter the process ID to view its status and related details.

```sh
(debug) process  # View details of the current process
(debug) process <process>  # View details of a specific process
```

This command can also be executed using `proc`.

### Frame

Display detailed information about the current frame. A frame represents a specific execution state in the code, and this command allows you to examine the current frame in detail.

```sh
(debug) frame
```

This command can also be used as `fm`.