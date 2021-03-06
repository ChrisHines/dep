// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dep

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/dep/internal/gps"
)

var errProjectNotFound = fmt.Errorf("could not find project %s, use dep init to initiate a manifest", ManifestName)
var errVendorBackupFailed = fmt.Errorf("Failed to create vendor backup. File with same name exists.")

// findProjectRoot searches from the starting directory upwards looking for a
// manifest file until we get to the root of the filesystem.
func findProjectRoot(from string) (string, error) {
	for {
		mp := filepath.Join(from, ManifestName)

		_, err := os.Stat(mp)
		if err == nil {
			return from, nil
		}
		if !os.IsNotExist(err) {
			// Some err other than non-existence - return that out
			return "", err
		}

		parent := filepath.Dir(from)
		if parent == from {
			return "", errProjectNotFound
		}
		from = parent
	}
}

type Project struct {
	// AbsRoot is the absolute path to the root directory of the project.
	AbsRoot string
	// ImportRoot is the import path of the project's root directory.
	ImportRoot gps.ProjectRoot
	Manifest   *Manifest
	Lock       *Lock
}

// MakeParams is a simple helper to create a gps.SolveParameters without setting
// any nils incorrectly.
func (p *Project) MakeParams() gps.SolveParameters {
	params := gps.SolveParameters{
		RootDir:         p.AbsRoot,
		ProjectAnalyzer: Analyzer{},
	}

	if p.Manifest != nil {
		params.Manifest = p.Manifest
	}

	if p.Lock != nil {
		params.Lock = p.Lock
	}

	return params
}

// BackupVendor looks for existing vendor directory and if it's not empty,
// creates a backup of it to a new directory with the provided suffix.
func BackupVendor(vpath, suffix string) (string, error) {
	// Check if there's a non-empty vendor directory
	vendorExists, err := IsNonEmptyDir(vpath)
	if err != nil {
		return "", err
	}
	if vendorExists {
		vendorbak := vpath + "-" + suffix
		// Check if a directory with same name exists
		if _, err = os.Stat(vendorbak); os.IsNotExist(err) {
			// Rename existing vendor to vendor-{suffix}
			if err := renameWithFallback(vpath, vendorbak); err != nil {
				return "", err
			}
			return vendorbak, nil
		} else {
			return "", errVendorBackupFailed
		}
	}

	return "", nil
}
