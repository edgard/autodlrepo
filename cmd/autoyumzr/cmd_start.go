package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/edgard/autoyumzr/crawlers"
	"github.com/edgard/goutil"
	"github.com/mcuadros/go-version"
	"github.com/spf13/cobra"
)

func compare(last string, repo string) bool {
	current, err := ioutil.ReadFile(filepath.Join(repo, "RELEASE"))
	if err != nil {
		return true
	}
	if version.Compare(last, strings.TrimSpace(string(current)), ">") {
		return true
	}

	return false
}

func download(release string, url string, module string) error {
	_, err := os.Stat(config.Downloads)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(config.Downloads, 0775)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := os.Create(filepath.Join(config.Downloads, fmt.Sprintf("%s-%s.x86_64.rpm", module, release)))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func update(release string, repo string, module string) error {
	_, err := os.Stat(repo)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(repo, 0775)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = goutil.MoveFile(filepath.Join(config.Downloads, fmt.Sprintf("%s-%s.x86_64.rpm", module, release)), filepath.Join(repo, fmt.Sprintf("%s-%s.x86_64.rpm", module, release)))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(repo, "RELEASE"), []byte(release), 0644)
	if err != nil {
		return err
	}
	_, err = exec.Command("/usr/bin/createrepo", "--database", repo).Output()
	if err != nil {
		return err
	}

	return nil
}

func run(module string, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		release string
		url     string
		repo    string
		err     error
	)
	switch module {
	case "atom":
		release, url, err = crawlers.Atom()
		repo = config.Repositories.Atom
	case "vagrant":
		release, url, err = crawlers.Vagrant()
		repo = config.Repositories.Vagrant
	}
	if err != nil {
		logrus.Error(err)
		return
	}

	state := compare(release, repo)
	if state {
		logrus.Infof("Starting download %s v%s", module, release)
		err = download(release, url, module)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Infof("Finished download %s v%s", module, release)
		err = update(release, repo, module)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Infof("Finished processing %s", module)
	}

	return
}

func start(cmd *cobra.Command, args []string) {
	if flContinuous {
		logrus.Info("Running in continuous mode")
	}

	var active []string
	if flModule != "all" && goutil.StringInSlice(flModule, config.Modules) {
		active = append(active, flModule)
	} else if flModule != "all" {
		logrus.Fatal("module does not exists or is disabled")
	} else {
		active = config.Modules
	}
	logrus.Infof("Active modules: %s", strings.Join(active, ", "))

	for {
		var wg sync.WaitGroup
		wg.Add(len(active))
		for _, module := range active {
			go run(module, &wg)
		}
		wg.Wait()

		if !flContinuous {
			break
		}

		time.Sleep(30 * time.Minute)
	}
}
