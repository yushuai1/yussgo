// /*
// * Tencent is pleased to support the open source community by making TKEStack available.
// *
// * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
// *
// * Licensed under the Apache License, Version 2.0 (the "License"); you may not use
// * this file except in compliance with the License. You may obtain a copy of the
// * License at
// *
// * https://opensource.org/licenses/Apache-2.0
// *
// * Unless required by applicable law or agreed to in writing, software
// * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
// * specific language governing permissions and limitations under the License.
// */
package main

//
//import (
//	"context"
//	"fmt"
//	"google.golang.org/grpc"
//	"k8s.io/klog"
//	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
//	"net"
//	"os"
//	"path/filepath"
//	"syscall"
//	"time"
//)
//
//type vcoreResourceServer struct {
//	Srv          *grpc.Server
//	SocketFile   string
//	Count        int
//	ResourceName string
//	SocketName   string
//	pluginapi.DevicePluginServer
//}
//
//var _ pluginapi.DevicePluginServer = &vcoreResourceServer{}
//
//func NewVDeviceServer(resorceName string, count int, socketName string) *vcoreResourceServer {
//	socketFile := filepath.Join()
//
//	return &vcoreResourceServer{
//		Srv:          grpc.NewServer(),
//		SocketFile:   socketFile,
//		ResourceName: resorceName,
//		Count:        count,
//		SocketName:   socketName,
//	}
//}
//
//func (vr *vcoreResourceServer) resgister() {
//	socketFile := filepath.Join("/var/lib/kubelet/device-plugins/", "kubelet.sock")
//	dialOptions := []grpc.DialOption{grpc.WithInsecure(), grpc.WithDialer(UnixDial), grpc.WithBlock(), grpc.WithTimeout(time.Second * 5)}
//
//	conn, err := grpc.Dial(socketFile, dialOptions...)
//	if err != nil {
//		println(err)
//	}
//	defer conn.Close()
//
//	client := pluginapi.NewRegistrationClient(conn)
//
//	req := &pluginapi.RegisterRequest{
//		Version:      pluginapi.Version,                                      // 版本信息
//		Endpoint:     vr.SocketName,                                          // 插件的endpoint
//		ResourceName: vr.ResourceName,                                        // 资源名称
//		Options:      &pluginapi.DevicePluginOptions{PreStartRequired: true}, // 插件选项 启动容器前调用DevicePlugin.PreStartContainer()
//	}
//	//endpoint=vcore.sock ResourceName=tencent.com/vcuda-core socketFile=/var/lib/kubelet/device-plugin/kubelet.sock
//	klog.V(2).Infof("Register to kubelet with endpoint=%s ResourceName=%s socketFile=%s  ", req.Endpoint, vr.ResourceName, socketFile)
//	_, err = client.Register(context.Background(), req)
//	if err != nil {
//		println(err)
//	}
//
//}
//
//func (vr *vcoreResourceServer) Run() error {
//	pluginapi.RegisterDevicePluginServer(vr.Srv, vr)
//
//	err := syscall.Unlink(vr.SocketFile)
//	if err != nil && !os.IsNotExist(err) {
//		return err
//	}
//
//	l, err := net.Listen("unix", vr.SocketFile)
//	if err != nil {
//		return err
//	}
//
//	return vr.Srv.Serve(l)
//}
//
//func (vr *vcoreResourceServer) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
//	klog.V(2).Infof("ListAndWatch request for resource")
//
//	devs := make([]*pluginapi.Device, vr.Count)
//	for i := 0; i < vr.Count; i++ {
//		devs[i] = &pluginapi.Device{
//			ID:     fmt.Sprintf("%s-%d", vr.ResourceName, i),
//			Health: pluginapi.Healthy,
//		}
//	}
//	klog.V(2).Infof("device start reported  resourceName = %s  count = %d", vr.ResourceName, len(devs))
//	s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
//
//	for {
//		time.Sleep(time.Second)
//	}
//
//	return nil
//}
//
//func (vr *vcoreResourceServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
//	return &pluginapi.DevicePluginOptions{}, nil
//}
//
//func (vr *vcoreResourceServer) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
//	return &pluginapi.PreStartContainerResponse{}, nil
//}
//
///** device plugin interface */
//func (vr *vcoreResourceServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
//	klog.V(2).Infof("%+v allocation request for vcore", reqs)
//	return &pluginapi.AllocateResponse{}, nil
//}
