package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"lesiw.io/ctrctl"
)

var containers []string
var cuid, ouid, ogid int
var dorestore bool

func containerCleanup() {
	if runtime.GOOS == "linux" && dorestore {
		_ = containerChown(cuid, cuid, ouid, ogid)
	}
	for _, ctr := range containers {
		_, _ = ctrctl.ContainerRm(&ctrctl.ContainerRmOpts{Force: true}, ctr)
	}
}

func containerSetup() error {
	image := os.Getenv("PBCTR")
	var err error
	container, err = ctrctl.ContainerRun(
		&ctrctl.ContainerRunOpts{
			Detach:  true,
			Rm:      true,
			Tty:     true,
			Volume:  root + ":/work",
			Workdir: "/work",
		},
		image,
		"cat",
	)
	if err != nil {
		return fmt.Errorf("could not start container: %s", err)
	}
	containers = append(containers, container)
	imageid, err := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Image}}"},
		container,
	)
	if err != nil {
		return fmt.Errorf("could not get image id of work container: %s", err)
	}
	osarch, err := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Os}}/{{.Architecture}}"},
		imageid,
	)
	if err != nil {
		return fmt.Errorf("could not get os/arch of work container: %s", err)
	}
	ctros, ctrarch, ok := strings.Cut(osarch, "/")
	if !ok {
		return fmt.Errorf("could not parse os/arch format: %s", err)
	}
	if err = installPbInContainer(ctros, ctrarch); err != nil {
		return err
	}
	if runtime.GOOS == "linux" {
		if err = fixFileOwners(); err != nil {
			return err
		}
	}
	return nil
}

func fixFileOwners() error {
	user, err := ctrctl.Inspect(
		&ctrctl.InspectOpts{Format: "{{.Config.User}}"},
		container,
	)
	if err != nil {
		return fmt.Errorf("could not get user id of container: %s", err)
	}
	if user != "" {
		cuid, err = strconv.Atoi(user)
		if err != nil {
			return fmt.Errorf("non-numeric user id: %s", user)
		}
	}
	if ouid, ogid, err = getOwner(".git"); err != nil {
		return fmt.Errorf("could not get owner of .git directory: %s", err)
	}
	dorestore = true
	return containerChown(ouid, ogid, cuid, cuid)
}

func installPbInContainer(ctros, ctrarch string) error {
	pbbin, err := fetchPb(ctros, ctrarch)
	if err != nil {
		return err
	}
	_, err = ctrctl.ContainerCp(
		&ctrctl.ContainerCpOpts{FollowLink: true},
		pbbin,
		container+":/usr/bin/pb",
	)
	if err != nil {
		return fmt.Errorf("could not copy pb into container: %s", err)
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
		"lesiw/pb", "-u", fmt.Sprintf("%d:%d::%d:%d", fuid, fgid, tuid, tgid),
	)
	if err != nil {
		return fmt.Errorf("could not run chown: %s", err)
	}
	return nil
}
