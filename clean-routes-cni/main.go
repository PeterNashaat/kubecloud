// Copyright 2015 CNI authors
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

package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/disk"
	"github.com/vishvananda/netlink"
)

func main() {
	skel.PluginMainFuncs(skel.CNIFuncs{
		Add:   cmdAdd,
		Check: cmdCheck,
		Del:   cmdDel,
		/* FIXME GC */
		/* FIXME Status */
	}, version.All, bv.BuildString("host-local"))
}

func cmdCheck(args *skel.CmdArgs) error {
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// Look to see if there is at least one IP address allocated to the container
	// in the data dir, irrespective of what that address actually is
	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	containerIPFound := store.FindByID(args.ContainerID, args.IfName)
	if !containerIPFound {
		return fmt.Errorf("host-local: Failed to find address added by container %v", args.ContainerID)
	}

	return nil
}
func RandomMyceliumIPSeed() ([]byte, error) {
	key := make([]byte, 6)
	_, err := rand.Read(key)
	return key, err
}

func cmdAdd(args *skel.CmdArgs) error {

	containerNetns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNetns.Close()

	// Setup in container netns
	err = containerNetns.Do(func(_ ns.NetNS) error {

		// find ipv6 default route and remove it to make all outside requests through ipv4
		// Parse the default route destination (IPv6 ::/0)
		_, routeDst, err := net.ParseCIDR("::/0")
		if err != nil {
			log.Fatalf("Failed to parse CIDR: %v", err)
		}

		// Get all routes
		routes, err := netlink.RouteList(nil, netlink.FAMILY_V6)
		if err != nil {
			log.Fatalf("Failed to list routes: %v", err)
		}

		for _, r := range routes {
			if r.Dst != nil && r.Dst.String() == routeDst.String() {
				log.Printf("Deleting default IPv6 route via %s dev %d", r.Gw, r.LinkIndex)
				if err := netlink.RouteDel(&r); err != nil {
					log.Printf("❌ Failed to delete route: %v", err)
				} else {
					log.Printf("✅ Deleted: %v", r)
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to clean ipv6 route in pod namespace: %v", err)
	}

	// Return result
	result := &current.Result{
		CNIVersion: "0.4.0",
		Interfaces: []*current.Interface{},
		IPs:        []*current.IPConfig{},
	}

	return types.PrintResult(result, "0.4.0")

}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}
