/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	//"github.com/kubernetes-csi/csi-driver-smb/pkg/smb"
	"k8s.io/klog/v2"
)

func init() {
	klog.InitFlags(nil)
}

var (
	endpoint       = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID         = flag.String("nodeid", "", "node id")
	version        = flag.Bool("version", false, "Print the version and exit.")
	metricsAddress = flag.String("metrics-address", "0.0.0.0:29644", "export the metrics")
	kubeconfig     = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Required only when running out of cluster.")
)

func main() {
	flag.Parse()

	if *version {
		info, err := smb.GetVersionYAML()
		if err != nil {
			klog.Fatalln(err)
		}
		fmt.Println(info)
		os.Exit(0)
	}
	if *nodeID == "" {
		// nodeid is not needed in controller component
		klog.Warning("nodeid is empty")
	}

	handle()
	os.Exit(0)
}

func handle() {

	driver := pos.NewDriver(*nodeID)
	if driver == nil {
		klog.Fatalln("Failed to initialize smb CSI Driver")
	}
	driver.Run(*endpoint, *kubeconfig, false)

}
