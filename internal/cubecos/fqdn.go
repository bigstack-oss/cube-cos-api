package cubecos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	defErrors "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fqdn"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

// GetFqdn returns cubesys.fqdn, the name the web UIs are served on, or "" when none is set.
func GetFqdn() (string, error) {
	value, err := GetTuningValue(fqdn.Tuning)
	if errors.Is(err, defErrors.ErrTuningNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return value, nil
}

// runFqdnSdk runs hex_sdk and returns its stdout and stderr; replaced in tests
var runFqdnSdk = func(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "hex_sdk", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// hexSdkFqdn runs one of hex_sdk's fqdn functions. On failure the error is what the function
// printed on stderr, which names the problem for the operator.
func hexSdkFqdn(timeout ostime.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stdout, stderr, err := runFqdnSdk(ctx, args...)
	if err == nil {
		return stdout, nil
	}

	log.Errorf("fqdn: hex_sdk %s failed(%v): %s", strings.Join(args, " "), err, stderr)
	msg := strings.TrimPrefix(stderr, "Error: ")
	if msg == "" {
		msg = err.Error()
	}

	return stdout, errors.New(msg)
}

func parseFqdnCertificate(out string) (*fqdn.Certificate, error) {
	cert := &fqdn.Certificate{}
	err := json.Unmarshal([]byte(out), cert)
	if err != nil {
		return nil, fmt.Errorf("unexpected certificate summary %q: %v", out, err)
	}
	if cert.IsEmpty() {
		return nil, nil
	}

	return cert, nil
}

// GetFqdnNode returns this node's view of the FQDN and its certificate.
func GetFqdnNode() (*fqdn.Node, error) {
	out, err := hexSdkFqdn(ostime.Minute, "fqdn_status")
	if err != nil {
		return nil, err
	}

	status := struct {
		Fqdn        string          `json:"fqdn"`
		PublicHost  string          `json:"publicHost"`
		ServedFrom  string          `json:"servedFrom"`
		Certificate json.RawMessage `json:"certificate"`
	}{}
	err = json.Unmarshal([]byte(out), &status)
	if err != nil {
		return nil, fmt.Errorf("unexpected fqdn status %q: %v", out, err)
	}

	cert, err := parseFqdnCertificate(string(status.Certificate))
	if err != nil {
		return nil, err
	}

	return &fqdn.Node{
		Hostname:    base.Hostname,
		Role:        base.CurrentRole,
		Fqdn:        status.Fqdn,
		PublicHost:  status.PublicHost,
		ServedFrom:  status.ServedFrom,
		Certificate: cert,
	}, nil
}

// VerifyFqdnCertificate checks that the installed certificate can be served for name.
func VerifyFqdnCertificate(name string) error {
	_, err := os.Stat(filepath.Join(fqdn.CertificateDir, fqdn.CertificateFile))
	if err != nil {
		return fmt.Errorf("no certificate is installed; install one that covers %s first", name)
	}

	_, err = hexSdkFqdn(ostime.Minute, "fqdn_cert_verify", fqdn.CertificateDir, name)
	return err
}

// InstallFqdnCertificate checks the certificate set in dir and installs it on every control.
func InstallFqdnCertificate(dir string) (*fqdn.Certificate, error) {
	out, err := hexSdkFqdn(10*ostime.Minute, "fqdn_cert_install", dir)
	if err != nil {
		return nil, err
	}

	return parseFqdnCertificate(out)
}

// ReloadFqdnCertificate puts a replaced certificate in front of browsers on this node.
func ReloadFqdnCertificate() error {
	_, err := hexSdkFqdn(10*ostime.Minute, "fqdn_node_reload")
	return err
}

// ApplyFqdnCluster runs the cluster-wide part of an FQDN change once, on this control; with
// certOnly, only the certificates of the ingress and the ceph dashboard. Its step log is
// returned either way.
func ApplyFqdnCluster(certOnly bool) (string, error) {
	args := []string{"fqdn_cluster_apply"}
	if certOnly {
		args = append(args, "cert")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*ostime.Minute)
	defer cancel()

	stdout, stderr, err := runFqdnSdk(ctx, args...)
	output := strings.TrimSpace(strings.Join([]string{stdout, stderr}, "\n"))
	if err != nil {
		log.Errorf("fqdn: hex_sdk %s failed(%v): %s", strings.Join(args, " "), err, output)
		return output, fmt.Errorf("the cluster-wide part failed: %s", strings.TrimPrefix(stderr, "Error: "))
	}

	return output, nil
}

// SetFqdnTuning sets cubesys.fqdn on this node, or removes it for "", through the tuning
// policy, so the node also applies it when it bootstraps. The commit that follows rewrites
// the node's configuration on the public host.
func SetFqdnTuning(name string) error {
	policy, err := GetTuningPolicy(TuningPolicyFile)
	if err != nil {
		return fmt.Errorf("cannot read the tuning policy: %v", err)
	}

	if name == "" {
		policy.DeleteTuning(fqdn.Tuning)
	} else {
		policy.UpdateOrAppendTuning(tunings.Tuning{Name: fqdn.Tuning, Value: name, Enabled: true})
	}

	err = ApplyTunings(policy.Tunings)
	if err != nil {
		return fmt.Errorf("applying %s failed: %v", fqdn.Tuning, err)
	}

	applied, err := GetFqdn()
	if err != nil {
		return err
	}
	if applied != name {
		return fmt.Errorf("%s is %q after the commit, not %q", fqdn.Tuning, applied, name)
	}

	return nil
}

// RestartApiSoon restarts cube-cos-api a few seconds from now, after the answer in hand went
// out: the public host, the SAML SP and every link it builds are read at start.
func RestartApiSoon() error {
	unit := fmt.Sprintf("cube-cos-api-fqdn-restart-%d", ostime.Now().Unix())
	out, err := exec.Command(
		"systemd-run", "--on-active=5", "--unit", unit,
		"/usr/bin/systemctl", "restart", "cube-cos-api",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("scheduling the restart failed(%v): %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}
