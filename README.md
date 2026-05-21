# ResolumeOSCStack

An OSC server for Resolume Arena 7 that enables triggering columns with custom transition times via OSC messages.

## Features

- Trigger Resolume columns with custom fade/transition times
- Automatic transition duration configuration for all layers
- QLab OSC proxy support (optional)
- WebSocket connection to Resolume with auto-reconnect

## Installation

Download the appropriate binary for your platform (Linux AMD64 or Windows AMD64).

The `config.json` file must be in the same directory as the executable.

## Configuration

Create a `config.json` file:

```json
{
    "osc_listen_port": 7000,
    "resolume": {
        "ip": "127.0.0.1",
        "websocket_port": 8080
    },
    "qlab": {
        "ip": "192.168.1.100",
        "osc_port": 53000
    }
}
```

### Configuration Options

- `osc_listen_port`: Port to listen for incoming OSC messages
- `resolume.ip`: IP address of the Resolume Arena instance
- `resolume.websocket_port`: WebSocket port (default: 8080)
- `qlab` (optional): QLab proxy configuration
  - `ip`: IP address of QLab instance
  - `osc_port`: OSC port for QLab

## Usage

Run the binary:

```bash
./ResolumeOSCStack
```

### OSC Commands

#### Trigger Column by Name

```
/column/{column_name} {transition_time}
```

**Parameters:**
- `column_name`: Column name (case-insensitive, supports spaces)
- `transition_time`: Transition duration in seconds (float or int)

**Examples:**
```
/column/intro 2.5           # Trigger column named "intro" with 2.5 second transition
/column/Verse 1 1.0         # Trigger column named "Verse 1" with 1.0 second transition
/column/MAIN BRIDGE 0.5     # Trigger column named "Main Bridge" (case-insensitive)
```

#### Trigger Column by Index

```
/column/index/{column_index} {transition_time}
```

**Parameters:**
- `column_index`: Column index (0-based)
- `transition_time`: Transition duration in seconds (float or int)

**Examples:**
```
/column/index/0 2.5    # Trigger first column with 2.5 second transition
/column/index/2 1.0    # Trigger third column with 1.0 second transition
```

### How It Works

1. The server connects to Resolume via WebSocket and receives composition data
2. Column names are tracked and updates are subscribed to automatically
3. When an OSC message is received:
   - Column name is matched (case-insensitive, first match for duplicates)
   - Transition duration is set for all layers in the composition
   - The column at the matched index is triggered
4. Resolume performs the transition with the specified duration

## QLab Integration

If QLab configuration is provided, the server will proxy all remaining OSC messages to the configured QLab instance. This allows the server to forward QLab control messages while handling Resolume-specific commands.

## Building from Source

```bash
go build
```