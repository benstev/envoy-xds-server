//   Copyright Steve Sloka 2021
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"context"
	"flag"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
	"github.com/stevesloka/envoy-xds-server/internal"
	"github.com/stevesloka/envoy-xds-server/internal/processor"
	"github.com/stevesloka/envoy-xds-server/internal/server"
	"github.com/stevesloka/envoy-xds-server/internal/watcher"
)

var (
	l log.FieldLogger

	watchDirectoryFileName string
	port                   uint

	nodeID         string
	debug          bool
	withAcccessLog bool
)

func init() {
	l = log.New()
	// log.SetLevel(log.DebugLevel)

	flag.BoolVar(&debug, "debug", false, "Enable xDS server debug logging")

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 9002, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	// Define the directory to watch for Envoy configuration files
	flag.StringVar(&watchDirectoryFileName, "watchDirectoryFileName", "config/config.yaml", "full path to directory to watch for files")

	flag.BoolVar(&withAcccessLog, "withAccessLog", false, "Enable envoy access log")
}

func main() {
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Create a processor
	proc := processor.NewProcessor(cache, nodeID, log.WithField("context", "processor"), withAcccessLog)

	// Create initial snapshot from file
	proc.ProcessFile(watcher.NotifyMessage{
		Operation: watcher.Create,
		FilePath:  watchDirectoryFileName,
	})

	// Notify channel for file system events
	notifyCh := make(chan watcher.NotifyMessage)

	go func() {
		// Watch for file changes
		watcher.Watch(watchDirectoryFileName, notifyCh)
	}()

	go func() {
		// Run the xDS server
		ctx := context.Background()
		cb := &internal.Callbacks{Debug: debug}
		srv := serverv3.NewServer(ctx, cache, cb)
		server.RunServer(ctx, srv, port)
	}()

	for msg := range notifyCh {
		proc.ProcessFile(msg)
	}
}
