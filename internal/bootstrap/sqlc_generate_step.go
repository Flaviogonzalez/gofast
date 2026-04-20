package bootstrap

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

type SQLCGenerateStep struct {
	WorkingDir string
}

func (s *SQLCGenerateStep) Name() string { return "Run sqlc generate" }

func (s *SQLCGenerateStep) Run(pctx *ProjectContext) error {
	if len(pctx.Tables) == 0 {
		log.Println("  ⚠ no tables to generate — skipping sqlc")
		return nil
	}

	if _, err := exec.LookPath("sqlc"); err != nil {
		return fmt.Errorf("sqlc CLI not found in PATH — install it from https://docs.sqlc.dev/en/latest/overview/install.html")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(pctx.Ctx, "cmd", "/C", "sqlc", "generate")
	} else {
		cmd = exec.CommandContext(pctx.Ctx, "sqlc", "generate")
	}

	cmd.Dir = pctx.WorkingDir
	log.Println("current working dir", cmd.Dir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sqlc generate failed: %v\nOutput: %s", err, string(output))
	}

	log.Println("sqlc generate completed successfully")
	if len(output) > 0 {
		log.Println("Output:", string(output))
	}

	return nil
}
