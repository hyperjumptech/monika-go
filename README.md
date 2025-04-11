# Monika Go

Monika is a monitoring tool that runs probes and sends notifications when a probe fails.

## Features

- HTTP probes
- Ping probes
- Alerting system
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
  - `recoveryThreshold`: The number of times the probe should recover before marking it as an incident. By default, it will use the largest value of `recoveryThreshold` from all requests.
  - `incidentThreshold`: The number of times the probe should fail before marking it as an incident. By default, it will use the largest value of `incidentThreshold` from all requests.
  - `alerts`: An array of alerts to be evaluated for the probe. (More details below)
    - `query`: The query to evaluate.
    - `message`: The message to send if the query evaluates to true.
- `ping`: Indicates that the probe is a Ping probe
  - `uri`: The URI to ping

### Alerts

Alerts allows you to define conditions for triggering alerts based on HTTP probe responses. Each alert has the following properties:

- `query`: The query to evaluate.
- `message`: The message to send if the query evaluates to true.

#### Alert Expression Syntax

Alerts in Monika are defined using expressions that evaluate to boolean values. These expressions are evaluated against the response data of your HTTP probes. If an expression evaluates to true, the alert is triggered.

**WARNING**: The expression language used in [Monika](https://github.com/hyperjumptech/monika) is not the same as the one used in [Monika GO](https://github.com/hyperjumptech/monika-go).

##### Available Response Data

| Variable           | Type   | Description                         |
| ------------------ | ------ | ----------------------------------- |
| `response.status`  | Number | HTTP status code of the response    |
| `response.time`    | Number | Response time in milliseconds       |
| `response.body`    | String | Response body as a string           |
| `response.headers` | Map    | Response headers as key-value pairs |
| `response.size`    | Number | Size of the response in bytes       |

##### Expression Examples

```yaml
# Alert when HTTP status code is not a success (not in the 200-299 range)
response.status < 200 || response.status > 299

# Alert when response time is greater than 2000 milliseconds
response.time > 2000

# Alert when response body contains the word "error"
contains(response.body, "error")

# Alert when a specific header is missing
response.headers["Content-Type"] == nil

# Alert when response size is too large
response.size > 1000000  # Over 1MB

# Combining multiple conditions
response.time > 1000 && (response.status != 200 || contains(response.body, "error"))
```

#### Operators and Functions

The expr language supports a variety of operators and functions:

##### Arithmetic Operators

- `+` (addition)
- `-` (subtraction)
- `*` (multiplication)
- `/` (division)
- `%` (modulo)

##### Comparison Operators

- `==` (equal)
- `!=` (not equal)
- `<` (less than)
- `>` (greater than)
- `<=` (less than or equal)
- `>=` (greater than or equal)

##### Logical Operators

- `&&` (logical AND)
- `||` (logical OR)
- `!` (logical NOT)

##### String Functions

- `contains(s, substr)`: Checks if string `s` contains substring `substr`
- `startsWith(s, prefix)`: Checks if string `s` starts with `prefix`
- `endsWith(s, suffix)`: Checks if string `s` ends with `suffix`
- `len(s)`: Returns the length of string `s`

##### Other Functions

- `len(array)`: Returns the length of an array
- `all(array, predicate)`: Returns true if all elements in the array satisfy the predicate
- `any(array, predicate)`: Returns true if any element in the array satisfies the predicate
- `filter(array, predicate)`: Returns a new array with elements that satisfy the predicate
- `map(array, function)`: Returns a new array with the results of applying the function to each element

### Notifications

Notifications are defined in the configuration file. Each notification has the following properties:

- `id`: A unique identifier for the notification.
- `type`: The type of notification to send.
- `data`: The data to send with the notification.
  - `url`: The webhook URL to send the notification to.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have any suggestions or improvements.

## License

This project is licensed under the MIT License.
