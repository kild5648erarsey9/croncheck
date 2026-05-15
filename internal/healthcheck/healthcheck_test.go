package healthcheck_test

import (
	"testing"

	"github.com/croncheck/internal/healthcheck"
)

func TestNew_EmptyChecker(t *testing.T) {
	c := healthcheck.New()
	status := c.Run()

	if !status.Healthy {
		t.Error("expected healthy with no checks registered")
	}
	if len(status.Checks) != 0 {
		t.Errorf("expected 0 checks, got %d", len(status.Checks))
	}
	if status.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestChecker_AllPassing(t *testing.T) {
	c := healthcheck.New()
	c.Register("registry", func() healthcheck.Check {
		return healthcheck.Check{OK: true}
	})
	c.Register("scheduler", func() healthcheck.Check {
		return healthcheck.Check{OK: true, Message: "running"}
	})

	status := c.Run()

	if !status.Healthy {
		t.Error("expected healthy when all checks pass")
	}
	if len(status.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(status.Checks))
	}
}

func TestChecker_OneFailing(t *testing.T) {
	c := healthcheck.New()
	c.Register("ok_check", func() healthcheck.Check {
		return healthcheck.Check{OK: true}
	})
	c.Register("bad_check", func() healthcheck.Check {
		return healthcheck.Check{OK: false, Message: "unreachable"}
	})

	status := c.Run()

	if status.Healthy {
		t.Error("expected unhealthy when a check fails")
	}
	if got := status.Checks["bad_check"]; got.OK {
		t.Error("expected bad_check to be not OK")
	}
	if got := status.Checks["bad_check"]; got.Message != "unreachable" {
		t.Errorf("unexpected message: %q", got.Message)
	}
}

func TestChecker_Register_Overwrite(t *testing.T) {
	c := healthcheck.New()
	c.Register("probe", func() healthcheck.Check {
		return healthcheck.Check{OK: false}
	})
	c.Register("probe", func() healthcheck.Check {
		return healthcheck.Check{OK: true}
	})

	status := c.Run()

	if !status.Healthy {
		t.Error("expected healthy after overwriting check with passing one")
	}
}
