package local

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/convox/praxis/types"
)

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	bid := types.Id("B", 10)

	args := []string{"run"}
	args = append(args, "--detach", "-i")
	args = append(args, "--link", "rack", "-e", "RACK_URL=https://rack:3000")
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, "-e", fmt.Sprintf("BUILD_APP=%s", app))
	args = append(args, "convox/praxis", "build", "-id", bid, "-url", url)

	cmd := exec.Command("docker", args...)

	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	pid := strings.TrimSpace(string(data))[0:10]

	build := &types.Build{
		Id:      bid,
		App:     app,
		Process: pid,
		Status:  "created",
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, bid), build); err != nil {
		return nil, err
	}

	return build, nil
}

func (p *Provider) BuildGet(app, id string) (build *types.Build, err error) {
	err = p.Load(fmt.Sprintf("apps/%s/builds/%s", app, id), &build)
	return
}

func (p *Provider) BuildLogs(app, id string) (io.Reader, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	return p.Logs(build.Process)
}

func (p *Provider) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	if opts.Manifest != "" {
		build.Manifest = opts.Manifest
	}

	if opts.Release != "" {
		build.Release = opts.Release
	}

	if opts.Status != "" {
		build.Status = opts.Status
	}

	if err := p.Store(fmt.Sprintf("apps/%s/builds/%s", app, id), build); err != nil {
		return nil, err
	}

	return build, nil
}