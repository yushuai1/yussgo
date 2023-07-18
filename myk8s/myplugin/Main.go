package main

import (
	"context"
	"fmt"
	"github.com/yu-jia-ying/go-util/strs"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type vcoreResourceServer struct {
	Srv          *grpc.Server
	Count        int
	ResourceName string
	SocketName   string
	pluginapi.DevicePluginServer
}

var _ pluginapi.DevicePluginServer = &vcoreResourceServer{}

func NewVDeviceServer(resorceName string, count int, socketName string) *vcoreResourceServer {

	return &vcoreResourceServer{
		Srv:          grpc.NewServer(),
		ResourceName: resorceName,
		Count:        count,
		SocketName:   socketName,
	}
}

func (vr *vcoreResourceServer) resgister() {
	socketFile := filepath.Join("/var/lib/kubelet/device-plugins/", "kubelet.sock")
	dialOptions := []grpc.DialOption{grpc.WithInsecure(), grpc.WithDialer(UnixDial), grpc.WithBlock(), grpc.WithTimeout(time.Second * 5)}

	conn, err := grpc.Dial(socketFile, dialOptions...)
	if err != nil {
		println(err)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)

	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,                                      // 版本信息
		Endpoint:     vr.SocketName,                                          // 插件的endpoint
		ResourceName: vr.ResourceName,                                        // 资源名称
		Options:      &pluginapi.DevicePluginOptions{PreStartRequired: true}, // 插件选项 启动容器前调用DevicePlugin.PreStartContainer()
	}
	//endpoint=vcore.sock ResourceName=tencent.com/vcuda-core socketFile=/var/lib/kubelet/device-plugin/kubelet.sock
	println(vr.SocketName, req.Endpoint, vr.ResourceName, socketFile)
	_, err = client.Register(context.Background(), req)
	if err != nil {
		println(err)
	}

}

func (vr *vcoreResourceServer) Run() error {
	pluginapi.RegisterDevicePluginServer(vr.Srv, vr)

	path := filepath.Join("/var/lib/kubelet/device-plugins", vr.SocketName)
	err := syscall.Unlink(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	println(path)
	return vr.Srv.Serve(l)
}

func (vr *vcoreResourceServer) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	klog.V(2).Infof("ListAndWatch request for resource")

	devs := make([]*pluginapi.Device, vr.Count)
	for i := 0; i < vr.Count; i++ {
		devs[i] = &pluginapi.Device{
			ID:     fmt.Sprintf("%s-%d", vr.ResourceName, i),
			Health: pluginapi.Healthy,
		}
	}
	klog.V(2).Infof("device start reported  resourceName = %s  count = %d", vr.ResourceName, len(devs))
	s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

	for {
		time.Sleep(time.Second)
	}

	return nil
}

func (vr *vcoreResourceServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (vr *vcoreResourceServer) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

/** device plugin interface */
func (vr *vcoreResourceServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	klog.V(2).Infof("%+v allocation request for vcore", reqs)
	return &pluginapi.AllocateResponse{}, nil
}

func UnixDial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", addr, timeout)
}

func getK8sClient() *kubernetes.Clientset {
	var (
		k8sclient *kubernetes.Clientset
		clientCfg *rest.Config
	)
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		println("invalid client config: err(%v)", err)
	}

	k8sclient, err = kubernetes.NewForConfig(clientCfg)

	return k8sclient
}

func main() {

	//ss := "amd.com/gput$10$amd.sock"
	cmds := os.Args
	for i, cmd := range cmds {
		fmt.Printf("cmd[%d] = %s \n", i, cmd)
		if i == 0 {
			continue
		}
		ar := strings.Split(cmd, "|")

		for _, v := range ar {

			b := strings.Split(v, "*")
			t, _ := strs.StrToInt(b[1])
			m := NewVDeviceServer(b[0], t, b[2])

			go m.Run()
			time.Sleep(10 * time.Second)
			go m.resgister()
		}
	}

	time.Sleep(1000 * time.Hour)
}
