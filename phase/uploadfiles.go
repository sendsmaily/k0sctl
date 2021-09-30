package phase

import (
	"path"

	"github.com/alessio/shellescape"
	"github.com/k0sproject/k0sctl/config"
	"github.com/k0sproject/k0sctl/config/cluster"
	"github.com/k0sproject/rig/exec"

	log "github.com/sirupsen/logrus"
)

// UploadFiles implements a phase which upload files to hosts
type UploadFiles struct {
	GenericPhase

	hosts cluster.Hosts
}

// Title for the phase
func (p *UploadFiles) Title() string {
	return "Upload files to hosts"
}

// Prepare the phase
func (p *UploadFiles) Prepare(config *config.Cluster) error {
	p.Config = config
	p.hosts = p.Config.Spec.Hosts.Filter(func(h *cluster.Host) bool {
		return len(h.Files) > 0
	})

	return nil
}

// ShouldRun is true when there are workers
func (p *UploadFiles) ShouldRun() bool {
	return len(p.hosts) > 0
}

// Run the phase
func (p *UploadFiles) Run() error {
	return p.Config.Spec.Hosts.ParallelEach(p.uploadFiles)
}

func (p *UploadFiles) uploadFiles(h *cluster.Host) error {
	var resolved []cluster.UploadFile

	for _, f := range h.Files {
		log.Infof("%s: starting upload of %s", h, f)
		files, err := f.Resolve()
		if err != nil {
			return err
		}
		resolved = append(resolved, files...)
	}

	for _, f := range resolved {
		tmpfile, err := h.Configurer.TempFile(h)
		if err != nil {
			return err
		}

		destdir, destfile, err := f.Destination()
		if err != nil {
			return err
		}

		if f.IsURL() {
			err = p.uploadURL(h, f, tmpfile)
		} else {
			err = p.uploadLocal(h, f, tmpfile)
		}

		if err != nil {
			return err
		}

		dest := path.Join(destdir, destfile)
		log.Infof("%s: installing %s to %s", h, f, dest)
		err = h.Execf("install -m %s -D %s %s", f.PermMode, shellescape.Quote(tmpfile), shellescape.Quote(dest), exec.Sudo(h))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *UploadFiles) uploadLocal(h *cluster.Host, f cluster.UploadFile, dest string) error {
	log.Infof("%s: uploading %s", h, f)
	return h.Upload(f.Source, dest)
}
func (p *UploadFiles) uploadURL(h *cluster.Host, f cluster.UploadFile, dest string) error {
	log.Infof("%s: downloading %s", h, f)
	return h.Configurer.DownloadURL(h, f.Source, dest)
}
