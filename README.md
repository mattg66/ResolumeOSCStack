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

#### Trigger Column with Transition

```
/column/{column_index} {transition_time}
```

**Parameters:**
- `column_index`: Column number (1-based index)
- `transition_time`: Transition duration in seconds (float or int)

**Examples:**
```
/column/1 2.5    # Trigger column 1 with 2.5 second transition
/column/3 1.0    # Trigger column 3 with 1.0 second transition
/column/5 0.5    # Trigger column 5 with 0.5 second transition
```

### How It Works

1. The server receives an OSC message with the column ID and transition time
2. It sets the transition duration parameter for all layers in the composition
3. It triggers the specified column to connect
4. Resolume performs the transition with the specified duration

## QLab Integration

If QLab configuration is provided, the server will proxy all remaining OSC messages to the configured QLab instance. This allows the server to forward QLab control messages while handling Resolume-specific commands.

## Building from Source

```bash
go build
```