// Command setup is an interactive CLI that generates config.yaml for the application.
// Run it with: make setup
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
)

// configData holds all values collected from the user during setup.
type configData struct {
	App      appData
	Database databaseData
	Pipefy   pipefyData
	Logger   loggerData
}

type appData struct {
	Name        string
	Port        int
	Environment string
}

type databaseData struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

type pipefyData struct {
	Token    string
	PipeID   string
	Simulate bool
}

type loggerData struct {
	Level  string
	Format string
}

// configTemplate mirrors config.yaml.example with comments preserved.
const configTemplate = `app:
  name: {{ .App.Name }}
  port: {{ .App.Port }}
  # Accepted values: dev | staging | prod
  # dev:     fake pipefy adapter, no token required
  # staging: real adapter with simulate=true, token required
  # prod:    real adapter, token required, mutations are sent
  environment: {{ .App.Environment }}

database:
  host: {{ .Database.Host }}
  port: {{ .Database.Port }}
  name: {{ .Database.Name }}
  user: {{ .Database.User }}
  password: {{ .Database.Password }}
  sslmode: {{ .Database.SSLMode }}

pipefy:
  # Personal Access Token from https://app.pipefy.com/tokens
  # Required for staging and prod. Leave empty for dev.
  token: "{{ .Pipefy.Token }}"
  # Target pipe ID where cards will be created
  pipe_id: "{{ .Pipefy.PipeID }}"
  # simulate: true  -> builds and logs the GraphQL payload but does not send the HTTP request
  # simulate: false -> sends the request to Pipefy API (prod behavior)
  # Only meaningful in staging. In dev, the fake adapter is used regardless.
  simulate: {{ .Pipefy.Simulate }}

logger:
  # Accepted values: debug | info | warn | error
  level: {{ .Logger.Level }}
  # Accepted values: json | console
  format: {{ .Logger.Format }}
`

func main() {
	fmt.Println("\npipefy-integration setup")
	fmt.Println("------------------------")

	if _, err := os.Stat("config.yaml"); err == nil {
		fmt.Print("\nconfig.yaml already exists. Overwrite? (y/n) [n]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("Setup cancelled. Existing config.yaml was not modified.")
			os.Exit(0)
		}
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)

	// --- App ---
	fmt.Println("[app]")
	appName := prompt(reader, "Name", "pipefy-integration")
	appPort := promptInt(reader, "Port", 8080)
	environment := promptEnvironment(reader)

	// --- Database ---
	fmt.Println("\n[database]")
	dbHost := prompt(reader, "Host", "localhost")
	dbPort := promptInt(reader, "Port", 5433)
	dbName := prompt(reader, "Name", "pipefy_integration")
	dbUser := prompt(reader, "User", "pipefy")
	dbPass := prompt(reader, "Password", "pipefy")
	dbSSL := promptSSLMode(reader)

	// --- Pipefy ---
	fmt.Println("\n[pipefy]")
	var pipefyToken, pipefyPipeID string
	pipefySimulate := false

	switch environment {
	case "staging", "prod":
		pipefyToken = promptRequired(reader, "Token (from https://app.pipefy.com/tokens)")
		pipefyPipeID = promptRequired(reader, "Pipe ID")
		if environment == "staging" {
			pipefySimulate = promptBool(reader, "Simulate calls (log only, do not send to Pipefy)", true)
		}
	default:
		fmt.Println("Skipped: environment is dev, fake adapter will be used.")
	}

	// --- Logger ---
	fmt.Println("\n[logger]")
	logLevel := promptLogLevel(reader)
	logFormat := promptLogFormat(reader)

	data := configData{
		App: appData{
			Name:        appName,
			Port:        appPort,
			Environment: environment,
		},
		Database: databaseData{
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
			User:     dbUser,
			Password: dbPass,
			SSLMode:  dbSSL,
		},
		Pipefy: pipefyData{
			Token:    pipefyToken,
			PipeID:   pipefyPipeID,
			Simulate: pipefySimulate,
		},
		Logger: loggerData{
			Level:  logLevel,
			Format: logFormat,
		},
	}

	if err := writeConfig(data); err != nil {
		fmt.Fprintf(os.Stderr, "\nerror writing config.yaml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nconfig.yaml created.")
	fmt.Println("Next steps:")
	fmt.Println("  make docker-up   start PostgreSQL")
	fmt.Println("  make run         start the API")
}

// writeConfig renders the config template and writes it to config.yaml.
func writeConfig(data configData) error {
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// 0600: owner read/write only — config contains secrets
	f, err := os.OpenFile("config.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// prompt prints a question with an optional default value and reads the user's input.
// If the input is empty, the default value is returned.
func prompt(reader *bufio.Reader, question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", question, defaultVal)
	} else {
		fmt.Printf("  %s: ", question)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

// promptRequired loops until the user provides a non-empty value.
func promptRequired(reader *bufio.Reader, question string) string {
	for {
		value := prompt(reader, question, "")
		if value != "" {
			return value
		}
		fmt.Printf("  %s is required.\n", question)
	}
}

// promptInt reads an integer input, returning the default on empty or invalid input.
func promptInt(reader *bufio.Reader, question string, defaultVal int) int {
	raw := prompt(reader, question, strconv.Itoa(defaultVal))
	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return val
}

// promptBool reads a y/n input.
func promptBool(reader *bufio.Reader, question string, defaultVal bool) bool {
	defaultStr := "n"
	if defaultVal {
		defaultStr = "y"
	}
	answer := prompt(reader, question+" (y/n)", defaultStr)
	return strings.ToLower(strings.TrimSpace(answer)) == "y"
}

// promptEnvironment loops until a valid environment is provided.
func promptEnvironment(reader *bufio.Reader) string {
	for {
		value := prompt(reader, "Environment (dev/staging/prod)", "dev")
		switch value {
		case "dev", "staging", "prod":
			return value
		default:
			fmt.Println("  Invalid value. Choose: dev, staging or prod.")
		}
	}
}

// promptSSLMode loops until a valid sslmode is provided.
func promptSSLMode(reader *bufio.Reader) string {
	for {
		value := prompt(reader, "SSL mode (disable/require/verify-full)", "disable")
		switch value {
		case "disable", "require", "verify-full":
			return value
		default:
			fmt.Println("  Invalid value. Choose: disable, require or verify-full.")
		}
	}
}

// promptLogLevel loops until a valid log level is provided.
func promptLogLevel(reader *bufio.Reader) string {
	for {
		value := prompt(reader, "Log level (debug/info/warn/error)", "info")
		switch value {
		case "debug", "info", "warn", "error":
			return value
		default:
			fmt.Println("  Invalid value. Choose: debug, info, warn or error.")
		}
	}
}

// promptLogFormat loops until a valid log format is provided.
func promptLogFormat(reader *bufio.Reader) string {
	for {
		value := prompt(reader, "Log format (json/console)", "json")
		switch value {
		case "json", "console":
			return value
		default:
			fmt.Println("  Invalid value. Choose: json or console.")
		}
	}
}
