package local

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/convox/logger"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=provider.aws")

type Provider struct {
	Root string
}

// NewProviderFromEnv returns a new AWS provider from env vars
func FromEnv() *Provider {
	return &Provider{Root: "/var/convox"}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (p *Provider) Delete(key string) error {
	if p.Root == "" {
		return fmt.Errorf("cannot delete with empty root")
	}

	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("no such key: %s", key)
	}

	return os.Remove(path)
}

func (p *Provider) DeleteAll(key string) error {
	if p.Root == "" {
		return fmt.Errorf("cannot delete with empty root")
	}

	return os.RemoveAll(filepath.Join(p.Root, key))
}

func (p *Provider) Exists(key string) bool {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func (p *Provider) Read(key string) (io.ReadCloser, error) {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("no such key: %s", key)
	}

	return os.Open(path)
}

func (p *Provider) List(key string) ([]string, error) {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(path)
	if err != nil {
		return []string{}, nil
	}

	return fd.Readdirnames(-1)
}

func (p *Provider) Load(key string, v interface{}) error {
	r, err := p.Read(key)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (p *Provider) Store(key string, v interface{}) error {
	path, err := filepath.Abs(filepath.Join(p.Root, key))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	if r, ok := v.(io.Reader); ok {
		fd, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		if _, err := io.Copy(fd, r); err != nil {
			return err
		}

		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0600)
}

func (p *Provider) Run(app, service, image, command string, args ...string) (string, error) {
	a := []string{"run", "--detach", "-i", image, command}
	a = append(a, args...)

	cmd := exec.Command("docker", a...)

	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(data)[0:10], nil
}

func (p *Provider) Logs(pid string) (io.Reader, error) {
	r, w := io.Pipe()

	cmd := exec.Command("docker", "logs", "--follow", pid)

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		w.Close()
	}()

	return r, nil
}
