// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"context"
	"time"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/fswatcher"
)

type folder struct {
	stateTracker
	config.FolderConfiguration

	scan                folderScanner
	model               *Model
	ctx                 context.Context
	cancel              context.CancelFunc
	initialScanFinished chan struct{}
	fsWatchChan         <-chan fswatcher.FsEventsBatch
}

func newFolder(model *Model, cfg config.FolderConfiguration,
	fsWatchChan <-chan fswatcher.FsEventsBatch) folder {
	ctx, cancel := context.WithCancel(context.Background())

	var intervalS int
	if fsWatchChan == nil {
		intervalS = cfg.RescanIntervalS
	} else {
		intervalS = cfg.LongRescanIntervalS
	}

	return folder{
		stateTracker:        newStateTracker(cfg.ID),
		FolderConfiguration: cfg,

		scan:                newFolderScanner(intervalS),
		ctx:                 ctx,
		cancel:              cancel,
		model:               model,
		initialScanFinished: make(chan struct{}),
		fsWatchChan:         fsWatchChan,
	}
}

func (f *folder) IndexUpdated() {
}

func (f *folder) DelayScan(next time.Duration) {
	f.scan.Delay(next)
}

func (f *folder) Scan(subdirs []string) error {
	<-f.initialScanFinished
	return f.scan.Scan(subdirs)
}

func (f *folder) Stop() {
	f.cancel()
}

func (f *folder) Jobs() ([]string, []string) {
	return nil, nil
}

func (f *folder) BringToFront(string) {}

func (f *folder) scanSubdirs(subDirs []string) error {
	if err := f.model.internalScanFolderSubdirs(f.ctx, f.folderID, subDirs); err != nil {
		// Potentially sets the error twice, once in the scanner just
		// by doing a check, and once here, if the error returned is
		// the same one as returned by CheckFolderHealth, though
		// duplicate set is handled by setError.
		f.setError(err)
		return err
	}
	return nil
}
