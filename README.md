# Monika Go

Monika is a monitoring tool that runs probes and sends notifications when a probe fails.

## Features

- HTTP probes
- Ping probes
- Notifications
  - Discord

## Prerequisites

1. Go v1.23 or higher

## Usage

```bash
# Clone the repository
git clone https://github.com/dennypradipta/monika-go.git

# Copy the example configuration file
cp monika.example.yml monika.yml

# Run Monika
make run

# Build Monika
make build

# Test Monika
make test
```

## Configuration

Monika uses a YAML configuration file to define probes and notifications. An example configuration file is provided in the `monika.yml` file.

To run Monika with a custom configuration file, run the following command:

```bash
./monika -c monika.yml
```

### Probes

Probes are defined in the configuration file. Each probe has the following properties:

- `id`: A unique identifier for the probe.
- `name`: A name for the probe.
- `interval`: The interval in seconds between probes.
- `requests`: An array of requests to be made by the probe.
  - `timeout`: The timeout in milliseconds for the request.
  - `method`: The HTTP method to use for the request.
  - `url`: The URL to make the request to.
  - `recoveryThreshold`: The number of times the probe should recover before marking it as an incident.
  - `incidentThreshold`: The number of times the probe should fail before marking it as an incident.
- `socket`: An optional socket to use for the probe.
  - `host`: The host to use for the socket.

### Notifications

Notifications are defined in the configuration file. Each notification has the following properties:

- `id`: A unique identifier for the notification.
- `type`: The type of notification to send.
- `data`: The data to send with the notification.
  - `url`: The URL to send the notification to.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have any suggestions or improvements.

## License

This project is licensed under the MIT License.
