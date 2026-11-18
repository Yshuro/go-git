package git

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v6/internal/trace"
)

func TestMain(m *testing.M) {
	// Set the trace targets based on the environment variables.
	trace.ReadEnv()

	// Keep the attributes files of the host, which Git reads for every
	// repository, out of the tests.
	home, err := os.MkdirTemp("", "go-git-xdg")
	if err != nil {
		panic(err)
	}
	for k, v := range map[string]string{"XDG_CONFIG_HOME": home, "GIT_ATTR_NOSYSTEM": "1"} {
		if err := os.Setenv(k, v); err != nil {
			panic(err)
		}
	}
	// Register a default ConfigSource so tests that call ConfigScoped
	// (directly or indirectly via Commit/CreateTag) don't fail with
	// "no config loader registered".
	registerTestConfigLoader()

	// Run the tests.
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}
