# 🛠️ Debugging

Using a debugger allows you to trace the execution flow of an application in real-time from the command line. It also enables you to pause execution at specific points to analyze the system state in detail.

## Getting Started

To start the debugger, use the `start` command with the `--debug` flag. This command activates the debugger and allows you to control the application's execution through the command-line interface.

```sh
./dist/uniflow start --debug
```

## Commands

While the debugger is running, you can use various commands to efficiently perform debugging tasks. Below is a list of commonly used commands and their usage.

### Quit

Exits the debugging session and returns to the command line.

```sh
(debug) quit
```

This command can also be shortened to `q`.

### Break

Sets breakpoints to stop code execution at specific points, allowing you to inspect the system state and diagnose issues.

```sh
(debug) break                 # Set a breakpoint on all symbols
(debug) break <symbol>        # Set a breakpoint on a specific symbol
(debug) break <symbol> <port> # Set a breakpoint on a specific port of a symbol
```

This command can also be shortened to `b`.

### Continue

Resumes execution from a breakpoint. The program will continue running until it reaches the next breakpoint.

```sh
(debug) continue
```

This command can also be shortened to `c`.

### Delete

Deletes set breakpoints. Each breakpoint has a unique ID, which you can use to delete a specific breakpoint.

```sh
(debug) delete              # Delete the current breakpoint
(debug) delete <breakpoint> # Delete a specific breakpoint
```

This command can also be shortened to `d`.

### Breakpoints

Lists all currently set breakpoints, showing their locations and statuses.

```sh
(debug) breakpoints
```

This command can also be shortened to `bps`.

### Breakpoint

Views detailed information about a specific breakpoint. You can query the status of a breakpoint using its ID.

```sh
(debug) breakpoint              # View details of the current breakpoint
(debug) breakpoint <breakpoint> # View details of a specific breakpoint
```

This command can also be shortened to `bp`.

### Symbols

Lists all currently available symbols. Symbols represent nodes running at runtime.

```sh
(debug) symbols
```

This command can also be shortened to `sbs`.

### Symbol

Views detailed information about a specific symbol. Enter the symbol's ID or name to check its status and related information.

```sh
(debug) symbol              # View details of the current symbol
(debug) symbol <symbol>     # View details of a specific symbol
```

This command can also be shortened to `sb`.

### Processes

Lists all currently running processes, showing their status.

```sh
(debug) processes
```

This command can also be shortened to `procs`.

### Process

Views detailed information about a specific process. Use the process's ID to check its status and related information.

```sh
(debug) process              # View details of the current process
(debug) process <process>    # View details of a specific process
```

This command can also be shortened to `proc`.

### Frames

Lists all currently active frames. Frames represent a specific execution state of the code.

```sh
(debug) frames              # View frame information of the current process
(debug) frames <process>    # View frame information of a specific process
```

This command can also be shortened to `frms`.

### Frame

Views detailed information about the current frame, allowing you to closely inspect its state.

```sh
(debug) frame
```

This command can also be shortened to `frm`.
