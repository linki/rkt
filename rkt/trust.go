// Copyright 2015 The rkt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//+build linux

// implements https://github.com/coreos/rkt/issues/367

package main

import (
	"net/url"

	"github.com/coreos/rkt/rkt/pubkey"

	"github.com/spf13/cobra"
)

var (
	cmdTrust = &cobra.Command{
		Use:   "trust [--prefix=PREFIX] [--insecure-allow-http] [--skip-fingerprint-review] [--root] [PUBKEY ...]",
		Short: "Trust a key for image verification",
		Long: `Adds keys to the local keystore for use in verifying signed images.
PUBKEY may be either a local file or URL,
PREFIX scopes the applicability of PUBKEY to image names sharing PREFIX.
Meta discovery of PUBKEY at PREFIX will be attempted if no PUBKEY is specified.
--root must be specified to add keys with no prefix; path to a key file must be given (no discovery).`,
		Run: runWrapper(runTrust),
	}
	flagPrefix                string
	flagRoot                  bool
	flagAllowHTTP             bool
	flagSkipFingerprintReview bool
)

func init() {
	cmdRkt.AddCommand(cmdTrust)
	cmdTrust.Flags().StringVar(&flagPrefix, "prefix", "", "prefix to limit trust to")
	cmdTrust.Flags().BoolVar(&flagRoot, "root", false, "add root key from filesystem without a prefix")
	cmdTrust.Flags().BoolVar(&flagSkipFingerprintReview, "skip-fingerprint-review", false, "accept key without fingerprint confirmation")
	cmdTrust.Flags().BoolVar(&flagAllowHTTP, "insecure-allow-http", false, "allow HTTP use for key discovery and/or retrieval")
}

func runTrust(cmd *cobra.Command, args []string) (exit int) {
	if flagPrefix == "" && !flagRoot {
		if len(args) != 0 {
			stderr.Print("--root required for non-prefixed (root) keys")
		} else {
			cmd.Usage()
		}
		return 1
	}

	if flagPrefix != "" && flagRoot {
		stderr.Print("--root and --prefix usage mutually exclusive")
		return 1
	}

	ks := getKeystore()
	if ks == nil {
		stderr.Print("could not get the keystore")
		return 1
	}

	// if the user included a scheme with the prefix, error on it
	u, err := url.Parse(flagPrefix)
	if err == nil && u.Scheme != "" {
		stderr.Printf("--prefix must not contain a URL scheme, omit %s://", u.Scheme)
		return 1
	}

	pkls := args
	m := &pubkey.Manager{
		InsecureAllowHTTP:  flagAllowHTTP,
		TrustKeysFromHTTPS: globalFlags.TrustKeysFromHTTPS,
		Ks:                 ks,
		Debug:              globalFlags.Debug,
	}
	if len(pkls) == 0 {
		pkls, err = m.GetPubKeyLocations(flagPrefix)
		if err != nil {
			stderr.PrintE("error determining key location", err)
			return 1
		}
	}

	acceptOpt := pubkey.AcceptAsk
	if flagSkipFingerprintReview {
		acceptOpt = pubkey.AcceptForce
	}
	if err := m.AddKeys(pkls, flagPrefix, acceptOpt, pubkey.OverrideDeny); err != nil {
		stderr.PrintE("error adding keys", err)
		return 1
	}

	return 0
}
