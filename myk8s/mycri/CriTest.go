package main

import (
	"bufio"
	"flag"
	"fmt"
	cgroupsystemd "github.com/opencontainers/runc/libcontainer/cgroups/systemd"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubectl/pkg/util/qos"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)
import "time"

const (
	PodNamespaceLabelKey            = "io.kubernetes.pod.namespace"
	ContainerNameLabelKey           = "io.kubernetes.container.name"
	PodNameLabelKey                 = "io.kubernetes.pod.name"
	PodUIDLabelKey                  = "io.kubernetes.pod.uid"
	DefaultContainerRuntimeEndpoint = "/var/run/dockershim.sock"
	//DefaultContainerRuntimeEndpoint = "/run/containerd/containerd.sock"
	//DefaultContainerRuntimeEndpoint ="/var/run/crio/crio.sock"
	PodCgroupNamePrefix = "pod"
	DefaultCgroupDriver = "systemd"
	Cgroupfs            = "cgroupfs"
	systemdSuffix       = ".slice"
	CGROUP_BASE         = "/sys/fs/cgroup/memory"
	CGROUP_PROCS        = "cgroup.procs"
)

type CgroupName []string

func escapeSystemdCgroupName(part string) string {
	return strings.Replace(part, "-", "_", -1)
}
func (cgroupName CgroupName) ToSystemd() string {
	if len(cgroupName) == 0 || (len(cgroupName) == 1 && cgroupName[0] == "") {
		return "/"
	}
	newparts := []string{}
	for _, part := range cgroupName {
		part = escapeSystemdCgroupName(part)
		newparts = append(newparts, part)
	}

	result, err := cgroupsystemd.ExpandSlice(strings.Join(newparts, "-") + systemdSuffix)
	if err != nil {
		// Should never happen...
		panic(fmt.Errorf("error converting cgroup name [%v] to systemd format: %v", cgroupName, err))
	}
	return result
}
func (cgroupName CgroupName) ToCgroupfs() string {
	return "/" + path.Join(cgroupName...)
}

func MYNewCgroupName(base CgroupName, components ...string) CgroupName {
	for _, component := range components {
		// Forbit using "_" in internal names. When remapping internal
		// names to systemd cgroup driver, we want to remap "-" => "_",
		// so we forbid "_" so that we can always reverse the mapping.
		if strings.Contains(component, "/") || strings.Contains(component, "_") {
			panic(fmt.Errorf("invalid character in component [%q] of CgroupName", component))
		}
	}
	// copy data from the base cgroup to eliminate cases where CgroupNames share underlying slices.  See #68416
	baseCopy := make([]string, len(base))
	copy(baseCopy, base)
	return CgroupName(append(baseCopy, components...))
}

var (
	containerRoot = MYNewCgroupName([]string{}, "kubepods")
)

type MYContainerRuntimeInterface interface {
	MYGetPidsInContainers(containerID string) ([]int, error)

	MYInspectContainer(containerID string) (*criapi.ContainerStatus, error)

	MYRuntimeName() string
}

type myContainerRuntimeManager struct {
	cgroupDriver   string
	runtimeName    string
	requestTimeout time.Duration
	client         criapi.RuntimeServiceClient
	k8sclient      *kubernetes.Clientset
}

var _ MYContainerRuntimeInterface = (*myContainerRuntimeManager)(nil)

func UnixDial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", addr, timeout)
}
func MYNewContainerRuntimeManager(cgroupDriver, endpoint string, requestTimeout time.Duration) (*myContainerRuntimeManager, error) {
	dialOptions := []grpc.DialOption{grpc.WithInsecure(), grpc.WithDialer(UnixDial), grpc.WithBlock(), grpc.WithTimeout(time.Second * 5)}
	conn, err := grpc.Dial(endpoint, dialOptions...)
	if err != nil {
		return nil, err
	}

	clientcri := criapi.NewRuntimeServiceClient(conn)

	m := &myContainerRuntimeManager{
		cgroupDriver:   cgroupDriver,
		client:         clientcri,
		requestTimeout: requestTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.requestTimeout)
	defer cancel()
	resp, err := m.client.Version(ctx, &criapi.VersionRequest{Version: "0.1.0"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Container runtime is %s \n", resp.RuntimeName)
	m.runtimeName = resp.RuntimeName

	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		fmt.Errorf("invalid client config: err(%v) \n", err)
	}

	k8sclient, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		fmt.Errorf("k8s client : err(%v) \n", err)
	}
	m.k8sclient = k8sclient
	return m, nil
}

func (m *myContainerRuntimeManager) GetPod(namespace, name string) (*v1.Pod, error) {
	vpod, err := m.k8sclient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return vpod, err
}

func (m *myContainerRuntimeManager) GetListPod() {
	vpod, err := m.k8sclient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		println(err)
	}
	fmt.Printf("podlist size %v\n", len(vpod.Items))
}

func (m *myContainerRuntimeManager) getCgroupName(pod *v1.Pod, containerID string) (string, error) {
	podQos := pod.Status.QOSClass
	if len(podQos) == 0 {
		podQos = qos.GetPodQOS(pod)
	}

	var parentContainer CgroupName
	switch podQos {
	case v1.PodQOSGuaranteed:
		parentContainer = MYNewCgroupName(containerRoot)
	case v1.PodQOSBurstable:
		parentContainer = MYNewCgroupName(containerRoot, strings.ToLower(string(v1.PodQOSBurstable)))
	case v1.PodQOSBestEffort:
		parentContainer = MYNewCgroupName(containerRoot, strings.ToLower(string(v1.PodQOSBestEffort)))
	}

	podContainer := PodCgroupNamePrefix + string(pod.UID)
	cgroupName := MYNewCgroupName(parentContainer, podContainer)

	switch m.cgroupDriver {
	case "systemd":
		return fmt.Sprintf("%s/%s-%s.scope", cgroupName.ToSystemd(), m.runtimeName, containerID), nil
	case "cgroupfs":
		return fmt.Sprintf("%s/%s", cgroupName.ToCgroupfs(), containerID), nil
	default:
	}

	return "", fmt.Errorf("unsupported cgroup driver")
}
func readProcsFile(file string) ([]int, error) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("can't read %s, %v\n", file, err)
		return nil, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	pids := make([]int, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if pid, err := strconv.Atoi(line); err == nil {
			pids = append(pids, pid)
		}
	}

	fmt.Printf("Read from %s, pids: %v\n", file, pids)
	return pids, nil
}
func (m *myContainerRuntimeManager) MYGetPidsInContainers(containerID string) ([]int, error) {

	req := &criapi.ContainerStatusRequest{
		ContainerId: containerID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.requestTimeout)
	defer cancel()

	resp, err := m.client.ContainerStatus(ctx, req)
	if err != nil {
		fmt.Printf("can't get container %s status, %v \n", containerID, err)
		return nil, err
	}

	ns := resp.Status.Labels[PodNamespaceLabelKey]
	podName := resp.Status.Labels[PodNameLabelKey]

	pod, err := m.GetPod(ns, podName)
	if err != nil {
		fmt.Printf("can't get pod %s/%s, %v\n", ns, podName, err)
		return nil, err
	}

	cgroupPath, err := m.getCgroupName(pod, containerID)
	if err != nil {
		fmt.Printf("can't get cgroup parent, %v \n", err)
		return nil, err
	}

	pids := make([]int, 0)
	baseDir := filepath.Clean(filepath.Join(CGROUP_BASE, cgroupPath))

	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() || info.Name() != CGROUP_PROCS {
			return nil
		}

		p, err := readProcsFile(path)
		if err == nil {
			pids = append(pids, p...)
		}

		return nil
	})

	return pids, nil
}

func (m *myContainerRuntimeManager) MYInspectContainer(containerID string) (*criapi.ContainerStatus, error) {
	//TODO implement me
	req := &criapi.ContainerStatusRequest{
		ContainerId: containerID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.requestTimeout)
	defer cancel()

	resp, err := m.client.ContainerStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Status, nil
}

func (m *myContainerRuntimeManager) MYRuntimeName() string {
	//TODO implement me
	return m.runtimeName
}

func (m *myContainerRuntimeManager) MYContainerStatus(conid string) {
	req := &criapi.ContainerStatusRequest{
		ContainerId: conid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.requestTimeout)
	defer cancel()

	resp, err := m.client.ContainerStatus(ctx, req)
	if err != nil {
		fmt.Printf("can't get container %s status, %v \n", conid, err)
		panic(err)
	}
	ns := resp.Status.Labels[PodNamespaceLabelKey]
	podName := resp.Status.Labels[PodNameLabelKey]
	fmt.Printf("ns is %s podName is %s\n", ns, podName)

}

func (m *myContainerRuntimeManager) MYContainerStop(conid string) {
	req := &criapi.StopContainerRequest{
		ContainerId: conid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.requestTimeout)
	defer cancel()

	_, err := m.client.StopContainer(ctx, req)
	if err != nil {
		fmt.Printf("can't get container %s status, %v \n", conid, err)
		panic(err)
	}

}

func (m *myContainerRuntimeManager) MYContainerUpdateResourc(conid string) {
	linux := &criapi.LinuxContainerResources{
		MemoryLimitInBytes: 1024 * 1024 * 5,
	}
	req := &criapi.UpdateContainerResourcesRequest{
		ContainerId: conid,
		Linux:       linux,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.requestTimeout)
	defer cancel()

	_, err := m.client.UpdateContainerResources(ctx, req)
	if err != nil {
		fmt.Printf("can't get container %s status, %v \n", conid, err)
		panic(err)
	}

}

var (
	myns         string
	mypodname    string
	cgroupDriver string
)

func main() {

	flag.StringVar(&myns, "ns", "test", "ns")
	flag.StringVar(&mypodname, "podname", "test1", "podname")
	flag.StringVar(&cgroupDriver, "cgroupDriver", DefaultCgroupDriver, "DefaultCgroupDriver")
	flag.Parse()

	containerRuntimeManager, err := MYNewContainerRuntimeManager(DefaultCgroupDriver, DefaultContainerRuntimeEndpoint, time.Second*5)
	if err != nil {
		fmt.Println(err)
	}
	containerRuntimeManager.GetListPod()
	pod, err := containerRuntimeManager.GetPod(myns, mypodname)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 2)
	fmt.Printf("podName %s\n", pod.Name)
	var containerId string
	for _, stat := range pod.Status.ContainerStatuses {
		containerId = strings.TrimPrefix(stat.ContainerID, fmt.Sprintf("%s://", containerRuntimeManager.MYRuntimeName()))
	}
	fmt.Printf("containerId %s\n", containerId)
	containerRuntimeManager.MYContainerStatus(containerId)
	time.Sleep(time.Second * 2)
	cgroup, _ := containerRuntimeManager.getCgroupName(pod, containerId)
	time.Sleep(time.Second * 2)
	fmt.Printf("cgroup %s\n", cgroup)
	pids, _ := containerRuntimeManager.MYGetPidsInContainers(containerId)
	for _, pid := range pids {
		print(pid)
		print(" ")
	}
	//containerRuntimeManager.MYContainerStop(containerId)
	containerRuntimeManager.MYContainerUpdateResourc(containerId)
}
