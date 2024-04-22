package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
	"lesiw.io/ctrctl"
)

var containers []string
var cuid, ouid, ogid int
var dorestore bool
var ctrctlclis = [][]string{
	{"docker"},
	{"podman"},
	{"nerdctl"},
	{"lima", "nerdctl"},
}

func containerCleanup() {
	if runtime.GOOS == "linux" && dorestore {
		_ = containerChown(cuid, cuid, ouid, ogid)
	}
	for _, ctr := range containers {
		_, _ = ctrctl.ContainerRm(&ctrctl.ContainerRmOpts{Force: true}, ctr)
	}
}

func containerSetup() error {
	if err := ctrctlSetup(); err != nil {
		return err
	}
	image := os.Getenv("RUNCTR")
	if len(image) > 0 && (image[0] == '/' || image[0] == '.') {
		var err error
		if image, err = buildContainer(image); err != nil {
			return err
		}
	}
	var err error
	container, err = ctrctl.ContainerRun(
		&ctrctl.ContainerRunOpts{
			Detach:  true,
			Tty:     true,
			Volume:  root + ":/work",
			Workdir: "/work",
		},
		image,
		"cat",
	)
	if err != nil {
		return fmt.Errorf("failed to start container: %s", err)
	}
	containers = append(containers, container)
	imageid, err := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Image}}"},
		container,
	)
	if err != nil {
		return fmt.Errorf("failed to get image id of work container: %s", err)
	}
	osarch, err := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Os}}/{{.Architecture}}"},
		imageid,
	)
	if err != nil {
		return fmt.Errorf("failed to get os/arch of work container: %s", err)
	}
	ctros, ctrarch, ok := strings.Cut(osarch, "/")
	if !ok {
		return fmt.Errorf("failed to parse os/arch format: %s", err)
	}
	if err = installRunInContainer(ctros, ctrarch); err != nil {
		return err
	}
	if runtime.GOOS == "linux" {
		if err = fixFileOwners(); err != nil {
			return err
		}
	}
	return nil
}

func ctrctlSetup() error {
	if os.Getenv("RUNCTRCTL") != "" {
		cli, err := shlex.Split(os.Getenv("RUNCTRCTL"))
		if err != nil {
			return fmt.Errorf("failed to parse RUNCTRCTL: %w", err)
		}
		ctrctl.Cli = cli
		return nil
	}
	var progs []string
	for _, cli := range ctrctlclis {
		progs = append(progs, cli[0])
		path, err := exec.LookPath(cli[0])
		if err != nil {
			continue
		}
		cli[0] = path
		ctrctl.Cli = cli
		return nil
	}
	return fmt.Errorf("no container cli found. " +
		"install one of these clis: " + strings.Join(progs, ", ") + ". " +
		"or set RUNCTRCTL to another cli.")
}

func buildContainer(path string) (image string, err error) {
	imagehash := sha1.New()
	imagehash.Write(runid[:])
	imagehash.Write([]byte(path))
	image = fmt.Sprintf("%x", imagehash.Sum(nil))
	ctimestr, inspectErr := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Created}}"},
		image,
	)
	mtime, err := getMtime(path)
	if err != nil {
		err = fmt.Errorf("failed to read Containerfile '%s': %s", path, err)
		return
	}
	if inspectErr == nil {
		var ctime time.Time
		ctime, err = time.Parse(time.RFC3339, ctimestr)
		if err != nil {
			err = fmt.Errorf(
				"failed to parse container created timestamp '%s': %s",
				ctimestr, err)
			return
		}
		if ctime.Unix() > mtime {
			return // Container is newer than Containerfile.
		}
	}
	_, err = ctrctl.ImageBuild(
		&ctrctl.ImageBuildOpts{
			Cmd:     captureCmdUnlessVerbose(),
			File:    path,
			NoCache: true,
			Tag:     image,
		},
		".",
	)
	if err != nil {
		fmt.Fprint(os.Stderr, lastlog.String())
		err = fmt.Errorf("error building container '%s': %s", path, err)
	}
	return
}

func fixFileOwners() error {
	user, err := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Config.User}}"},
		container,
	)
	if err != nil {
		return fmt.Errorf("failed to get user id of container: %s", err)
	}
	if user != "" {
		cuid, err = strconv.Atoi(user)
		if err != nil {
			return fmt.Errorf("non-numeric user id: %s", user)
		}
	}
	if ouid, ogid, err = getOwner(".git"); err != nil {
		return fmt.Errorf("failed to get owner of .git directory: %s", err)
	}
	dorestore = true
	return containerChown(ouid, ogid, cuid, cuid)
}

func installRunInContainer(ctros, ctrarch string) error {
	runbin, err := fetchRun(ctros, ctrarch)
	if err != nil {
		return err
	}
	_, err = ctrctl.ContainerCp(
		&ctrctl.ContainerCpOpts{FollowLink: true},
		runbin,
		container+":/usr/bin/run",
	)
	if err != nil {
		return fmt.Errorf("failed to copy run into container: %s", err)
	}
	return nil
}

func containerChown(fuid, fgid, tuid, tgid int) error {
	_, err := ctrctl.ContainerRun(
		&ctrctl.ContainerRunOpts{
			Rm:      true,
			Volume:  root + ":/work",
			Workdir: "/work",
		},
		"lesiw/run", "-u", fmt.Sprintf("%d:%d::%d:%d", fuid, fgid, tuid, tgid),
	)
	if err != nil {
		return fmt.Errorf("failed to run chown: %s", err)
	}
	return nil
}
