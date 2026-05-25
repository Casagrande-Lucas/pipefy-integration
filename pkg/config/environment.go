package config

// Environment represents the application runtime environment.
type Environment string

const (
	Dev     Environment = "dev"
	Staging Environment = "staging"
	Prod    Environment = "prod"
)

// IsValid reports whether the environment value is recognized.
func (e Environment) IsValid() bool {
	switch e {
	case Dev, Staging, Prod:
		return true
	}
	return false
}

// RequiresToken reports whether a Pipefy token is mandatory in this environment.
// dev does not require a token; staging and prod do.
func (e Environment) RequiresToken() bool {
	return e == Staging || e == Prod
}

// UsesFakeAdapter reports whether the fake Pipefy adapter should be used.
// In dev, no real HTTP calls are made and no token is needed.
func (e Environment) UsesFakeAdapter() bool {
	return e == Dev
}

// AllowsSimulate reports whether the simulate flag is meaningful.
// In prod, simulate must always be false.
func (e Environment) AllowsSimulate() bool {
	return e != Prod
}
