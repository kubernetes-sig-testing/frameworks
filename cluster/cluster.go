/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package cluster is a Test Cluster Framework -- see the motivating
// document here:
// https://docs.google.com/document/d/13bMjmWpsdkgbY-JayrcU-e_QNwRJCP-rHjtqdeeoQHo/edit?ts=5aa005c9#heading=h.75awtuvlo3ad
//
// This package aims to provide a consistent abstraction over a number
// of different ways of creating kubernetes clusters for testing.
//
// To make your test cluster compatible with this framework, just
// implement the Fixture interface.  If you require more config
// options than we have so far, add them to the Config struct. Bear in
// mind while doing so that we aim to maintain compatibility with
// kubeadm, so if your new config options describe properties that
// already exist in the kubeadm config, you should use
// kubeadm-compatibile terminology.
package cluster

import (
	"sigs.k8s.io/testing_frameworks/cluster/type/base"
	"sigs.k8s.io/testing_frameworks/cluster/type/lightweight"
)

// Fixture is some kind of test cluster fixture, which can be started, interacted with, and stopped.
type Fixture interface {
	// Setup starts the test cluster according to the provided
	// configuration. If the config asks for a feature that is not
	// supported by this Fixture implementation, or if the test
	// cluster fails to start, return an error.
	//
	// This should block until the test cluster control plane is
	// ready to recieve client connections.
	Setup(config Config) error

	// TearDown cleanly stops the test cluster. If we can't stop
	// cleanly, return an error.
	//
	// TearDown should block until the test cluster has stopped,
	// or we have given up on stopping it and returned an error.
	//
	// TearDown should be idempotent. If a user calls TearDown
	// twice in a row, and the first call succeded, then the
	// second call should also succeed.
	TearDown() error

	// ClientConfig returns all the configuration a client needs to connect to a
	// fixture's APIServer.
	// The type returned is a copy of `cliencmd.Config` from `k/k` and therefore
	// should be compatible when serialized. When serialized to a file, it can be
	// used as a kubeconfig file.
	ClientConfig() base.Config
}

func DoDefaulting(c Config) Config {
	if c.Etcd.Local == nil {
		c.Etcd.Local = &base.LocalEtcd{}
	}
	return c
}

// Those type aliases are only used to work around the 'duplicate field' error
// while allowing the packages which define those nested types to still use the
// same names.
//
// When new types to be nested are introduced they can still be called e.g.
// `MasterConfiguration` or `Etcd` in their package. We can just use the type
// aliases here and use those to nest into the "main" structs.
type (
	lightweightMasterConfiguration = lightweight.MasterConfiguration
	lightweightEtcd                = lightweight.Etcd
	lightweightAPI                 = lightweight.API
)

// Config is a struct into which you can parse a YAML or JSON config
// file to describe your test cluster.
//
// It consists of the a base MasterConfig and additional configuration
// extensions for different test cluster implementations.
type Config struct {
	// Etcd holds configuration for etcd.
	Etcd Etcd

	// API holds configuration for the k8s apiserver.
	API API

	// Shape describes the shape of a cluster.
	Shape Shape

	base.ClusterConfiguration
	lightweightMasterConfiguration
}

// Etcd contains elements describing Etcd configuration.
//
// It consists of a base Etcd and additional configuration
// extensions for different test cluster implementations.
type Etcd struct {
	base.Etcd
	lightweightEtcd
}

// API contains elements describing APIServer configuration.
//
// It consists of a base API struct and additional configuration extensions for
// different test cluster implementations.
type API struct {
	base.API
	lightweightAPI
}

// Shape describs the shape of a cluster to be brought up.
type Shape struct {
	NodeCount int
}
