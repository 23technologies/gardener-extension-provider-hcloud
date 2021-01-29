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

package connectionmanager

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/mo"
	"k8s.io/klog"

	vclib "k8s.io/cloud-provider-vsphere/pkg/common/vclib"
)

// String returns the string representation of the FindVM constant.
func (f FindVM) String() string {
	switch f {
	case FindVMByUUID:
		return "byUUID"
	case FindVMByName:
		return "byName"
	case FindVMByIP:
		return "byIP"
	default:
		return "byUnknown"
	}
}

// WhichVCandDCByNodeID finds the VC/DC combo that owns a particular VM
func (cm *ConnectionManager) WhichVCandDCByNodeID(ctx context.Context, nodeID string, searchBy FindVM) (*VMDiscoveryInfo, error) {
	if nodeID == "" {
		klog.V(3).Info("WhichVCandDCByNodeID called but nodeID is empty")
		return nil, errors.New("nodeID is empty")
	}
	type vmSearch struct {
		tenantRef  string
		vc         string
		datacenter *vclib.Datacenter
	}

	var mutex = &sync.Mutex{}
	var globalErrMutex = &sync.Mutex{}
	var queueChannel chan *vmSearch
	var wg sync.WaitGroup
	var globalErr *error

	queueChannel = make(chan *vmSearch, QueueSize)

	myNodeID := nodeID
	switch searchBy {
	case FindVMByUUID:
		klog.V(3).Info("WhichVCandDCByNodeID by UUID")
		myNodeID = strings.TrimSpace(strings.ToLower(nodeID))
	case FindVMByIP:
		klog.V(3).Info("WhichVCandDCByNodeID by IP")
	default:
		klog.V(3).Info("WhichVCandDCByNodeID by Name")
	}
	klog.V(2).Info("WhichVCandDCByNodeID nodeID: ", myNodeID)

	vmFound := false
	globalErr = nil

	setGlobalErr := func(err error) {
		globalErrMutex.Lock()
		globalErr = &err
		globalErrMutex.Unlock()
	}

	setVMFound := func(found bool) {
		mutex.Lock()
		vmFound = found
		mutex.Unlock()
	}

	getVMFound := func() bool {
		mutex.Lock()
		found := vmFound
		mutex.Unlock()
		return found
	}

	go func() {
		for _, vsi := range cm.VsphereInstanceMap {
			var datacenterObjs []*vclib.Datacenter

			if getVMFound() {
				break
			}

			var err error
			for i := 0; i < NumConnectionAttempts; i++ {
				err = cm.Connect(ctx, vsi)
				if err == nil {
					break
				}
				time.Sleep(time.Duration(RetryAttemptDelaySecs) * time.Second)
			}

			if err != nil {
				klog.Error("WhichVCandDCByNodeID error vc:", err)
				setGlobalErr(err)
				continue
			}

			if vsi.Cfg.Datacenters == "" {
				datacenterObjs, err = vclib.GetAllDatacenter(ctx, vsi.Conn)
				if err != nil {
					klog.Error("WhichVCandDCByNodeID error dc:", err)
					setGlobalErr(err)
					continue
				}
			} else {
				datacenters := strings.Split(vsi.Cfg.Datacenters, ",")
				for _, dc := range datacenters {
					dc = strings.TrimSpace(dc)
					if dc == "" {
						continue
					}
					datacenterObj, err := vclib.GetDatacenter(ctx, vsi.Conn, dc)
					if err != nil {
						klog.Error("WhichVCandDCByNodeID error dc:", err)
						setGlobalErr(err)
						continue
					}
					datacenterObjs = append(datacenterObjs, datacenterObj)
				}
			}

			for _, datacenterObj := range datacenterObjs {
				if getVMFound() {
					break
				}

				klog.V(4).Infof("Finding node %s in vc=%s and datacenter=%s", myNodeID, vsi.Cfg.VCenterIP, datacenterObj.Name())
				queueChannel <- &vmSearch{
					tenantRef:  vsi.Cfg.TenantRef,
					vc:         vsi.Cfg.VCenterIP,
					datacenter: datacenterObj,
				}
			}
		}
		close(queueChannel)
	}()

	var vmInfo *VMDiscoveryInfo
	for i := 0; i < PoolSize; i++ {
		wg.Add(1)
		go func() {
			for res := range queueChannel {
				var vm *vclib.VirtualMachine
				var err error

				switch searchBy {
				case FindVMByUUID:
					vm, err = res.datacenter.GetVMByUUID(ctx, myNodeID)
				case FindVMByIP:
					vm, err = res.datacenter.GetVMByIP(ctx, myNodeID)
				default:
					vm, err = res.datacenter.GetVMByDNSName(ctx, myNodeID)
				}

				if err != nil {
					klog.Errorf("Error while looking for vm=%s(%s) in vc=%s and datacenter=%s: %v",
						myNodeID, searchBy, res.vc, res.datacenter.Name(), err)
					if err != vclib.ErrNoVMFound {
						setGlobalErr(err)
					} else {
						klog.V(2).Infof("Did not find node %s in vc=%s and datacenter=%s",
							myNodeID, res.vc, res.datacenter.Name())
					}
					continue
				}

				var oVM mo.VirtualMachine
				err = vm.Properties(ctx, vm.Reference(), []string{"config", "summary", "guest"}, &oVM)
				if err != nil {
					klog.Errorf("Error collecting properties for vm=%+v in vc=%s and datacenter=%s: %v",
						vm, res.vc, res.datacenter.Name(), err)
					continue
				}

				hostName := oVM.Guest.HostName
				if searchBy == FindVMByIP {
					klog.V(2).Infof("WhichVCandDCByNodeID by IP. Overriding VMName from=%s to to=%s", oVM.Guest.HostName, myNodeID)
					hostName = myNodeID
				}

				UUID := strings.ToLower(strings.TrimSpace(oVM.Summary.Config.Uuid))

				klog.V(2).Infof("Found node %s as vm=%+v in vc=%s and datacenter=%s",
					nodeID, vm, res.vc, res.datacenter.Name())
				klog.V(2).Infof("Hostname: %s, UUID: %s", hostName, UUID)

				vmInfo = &VMDiscoveryInfo{TenantRef: res.tenantRef, DataCenter: res.datacenter, VM: vm, VcServer: res.vc,
					UUID: UUID, NodeName: hostName}
				setVMFound(true)
				break
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if vmFound {
		return vmInfo, nil
	}
	if globalErr != nil {
		return nil, *globalErr
	}

	klog.V(4).Infof("WhichVCandDCByNodeID: %q vm not found", myNodeID)
	return nil, vclib.ErrNoVMFound
}

// WhichVCandDCByFCDId searches for an FCD using the provided ID.
func (cm *ConnectionManager) WhichVCandDCByFCDId(ctx context.Context, fcdID string) (*FcdDiscoveryInfo, error) {
	if fcdID == "" {
		klog.V(3).Info("WhichVCandDCByFCDId called but fcdID is empty")
		return nil, vclib.ErrNoDiskIDFound
	}
	klog.V(2).Info("WhichVCandDCByFCDId fcdID: ", fcdID)

	type fcdSearch struct {
		tenantRef  string
		vc         string
		datacenter *vclib.Datacenter
	}

	var mutex = &sync.Mutex{}
	var globalErrMutex = &sync.Mutex{}
	var queueChannel chan *fcdSearch
	var wg sync.WaitGroup
	var globalErr *error

	queueChannel = make(chan *fcdSearch, QueueSize)

	fcdFound := false
	globalErr = nil

	setGlobalErr := func(err error) {
		globalErrMutex.Lock()
		globalErr = &err
		globalErrMutex.Unlock()
	}

	setFCDFound := func(found bool) {
		mutex.Lock()
		fcdFound = found
		mutex.Unlock()
	}

	getFCDFound := func() bool {
		mutex.Lock()
		found := fcdFound
		mutex.Unlock()
		return found
	}

	go func() {
		for _, vsi := range cm.VsphereInstanceMap {
			var datacenterObjs []*vclib.Datacenter

			if getFCDFound() {
				break
			}

			var err error
			for i := 0; i < NumConnectionAttempts; i++ {
				err = cm.Connect(ctx, vsi)
				if err == nil {
					break
				}
				time.Sleep(time.Duration(RetryAttemptDelaySecs) * time.Second)
			}

			if err != nil {
				klog.Error("WhichVCandDCByFCDId error vc:", err)
				setGlobalErr(err)
				continue
			}

			if vsi.Cfg.Datacenters == "" {
				datacenterObjs, err = vclib.GetAllDatacenter(ctx, vsi.Conn)
				if err != nil {
					klog.Error("WhichVCandDCByFCDId error dc:", err)
					setGlobalErr(err)
					continue
				}
			} else {
				datacenters := strings.Split(vsi.Cfg.Datacenters, ",")
				for _, dc := range datacenters {
					dc = strings.TrimSpace(dc)
					if dc == "" {
						continue
					}
					datacenterObj, err := vclib.GetDatacenter(ctx, vsi.Conn, dc)
					if err != nil {
						klog.Error("WhichVCandDCByFCDId error dc:", err)
						setGlobalErr(err)
						continue
					}
					datacenterObjs = append(datacenterObjs, datacenterObj)
				}
			}

			for _, datacenterObj := range datacenterObjs {
				if getFCDFound() {
					break
				}

				klog.V(4).Infof("Finding FCD %s in vc=%s and datacenter=%s", fcdID, vsi.Cfg.VCenterIP, datacenterObj.Name())
				queueChannel <- &fcdSearch{
					tenantRef:  vsi.Cfg.TenantRef,
					vc:         vsi.Cfg.VCenterIP,
					datacenter: datacenterObj,
				}
			}
		}
		close(queueChannel)
	}()

	var fcdInfo *FcdDiscoveryInfo
	for i := 0; i < PoolSize; i++ {
		wg.Add(1)
		go func() {
			for res := range queueChannel {

				fcd, err := res.datacenter.DoesFirstClassDiskExist(ctx, fcdID)
				if err != nil {
					klog.Errorf("Error while looking for FCD=%+v in vc=%s and datacenter=%s: %v",
						fcd, res.vc, res.datacenter.Name(), err)
					if err != vclib.ErrNoDiskIDFound {
						setGlobalErr(err)
					} else {
						klog.V(2).Infof("Did not find FCD %s in vc=%s and datacenter=%s",
							fcdID, res.vc, res.datacenter.Name())
					}
					continue
				}

				klog.V(2).Infof("Found FCD %s as vm=%+v in vc=%s and datacenter=%s",
					fcdID, fcd, res.vc, res.datacenter.Name())

				fcdInfo = &FcdDiscoveryInfo{TenantRef: res.tenantRef, DataCenter: res.datacenter, FCDInfo: fcd, VcServer: res.vc}
				setFCDFound(true)
				break
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if fcdFound {
		return fcdInfo, nil
	}
	if globalErr != nil {
		return nil, *globalErr
	}

	klog.V(4).Infof("WhichVCandDCByFCDId: %q FCD not found", fcdID)
	return nil, vclib.ErrNoDiskIDFound
}
