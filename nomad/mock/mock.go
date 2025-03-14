package mock

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	fake "github.com/brianvoe/gofakeit/v6"
	"github.com/hashicorp/nomad/helper/envoy"
	"github.com/hashicorp/nomad/helper/pointer"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/nomad/structs"
	psstructs "github.com/hashicorp/nomad/plugins/shared/structs"
)

func Node() *structs.Node {
	node := &structs.Node{
		ID:         uuid.Generate(),
		SecretID:   uuid.Generate(),
		Datacenter: "dc1",
		Name:       "foobar",
		Drivers: map[string]*structs.DriverInfo{
			"exec": {
				Detected: true,
				Healthy:  true,
			},
			"mock_driver": {
				Detected: true,
				Healthy:  true,
			},
		},
		Attributes: map[string]string{
			"kernel.name":        "linux",
			"arch":               "x86",
			"nomad.version":      "0.5.0",
			"driver.exec":        "1",
			"driver.mock_driver": "1",
			"consul.version":     "1.11.4",
		},

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      4000,
			MemoryMB: 8192,
			DiskMB:   100 * 1024,
		},
		Reserved: &structs.Resources{
			CPU:      100,
			MemoryMB: 256,
			DiskMB:   4 * 1024,
			Networks: []*structs.NetworkResource{
				{
					Device:        "eth0",
					IP:            "192.168.0.100",
					ReservedPorts: []structs.Port{{Label: "ssh", Value: 22}},
					MBits:         1,
				},
			},
		},

		NodeResources: &structs.NodeResources{
			Cpu: structs.NodeCpuResources{
				CpuShares: 4000,
			},
			Memory: structs.NodeMemoryResources{
				MemoryMB: 8192,
			},
			Disk: structs.NodeDiskResources{
				DiskMB: 100 * 1024,
			},
			Networks: []*structs.NetworkResource{
				{
					Mode:   "host",
					Device: "eth0",
					CIDR:   "192.168.0.100/32",
					MBits:  1000,
				},
			},
			NodeNetworks: []*structs.NodeNetworkResource{
				{
					Mode:   "host",
					Device: "eth0",
					Speed:  1000,
					Addresses: []structs.NodeNetworkAddress{
						{
							Alias:   "default",
							Address: "192.168.0.100",
							Family:  structs.NodeNetworkAF_IPv4,
						},
					},
				},
			},
		},
		ReservedResources: &structs.NodeReservedResources{
			Cpu: structs.NodeReservedCpuResources{
				CpuShares: 100,
			},
			Memory: structs.NodeReservedMemoryResources{
				MemoryMB: 256,
			},
			Disk: structs.NodeReservedDiskResources{
				DiskMB: 4 * 1024,
			},
			Networks: structs.NodeReservedNetworkResources{
				ReservedHostPorts: "22",
			},
		},
		Links: map[string]string{
			"consul": "foobar.dc1",
		},
		Meta: map[string]string{
			"pci-dss":  "true",
			"database": "mysql",
			"version":  "5.6",
		},
		NodeClass:             "linux-medium-pci",
		Status:                structs.NodeStatusReady,
		SchedulingEligibility: structs.NodeSchedulingEligible,
	}
	node.ComputeClass()
	return node
}

func DrainNode() *structs.Node {
	node := Node()
	node.DrainStrategy = &structs.DrainStrategy{
		DrainSpec: structs.DrainSpec{},
	}
	node.Canonicalize()
	return node
}

// NvidiaNode returns a node with two instances of an Nvidia GPU
func NvidiaNode() *structs.Node {
	n := Node()
	n.NodeResources.Devices = []*structs.NodeDeviceResource{
		{
			Type:   "gpu",
			Vendor: "nvidia",
			Name:   "1080ti",
			Attributes: map[string]*psstructs.Attribute{
				"memory":           psstructs.NewIntAttribute(11, psstructs.UnitGiB),
				"cuda_cores":       psstructs.NewIntAttribute(3584, ""),
				"graphics_clock":   psstructs.NewIntAttribute(1480, psstructs.UnitMHz),
				"memory_bandwidth": psstructs.NewIntAttribute(11, psstructs.UnitGBPerS),
			},
			Instances: []*structs.NodeDevice{
				{
					ID:      uuid.Generate(),
					Healthy: true,
				},
				{
					ID:      uuid.Generate(),
					Healthy: true,
				},
			},
		},
	}
	n.ComputeClass()
	return n
}

func HCL() string {
	return `job "my-job" {
	datacenters = ["dc1"]
	type = "service"
	constraint {
		attribute = "${attr.kernel.name}"
		value = "linux"
	}

	group "web" {
		count = 10
		restart {
			attempts = 3
			interval = "10m"
			delay = "1m"
			mode = "delay"
		}
		task "web" {
			driver = "exec"
			config {
				command = "/bin/date"
			}
			resources {
				cpu = 500
				memory = 256
			}
		}
	}
}
`
}

func SystemBatchJob() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-sysbatch-%s", uuid.Short()),
		Name:        "my-sysbatch",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeSysBatch,
		Priority:    10,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{{
			Count: 1,
			Name:  "pinger",
			Tasks: []*structs.Task{{
				Name:   "ping-example",
				Driver: "exec",
				Config: map[string]interface{}{
					"command": "/usr/bin/ping",
					"args":    []string{"-c", "5", "example.com"},
				},
				LogConfig: structs.DefaultLogConfig(),
			}},
		}},

		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func Job() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-service-%s", uuid.Generate()),
		Name:        "my-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeService,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:  "web",
				Count: 10,
				Constraints: []*structs.Constraint{
					{
						LTarget: "${attr.consul.version}",
						RTarget: ">= 1.7.0",
						Operand: structs.ConstraintSemver,
					},
				},
				EphemeralDisk: &structs.EphemeralDisk{
					SizeMB: 150,
				},
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 3,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeDelay,
				},
				ReschedulePolicy: &structs.ReschedulePolicy{
					Attempts:      2,
					Interval:      10 * time.Minute,
					Delay:         5 * time.Second,
					DelayFunction: "constant",
				},
				Migrate: structs.DefaultMigrateStrategy(),
				Networks: []*structs.NetworkResource{
					{
						Mode: "host",
						DynamicPorts: []structs.Port{
							{Label: "http"},
							{Label: "admin"},
						},
					},
				},
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "exec",
						Config: map[string]interface{}{
							"command": "/bin/date",
						},
						Env: map[string]string{
							"FOO": "bar",
						},
						Services: []*structs.Service{
							{
								Name:      "${TASK}-frontend",
								PortLabel: "http",
								Tags:      []string{"pci:${meta.pci-dss}", "datacenter:${node.datacenter}"},
								Checks: []*structs.ServiceCheck{
									{
										Name:     "check-table",
										Type:     structs.ServiceCheckScript,
										Command:  "/usr/local/check-table-${meta.database}",
										Args:     []string{"${meta.version}"},
										Interval: 30 * time.Second,
										Timeout:  5 * time.Second,
									},
								},
							},
							{
								Name:      "${TASK}-admin",
								PortLabel: "admin",
							},
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      500,
							MemoryMB: 256,
						},
						Meta: map[string]string{
							"foo": "bar",
						},
					},
				},
				Meta: map[string]string{
					"elb_check_type":     "http",
					"elb_check_interval": "30s",
					"elb_check_min":      "3",
				},
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func MultiTaskGroupJob() *structs.Job {
	job := Job()
	apiTaskGroup := &structs.TaskGroup{
		Name:  "api",
		Count: 10,
		EphemeralDisk: &structs.EphemeralDisk{
			SizeMB: 150,
		},
		RestartPolicy: &structs.RestartPolicy{
			Attempts: 3,
			Interval: 10 * time.Minute,
			Delay:    1 * time.Minute,
			Mode:     structs.RestartPolicyModeDelay,
		},
		ReschedulePolicy: &structs.ReschedulePolicy{
			Attempts:      2,
			Interval:      10 * time.Minute,
			Delay:         5 * time.Second,
			DelayFunction: "constant",
		},
		Migrate: structs.DefaultMigrateStrategy(),
		Networks: []*structs.NetworkResource{
			{
				Mode: "host",
				DynamicPorts: []structs.Port{
					{Label: "http"},
					{Label: "admin"},
				},
			},
		},
		Tasks: []*structs.Task{
			{
				Name:   "api",
				Driver: "exec",
				Config: map[string]interface{}{
					"command": "/bin/date",
				},
				Env: map[string]string{
					"FOO": "bar",
				},
				Services: []*structs.Service{
					{
						Name:      "${TASK}-backend",
						PortLabel: "http",
						Tags:      []string{"pci:${meta.pci-dss}", "datacenter:${node.datacenter}"},
						Checks: []*structs.ServiceCheck{
							{
								Name:     "check-table",
								Type:     structs.ServiceCheckScript,
								Command:  "/usr/local/check-table-${meta.database}",
								Args:     []string{"${meta.version}"},
								Interval: 30 * time.Second,
								Timeout:  5 * time.Second,
							},
						},
					},
					{
						Name:      "${TASK}-admin",
						PortLabel: "admin",
					},
				},
				LogConfig: structs.DefaultLogConfig(),
				Resources: &structs.Resources{
					CPU:      500,
					MemoryMB: 256,
				},
				Meta: map[string]string{
					"foo": "bar",
				},
			},
		},
		Meta: map[string]string{
			"elb_check_type":     "http",
			"elb_check_interval": "30s",
			"elb_check_min":      "3",
		},
	}
	job.TaskGroups = append(job.TaskGroups, apiTaskGroup)
	job.Canonicalize()
	return job
}

func LifecycleSideTask(resources structs.Resources, i int) *structs.Task {
	return &structs.Task{
		Name:   fmt.Sprintf("side-%d", i),
		Driver: "exec",
		Config: map[string]interface{}{
			"command": "/bin/date",
		},
		Lifecycle: &structs.TaskLifecycleConfig{
			Hook:    structs.TaskLifecycleHookPrestart,
			Sidecar: true,
		},
		LogConfig: structs.DefaultLogConfig(),
		Resources: &resources,
	}
}

func LifecycleInitTask(resources structs.Resources, i int) *structs.Task {
	return &structs.Task{
		Name:   fmt.Sprintf("init-%d", i),
		Driver: "exec",
		Config: map[string]interface{}{
			"command": "/bin/date",
		},
		Lifecycle: &structs.TaskLifecycleConfig{
			Hook:    structs.TaskLifecycleHookPrestart,
			Sidecar: false,
		},
		LogConfig: structs.DefaultLogConfig(),
		Resources: &resources,
	}
}

func LifecycleMainTask(resources structs.Resources, i int) *structs.Task {
	return &structs.Task{
		Name:   fmt.Sprintf("main-%d", i),
		Driver: "exec",
		Config: map[string]interface{}{
			"command": "/bin/date",
		},
		LogConfig: structs.DefaultLogConfig(),
		Resources: &resources,
	}
}
func VariableLifecycleJob(resources structs.Resources, main int, init int, side int) *structs.Job {
	tasks := []*structs.Task{}
	for i := 0; i < main; i++ {
		tasks = append(tasks, LifecycleMainTask(resources, i))
	}
	for i := 0; i < init; i++ {
		tasks = append(tasks, LifecycleInitTask(resources, i))
	}
	for i := 0; i < side; i++ {
		tasks = append(tasks, LifecycleSideTask(resources, i))
	}
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-service-%s", uuid.Generate()),
		Name:        "my-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeService,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:  "web",
				Count: 1,
				Tasks: tasks,
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func LifecycleJob() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-service-%s", uuid.Generate()),
		Name:        "my-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeBatch,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:  "web",
				Count: 1,
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 0,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeFail,
				},
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "side",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPrestart,
							Sidecar: true,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "init",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPrestart,
							Sidecar: false,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "poststart",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPoststart,
							Sidecar: false,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
				},
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func LifecycleAlloc() *structs.Allocation {
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "web",

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      500,
			MemoryMB: 256,
		},
		TaskResources: map[string]*structs.Resources{
			"web": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"init": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"side": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"poststart": {
				CPU:      1000,
				MemoryMB: 256,
			},
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"web": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"init": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"side": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"poststart": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
			},
		},
		Job:           LifecycleJob(),
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

type LifecycleTaskDef struct {
	Name      string
	RunFor    string
	ExitCode  int
	Hook      string
	IsSidecar bool
}

// LifecycleAllocFromTasks generates an Allocation with mock tasks that have
// the provided lifecycles.
func LifecycleAllocFromTasks(tasks []LifecycleTaskDef) *structs.Allocation {
	alloc := LifecycleAlloc()
	alloc.Job.TaskGroups[0].Tasks = []*structs.Task{}
	for _, task := range tasks {
		var lc *structs.TaskLifecycleConfig
		if task.Hook != "" {
			// TODO: task coordinator doesn't treat nil and empty structs the same
			lc = &structs.TaskLifecycleConfig{
				Hook:    task.Hook,
				Sidecar: task.IsSidecar,
			}
		}

		alloc.Job.TaskGroups[0].Tasks = append(alloc.Job.TaskGroups[0].Tasks,
			&structs.Task{
				Name:   task.Name,
				Driver: "mock_driver",
				Config: map[string]interface{}{
					"run_for":   task.RunFor,
					"exit_code": task.ExitCode},
				Lifecycle: lc,
				LogConfig: structs.DefaultLogConfig(),
				Resources: &structs.Resources{CPU: 100, MemoryMB: 256},
			},
		)
		alloc.TaskResources[task.Name] = &structs.Resources{CPU: 100, MemoryMB: 256}
		alloc.AllocatedResources.Tasks[task.Name] = &structs.AllocatedTaskResources{
			Cpu:    structs.AllocatedCpuResources{CpuShares: 100},
			Memory: structs.AllocatedMemoryResources{MemoryMB: 256},
		}
	}
	return alloc
}

func LifecycleJobWithPoststopDeploy() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-service-%s", uuid.Generate()),
		Name:        "my-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeBatch,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:    "web",
				Count:   1,
				Migrate: structs.DefaultMigrateStrategy(),
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 0,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeFail,
				},
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "side",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPrestart,
							Sidecar: true,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "post",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook: structs.TaskLifecycleHookPoststop,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "init",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPrestart,
							Sidecar: false,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
				},
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func LifecycleJobWithPoststartDeploy() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-service-%s", uuid.Generate()),
		Name:        "my-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeBatch,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:    "web",
				Count:   1,
				Migrate: structs.DefaultMigrateStrategy(),
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 0,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeFail,
				},
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "side",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPrestart,
							Sidecar: true,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "post",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook: structs.TaskLifecycleHookPoststart,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
					{
						Name:   "init",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "1s",
						},
						Lifecycle: &structs.TaskLifecycleConfig{
							Hook:    structs.TaskLifecycleHookPrestart,
							Sidecar: false,
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      1000,
							MemoryMB: 256,
						},
					},
				},
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func LifecycleAllocWithPoststopDeploy() *structs.Allocation {
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "web",

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      500,
			MemoryMB: 256,
		},
		TaskResources: map[string]*structs.Resources{
			"web": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"init": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"side": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"post": {
				CPU:      1000,
				MemoryMB: 256,
			},
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"web": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"init": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"side": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"post": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
			},
		},
		Job:           LifecycleJobWithPoststopDeploy(),
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

func LifecycleAllocWithPoststartDeploy() *structs.Allocation {
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789xyz",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "web",

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      500,
			MemoryMB: 256,
		},
		TaskResources: map[string]*structs.Resources{
			"web": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"init": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"side": {
				CPU:      1000,
				MemoryMB: 256,
			},
			"post": {
				CPU:      1000,
				MemoryMB: 256,
			},
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"web": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"init": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"side": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
				"post": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 1000,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
				},
			},
		},
		Job:           LifecycleJobWithPoststartDeploy(),
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

func MaxParallelJob() *structs.Job {
	update := *structs.DefaultUpdateStrategy
	update.MaxParallel = 0
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-service-%s", uuid.Generate()),
		Name:        "my-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeService,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		Update: update,
		TaskGroups: []*structs.TaskGroup{
			{
				Name:  "web",
				Count: 10,
				EphemeralDisk: &structs.EphemeralDisk{
					SizeMB: 150,
				},
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 3,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeDelay,
				},
				ReschedulePolicy: &structs.ReschedulePolicy{
					Attempts:      2,
					Interval:      10 * time.Minute,
					Delay:         5 * time.Second,
					DelayFunction: "constant",
				},
				Migrate: structs.DefaultMigrateStrategy(),
				Update:  &update,
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "exec",
						Config: map[string]interface{}{
							"command": "/bin/date",
						},
						Env: map[string]string{
							"FOO": "bar",
						},
						Services: []*structs.Service{
							{
								Name:      "${TASK}-frontend",
								PortLabel: "http",
								Tags:      []string{"pci:${meta.pci-dss}", "datacenter:${node.datacenter}"},
								Checks: []*structs.ServiceCheck{
									{
										Name:     "check-table",
										Type:     structs.ServiceCheckScript,
										Command:  "/usr/local/check-table-${meta.database}",
										Args:     []string{"${meta.version}"},
										Interval: 30 * time.Second,
										Timeout:  5 * time.Second,
									},
								},
							},
							{
								Name:      "${TASK}-admin",
								PortLabel: "admin",
							},
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      500,
							MemoryMB: 256,
							Networks: []*structs.NetworkResource{
								{
									MBits: 50,
									DynamicPorts: []structs.Port{
										{Label: "http"},
										{Label: "admin"},
									},
								},
							},
						},
						Meta: map[string]string{
							"foo": "bar",
						},
					},
				},
				Meta: map[string]string{
					"elb_check_type":     "http",
					"elb_check_interval": "30s",
					"elb_check_min":      "3",
				},
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

// ConnectJob adds a Connect proxy sidecar group service to mock.Job.
//
// Note this does *not* include the Job.Register mutation that inserts the
// associated Sidecar Task (nor the hook that configures envoy as the default).
func ConnectJob() *structs.Job {
	job := Job()
	tg := job.TaskGroups[0]
	tg.Services = []*structs.Service{{
		Name:      "testconnect",
		PortLabel: "9999",
		Connect: &structs.ConsulConnect{
			SidecarService: new(structs.ConsulSidecarService),
		},
	}}
	tg.Networks = structs.Networks{{
		Mode: "bridge", // always bridge ... for now?
	}}
	return job
}

func ConnectNativeJob(mode string) *structs.Job {
	job := Job()
	tg := job.TaskGroups[0]
	tg.Networks = []*structs.NetworkResource{{
		Mode: mode,
	}}
	tg.Services = []*structs.Service{{
		Name:      "test_connect_native",
		PortLabel: "9999",
		Connect: &structs.ConsulConnect{
			Native: true,
		},
	}}
	tg.Tasks = []*structs.Task{{
		Name: "native_task",
	}}
	return job
}

// ConnectIngressGatewayJob creates a structs.Job that contains the definition
// of a Consul Ingress Gateway service. The mode is the name of the network
// mode assumed by the task group. If inject is true, a corresponding Task is
// set on the group's Tasks (i.e. what the job would look like after job mutation).
func ConnectIngressGatewayJob(mode string, inject bool) *structs.Job {
	job := Job()
	tg := job.TaskGroups[0]
	tg.Networks = []*structs.NetworkResource{{
		Mode: mode,
	}}
	tg.Services = []*structs.Service{{
		Name:      "my-ingress-service",
		PortLabel: "9999",
		Connect: &structs.ConsulConnect{
			Gateway: &structs.ConsulGateway{
				Proxy: &structs.ConsulGatewayProxy{
					ConnectTimeout:            pointer.Of(3 * time.Second),
					EnvoyGatewayBindAddresses: make(map[string]*structs.ConsulGatewayBindAddress),
				},
				Ingress: &structs.ConsulIngressConfigEntry{
					Listeners: []*structs.ConsulIngressListener{{
						Port:     2000,
						Protocol: "tcp",
						Services: []*structs.ConsulIngressService{{
							Name: "service1",
						}},
					}},
				},
			},
		},
	}}

	tg.Tasks = nil

	// some tests need to assume the gateway proxy task has already been injected
	if inject {
		tg.Tasks = []*structs.Task{{
			Name:          fmt.Sprintf("%s-%s", structs.ConnectIngressPrefix, "my-ingress-service"),
			Kind:          structs.NewTaskKind(structs.ConnectIngressPrefix, "my-ingress-service"),
			Driver:        "docker",
			Config:        make(map[string]interface{}),
			ShutdownDelay: 5 * time.Second,
			LogConfig: &structs.LogConfig{
				MaxFiles:      2,
				MaxFileSizeMB: 2,
			},
		}}
	}
	return job
}

// ConnectTerminatingGatewayJob creates a structs.Job that contains the definition
// of a Consul Terminating Gateway service. The mode is the name of the network mode
// assumed by the task group. If inject is true, a corresponding task is set on the
// group's Tasks (i.e. what the job would look like after mutation).
func ConnectTerminatingGatewayJob(mode string, inject bool) *structs.Job {
	job := Job()
	tg := job.TaskGroups[0]
	tg.Networks = []*structs.NetworkResource{{
		Mode: mode,
	}}
	tg.Services = []*structs.Service{{
		Name:      "my-terminating-service",
		PortLabel: "9999",
		Connect: &structs.ConsulConnect{
			Gateway: &structs.ConsulGateway{
				Proxy: &structs.ConsulGatewayProxy{
					ConnectTimeout:            pointer.Of(3 * time.Second),
					EnvoyGatewayBindAddresses: make(map[string]*structs.ConsulGatewayBindAddress),
				},
				Terminating: &structs.ConsulTerminatingConfigEntry{
					Services: []*structs.ConsulLinkedService{{
						Name:     "service1",
						CAFile:   "/ssl/ca_file",
						CertFile: "/ssl/cert_file",
						KeyFile:  "/ssl/key_file",
						SNI:      "sni-name",
					}},
				},
			},
		},
	}}

	tg.Tasks = nil

	// some tests need to assume the gateway proxy task has already been injected
	if inject {
		tg.Tasks = []*structs.Task{{
			Name:          fmt.Sprintf("%s-%s", structs.ConnectTerminatingPrefix, "my-terminating-service"),
			Kind:          structs.NewTaskKind(structs.ConnectTerminatingPrefix, "my-terminating-service"),
			Driver:        "docker",
			Config:        make(map[string]interface{}),
			ShutdownDelay: 5 * time.Second,
			LogConfig: &structs.LogConfig{
				MaxFiles:      2,
				MaxFileSizeMB: 2,
			},
		}}
	}
	return job
}

// ConnectMeshGatewayJob creates a structs.Job that contains the definition of a
// Consul Mesh Gateway service. The mode is the name of the network mode assumed
// by the task group. If inject is true, a corresponding task is set on the group's
// Tasks (i.e. what the job would look like after job mutation).
func ConnectMeshGatewayJob(mode string, inject bool) *structs.Job {
	job := Job()
	tg := job.TaskGroups[0]
	tg.Networks = []*structs.NetworkResource{{
		Mode: mode,
	}}
	tg.Services = []*structs.Service{{
		Name:      "my-mesh-service",
		PortLabel: "public_port",
		Connect: &structs.ConsulConnect{
			Gateway: &structs.ConsulGateway{
				Proxy: &structs.ConsulGatewayProxy{
					ConnectTimeout:            pointer.Of(3 * time.Second),
					EnvoyGatewayBindAddresses: make(map[string]*structs.ConsulGatewayBindAddress),
				},
				Mesh: &structs.ConsulMeshConfigEntry{
					// nothing to configure
				},
			},
		},
	}}

	tg.Tasks = nil

	// some tests need to assume the gateway task has already been injected
	if inject {
		tg.Tasks = []*structs.Task{{
			Name:          fmt.Sprintf("%s-%s", structs.ConnectMeshPrefix, "my-mesh-service"),
			Kind:          structs.NewTaskKind(structs.ConnectMeshPrefix, "my-mesh-service"),
			Driver:        "docker",
			Config:        make(map[string]interface{}),
			ShutdownDelay: 5 * time.Second,
			LogConfig: &structs.LogConfig{
				MaxFiles:      2,
				MaxFileSizeMB: 2,
			},
		}}
	}
	return job
}

func ConnectSidecarTask() *structs.Task {
	return &structs.Task{
		Name:   "mysidecar-sidecar-task",
		Driver: "docker",
		User:   "nobody",
		Config: map[string]interface{}{
			"image": envoy.SidecarConfigVar,
		},
		Env: nil,
		Resources: &structs.Resources{
			CPU:      150,
			MemoryMB: 350,
		},
		Kind: structs.NewTaskKind(structs.ConnectProxyPrefix, "mysidecar"),
	}
}

func BatchJob() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-batch-%s", uuid.Generate()),
		Name:        "batch-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeBatch,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:  "web",
				Count: 10,
				EphemeralDisk: &structs.EphemeralDisk{
					SizeMB: 150,
				},
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 3,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeDelay,
				},
				ReschedulePolicy: &structs.ReschedulePolicy{
					Attempts:      2,
					Interval:      10 * time.Minute,
					Delay:         5 * time.Second,
					DelayFunction: "constant",
				},
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "mock_driver",
						Config: map[string]interface{}{
							"run_for": "500ms",
						},
						Env: map[string]string{
							"FOO": "bar",
						},
						LogConfig: structs.DefaultLogConfig(),
						Resources: &structs.Resources{
							CPU:      100,
							MemoryMB: 100,
							Networks: []*structs.NetworkResource{
								{
									MBits: 50,
								},
							},
						},
						Meta: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
		},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    43,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func SystemJob() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		Namespace:   structs.DefaultNamespace,
		ID:          fmt.Sprintf("mock-system-%s", uuid.Generate()),
		Name:        "my-job",
		Type:        structs.JobTypeSystem,
		Priority:    100,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		Constraints: []*structs.Constraint{
			{
				LTarget: "${attr.kernel.name}",
				RTarget: "linux",
				Operand: "=",
			},
		},
		TaskGroups: []*structs.TaskGroup{
			{
				Name:  "web",
				Count: 1,
				RestartPolicy: &structs.RestartPolicy{
					Attempts: 3,
					Interval: 10 * time.Minute,
					Delay:    1 * time.Minute,
					Mode:     structs.RestartPolicyModeDelay,
				},
				EphemeralDisk: structs.DefaultEphemeralDisk(),
				Tasks: []*structs.Task{
					{
						Name:   "web",
						Driver: "exec",
						Config: map[string]interface{}{
							"command": "/bin/date",
						},
						Env: map[string]string{},
						Resources: &structs.Resources{
							CPU:      500,
							MemoryMB: 256,
							Networks: []*structs.NetworkResource{
								{
									MBits:        50,
									DynamicPorts: []structs.Port{{Label: "http"}},
								},
							},
						},
						LogConfig: structs.DefaultLogConfig(),
					},
				},
			},
		},
		Meta: map[string]string{
			"owner": "armon",
		},
		Status:      structs.JobStatusPending,
		CreateIndex: 42,
		ModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

func PeriodicJob() *structs.Job {
	job := Job()
	job.Type = structs.JobTypeBatch
	job.Periodic = &structs.PeriodicConfig{
		Enabled:  true,
		SpecType: structs.PeriodicSpecCron,
		Spec:     "*/30 * * * *",
	}
	job.Status = structs.JobStatusRunning
	job.TaskGroups[0].Migrate = nil
	return job
}

func Eval() *structs.Evaluation {
	now := time.Now().UTC().UnixNano()
	eval := &structs.Evaluation{
		ID:         uuid.Generate(),
		Namespace:  structs.DefaultNamespace,
		Priority:   50,
		Type:       structs.JobTypeService,
		JobID:      uuid.Generate(),
		Status:     structs.EvalStatusPending,
		CreateTime: now,
		ModifyTime: now,
	}
	return eval
}

func BlockedEval() *structs.Evaluation {
	e := Eval()
	e.Status = structs.EvalStatusBlocked
	e.FailedTGAllocs = map[string]*structs.AllocMetric{
		"cache": {
			DimensionExhausted: map[string]int{
				"memory": 1,
			},
			ResourcesExhausted: map[string]*structs.Resources{
				"redis": {
					CPU:      100,
					MemoryMB: 1024,
				},
			},
		},
	}

	return e
}

func JobSummary(jobID string) *structs.JobSummary {
	return &structs.JobSummary{
		JobID:     jobID,
		Namespace: structs.DefaultNamespace,
		Summary: map[string]structs.TaskGroupSummary{
			"web": {
				Queued:   0,
				Starting: 0,
			},
		},
	}
}

func JobSysBatchSummary(jobID string) *structs.JobSummary {
	return &structs.JobSummary{
		JobID:     jobID,
		Namespace: structs.DefaultNamespace,
		Summary: map[string]structs.TaskGroupSummary{
			"pinger": {
				Queued:   0,
				Starting: 0,
			},
		},
	}
}

func Alloc() *structs.Allocation {
	job := Job()
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "web",

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      500,
			MemoryMB: 256,
			DiskMB:   150,
			Networks: []*structs.NetworkResource{
				{
					Device:        "eth0",
					IP:            "192.168.0.100",
					ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
					MBits:         50,
					DynamicPorts:  []structs.Port{{Label: "http"}},
				},
			},
		},
		TaskResources: map[string]*structs.Resources{
			"web": {
				CPU:      500,
				MemoryMB: 256,
				Networks: []*structs.NetworkResource{
					{
						Device:        "eth0",
						IP:            "192.168.0.100",
						ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
						MBits:         50,
						DynamicPorts:  []structs.Port{{Label: "http", Value: 9876}},
					},
				},
			},
		},
		SharedResources: &structs.Resources{
			DiskMB: 150,
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"web": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 500,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
					Networks: []*structs.NetworkResource{
						{
							Device:        "eth0",
							IP:            "192.168.0.100",
							ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
							MBits:         50,
							DynamicPorts:  []structs.Port{{Label: "http", Value: 9876}},
						},
					},
				},
			},
			Shared: structs.AllocatedSharedResources{
				DiskMB: 150,
			},
		},
		Job:           job,
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

func AllocWithoutReservedPort() *structs.Allocation {
	alloc := Alloc()
	alloc.Resources.Networks[0].ReservedPorts = nil
	alloc.TaskResources["web"].Networks[0].ReservedPorts = nil
	alloc.AllocatedResources.Tasks["web"].Networks[0].ReservedPorts = nil

	return alloc
}

func AllocForNode(n *structs.Node) *structs.Allocation {
	nodeIP := n.NodeResources.NodeNetworks[0].Addresses[0].Address

	dynamicPortRange := structs.DefaultMaxDynamicPort - structs.DefaultMinDynamicPort
	randomDynamicPort := rand.Intn(dynamicPortRange) + structs.DefaultMinDynamicPort

	alloc := Alloc()
	alloc.NodeID = n.ID

	// Set node IP address.
	alloc.Resources.Networks[0].IP = nodeIP
	alloc.TaskResources["web"].Networks[0].IP = nodeIP
	alloc.AllocatedResources.Tasks["web"].Networks[0].IP = nodeIP

	// Set dynamic port to a random value.
	alloc.TaskResources["web"].Networks[0].DynamicPorts = []structs.Port{{Label: "http", Value: randomDynamicPort}}
	alloc.AllocatedResources.Tasks["web"].Networks[0].DynamicPorts = []structs.Port{{Label: "http", Value: randomDynamicPort}}

	return alloc

}

func AllocForNodeWithoutReservedPort(n *structs.Node) *structs.Allocation {
	nodeIP := n.NodeResources.NodeNetworks[0].Addresses[0].Address

	dynamicPortRange := structs.DefaultMaxDynamicPort - structs.DefaultMinDynamicPort
	randomDynamicPort := rand.Intn(dynamicPortRange) + structs.DefaultMinDynamicPort

	alloc := AllocWithoutReservedPort()
	alloc.NodeID = n.ID

	// Set node IP address.
	alloc.Resources.Networks[0].IP = nodeIP
	alloc.TaskResources["web"].Networks[0].IP = nodeIP
	alloc.AllocatedResources.Tasks["web"].Networks[0].IP = nodeIP

	// Set dynamic port to a random value.
	alloc.TaskResources["web"].Networks[0].DynamicPorts = []structs.Port{{Label: "http", Value: randomDynamicPort}}
	alloc.AllocatedResources.Tasks["web"].Networks[0].DynamicPorts = []structs.Port{{Label: "http", Value: randomDynamicPort}}

	return alloc
}

// ConnectAlloc adds a Connect proxy sidecar group service to mock.Alloc.
func ConnectAlloc() *structs.Allocation {
	alloc := Alloc()
	alloc.Job = ConnectJob()
	alloc.AllocatedResources.Shared.Networks = []*structs.NetworkResource{
		{
			Mode: "bridge",
			IP:   "10.0.0.1",
			DynamicPorts: []structs.Port{
				{
					Label: "connect-proxy-testconnect",
					Value: 9999,
					To:    9999,
				},
			},
		},
	}
	return alloc
}

// ConnectNativeAlloc creates an alloc with a connect native task.
func ConnectNativeAlloc(mode string) *structs.Allocation {
	alloc := Alloc()
	alloc.Job = ConnectNativeJob(mode)
	alloc.AllocatedResources.Shared.Networks = []*structs.NetworkResource{{
		Mode: mode,
		IP:   "10.0.0.1",
	}}
	return alloc
}

func ConnectIngressGatewayAlloc(mode string) *structs.Allocation {
	alloc := Alloc()
	alloc.Job = ConnectIngressGatewayJob(mode, true)
	alloc.AllocatedResources.Shared.Networks = []*structs.NetworkResource{{
		Mode: mode,
		IP:   "10.0.0.1",
	}}
	return alloc
}

func BatchConnectJob() *structs.Job {
	job := &structs.Job{
		Region:      "global",
		ID:          fmt.Sprintf("mock-connect-batch-job%s", uuid.Generate()),
		Name:        "mock-connect-batch-job",
		Namespace:   structs.DefaultNamespace,
		Type:        structs.JobTypeBatch,
		Priority:    50,
		AllAtOnce:   false,
		Datacenters: []string{"dc1"},
		TaskGroups: []*structs.TaskGroup{{
			Name:          "mock-connect-batch-job",
			Count:         1,
			EphemeralDisk: &structs.EphemeralDisk{SizeMB: 150},
			Networks: []*structs.NetworkResource{{
				Mode: "bridge",
			}},
			Tasks: []*structs.Task{{
				Name:   "connect-proxy-testconnect",
				Kind:   "connect-proxy:testconnect",
				Driver: "mock_driver",
				Config: map[string]interface{}{
					"run_for": "500ms",
				},
				LogConfig: structs.DefaultLogConfig(),
				Resources: &structs.Resources{
					CPU:      500,
					MemoryMB: 256,
					Networks: []*structs.NetworkResource{{
						MBits:        50,
						DynamicPorts: []structs.Port{{Label: "port1"}},
					}},
				},
			}},
			Services: []*structs.Service{{
				Name: "testconnect",
			}},
		}},
		Meta:           map[string]string{"owner": "shoenig"},
		Status:         structs.JobStatusPending,
		Version:        0,
		CreateIndex:    42,
		ModifyIndex:    99,
		JobModifyIndex: 99,
	}
	job.Canonicalize()
	return job
}

// BatchConnectAlloc is useful for testing task runner things.
func BatchConnectAlloc() *structs.Allocation {
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "mock-connect-batch-job",
		TaskResources: map[string]*structs.Resources{
			"connect-proxy-testconnect": {
				CPU:      500,
				MemoryMB: 256,
			},
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"connect-proxy-testconnect": {
					Cpu:    structs.AllocatedCpuResources{CpuShares: 500},
					Memory: structs.AllocatedMemoryResources{MemoryMB: 256},
				},
			},
			Shared: structs.AllocatedSharedResources{
				Networks: []*structs.NetworkResource{{
					Mode: "bridge",
					IP:   "10.0.0.1",
					DynamicPorts: []structs.Port{{
						Label: "connect-proxy-testconnect",
						Value: 9999,
						To:    9999,
					}},
				}},
				DiskMB: 0,
			},
		},
		Job:           BatchConnectJob(),
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

func BatchAlloc() *structs.Allocation {
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "web",

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      500,
			MemoryMB: 256,
			DiskMB:   150,
			Networks: []*structs.NetworkResource{
				{
					Device:        "eth0",
					IP:            "192.168.0.100",
					ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
					MBits:         50,
					DynamicPorts:  []structs.Port{{Label: "http"}},
				},
			},
		},
		TaskResources: map[string]*structs.Resources{
			"web": {
				CPU:      500,
				MemoryMB: 256,
				Networks: []*structs.NetworkResource{
					{
						Device:        "eth0",
						IP:            "192.168.0.100",
						ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
						MBits:         50,
						DynamicPorts:  []structs.Port{{Label: "http", Value: 9876}},
					},
				},
			},
		},
		SharedResources: &structs.Resources{
			DiskMB: 150,
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"web": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 500,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
					Networks: []*structs.NetworkResource{
						{
							Device:        "eth0",
							IP:            "192.168.0.100",
							ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
							MBits:         50,
							DynamicPorts:  []structs.Port{{Label: "http", Value: 9876}},
						},
					},
				},
			},
			Shared: structs.AllocatedSharedResources{
				DiskMB: 150,
			},
		},
		Job:           BatchJob(),
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

func SysBatchAlloc() *structs.Allocation {
	job := SystemBatchJob()
	return &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "pinger",
		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"ping-example": {
					Cpu:    structs.AllocatedCpuResources{CpuShares: 500},
					Memory: structs.AllocatedMemoryResources{MemoryMB: 256},
					Networks: []*structs.NetworkResource{{
						Device: "eth0",
						IP:     "192.168.0.100",
					}},
				},
			},
			Shared: structs.AllocatedSharedResources{DiskMB: 150},
		},
		Job:           job,
		JobID:         job.ID,
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
}

func SystemAlloc() *structs.Allocation {
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		EvalID:    uuid.Generate(),
		NodeID:    "12345678-abcd-efab-cdef-123456789abc",
		Namespace: structs.DefaultNamespace,
		TaskGroup: "web",

		// TODO Remove once clientv2 gets merged
		Resources: &structs.Resources{
			CPU:      500,
			MemoryMB: 256,
			DiskMB:   150,
			Networks: []*structs.NetworkResource{
				{
					Device:        "eth0",
					IP:            "192.168.0.100",
					ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
					MBits:         50,
					DynamicPorts:  []structs.Port{{Label: "http"}},
				},
			},
		},
		TaskResources: map[string]*structs.Resources{
			"web": {
				CPU:      500,
				MemoryMB: 256,
				Networks: []*structs.NetworkResource{
					{
						Device:        "eth0",
						IP:            "192.168.0.100",
						ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
						MBits:         50,
						DynamicPorts:  []structs.Port{{Label: "http", Value: 9876}},
					},
				},
			},
		},
		SharedResources: &structs.Resources{
			DiskMB: 150,
		},

		AllocatedResources: &structs.AllocatedResources{
			Tasks: map[string]*structs.AllocatedTaskResources{
				"web": {
					Cpu: structs.AllocatedCpuResources{
						CpuShares: 500,
					},
					Memory: structs.AllocatedMemoryResources{
						MemoryMB: 256,
					},
					Networks: []*structs.NetworkResource{
						{
							Device:        "eth0",
							IP:            "192.168.0.100",
							ReservedPorts: []structs.Port{{Label: "admin", Value: 5000}},
							MBits:         50,
							DynamicPorts:  []structs.Port{{Label: "http", Value: 9876}},
						},
					},
				},
			},
			Shared: structs.AllocatedSharedResources{
				DiskMB: 150,
			},
		},
		Job:           SystemJob(),
		DesiredStatus: structs.AllocDesiredStatusRun,
		ClientStatus:  structs.AllocClientStatusPending,
	}
	alloc.JobID = alloc.Job.ID
	return alloc
}

func VaultAccessor() *structs.VaultAccessor {
	return &structs.VaultAccessor{
		Accessor:    uuid.Generate(),
		NodeID:      uuid.Generate(),
		AllocID:     uuid.Generate(),
		CreationTTL: 86400,
		Task:        "foo",
	}
}

func SITokenAccessor() *structs.SITokenAccessor {
	return &structs.SITokenAccessor{
		NodeID:     uuid.Generate(),
		AllocID:    uuid.Generate(),
		AccessorID: uuid.Generate(),
		TaskName:   "foo",
	}
}

func Deployment() *structs.Deployment {
	return &structs.Deployment{
		ID:             uuid.Generate(),
		JobID:          uuid.Generate(),
		Namespace:      structs.DefaultNamespace,
		JobVersion:     2,
		JobModifyIndex: 20,
		JobCreateIndex: 18,
		TaskGroups: map[string]*structs.DeploymentState{
			"web": {
				DesiredTotal: 10,
			},
		},
		Status:            structs.DeploymentStatusRunning,
		StatusDescription: structs.DeploymentStatusDescriptionRunning,
		ModifyIndex:       23,
		CreateIndex:       21,
	}
}

func Plan() *structs.Plan {
	return &structs.Plan{
		Priority: 50,
	}
}

func PlanResult() *structs.PlanResult {
	return &structs.PlanResult{}
}

func ACLPolicy() *structs.ACLPolicy {
	ap := &structs.ACLPolicy{
		Name:        fmt.Sprintf("policy-%s", uuid.Generate()),
		Description: "Super cool policy!",
		Rules: `
		namespace "default" {
			policy = "write"
		}
		node {
			policy = "read"
		}
		agent {
			policy = "read"
		}
		`,
		CreateIndex: 10,
		ModifyIndex: 20,
	}
	ap.SetHash()
	return ap
}

func ACLToken() *structs.ACLToken {
	tk := &structs.ACLToken{
		AccessorID:  uuid.Generate(),
		SecretID:    uuid.Generate(),
		Name:        "my cool token " + uuid.Generate(),
		Type:        "client",
		Policies:    []string{"foo", "bar"},
		Global:      false,
		CreateTime:  time.Now().UTC(),
		CreateIndex: 10,
		ModifyIndex: 20,
	}
	tk.SetHash()
	return tk
}

func ACLManagementToken() *structs.ACLToken {
	return &structs.ACLToken{
		AccessorID:  uuid.Generate(),
		SecretID:    uuid.Generate(),
		Name:        "management " + uuid.Generate(),
		Type:        "management",
		Global:      true,
		CreateTime:  time.Now().UTC(),
		CreateIndex: 10,
		ModifyIndex: 20,
	}
}

func ScalingPolicy() *structs.ScalingPolicy {
	return &structs.ScalingPolicy{
		ID:   uuid.Generate(),
		Min:  1,
		Max:  100,
		Type: structs.ScalingPolicyTypeHorizontal,
		Target: map[string]string{
			structs.ScalingTargetNamespace: structs.DefaultNamespace,
			structs.ScalingTargetJob:       uuid.Generate(),
			structs.ScalingTargetGroup:     uuid.Generate(),
			structs.ScalingTargetTask:      uuid.Generate(),
		},
		Policy: map[string]interface{}{
			"a": "b",
		},
		Enabled: true,
	}
}

func JobWithScalingPolicy() (*structs.Job, *structs.ScalingPolicy) {
	job := Job()
	policy := &structs.ScalingPolicy{
		ID:      uuid.Generate(),
		Min:     int64(job.TaskGroups[0].Count),
		Max:     int64(job.TaskGroups[0].Count),
		Type:    structs.ScalingPolicyTypeHorizontal,
		Policy:  map[string]interface{}{},
		Enabled: true,
	}
	policy.TargetTaskGroup(job, job.TaskGroups[0])
	job.TaskGroups[0].Scaling = policy
	return job, policy
}

func MultiregionJob() *structs.Job {
	job := Job()
	update := *structs.DefaultUpdateStrategy
	job.Update = update
	job.TaskGroups[0].Update = &update
	job.Multiregion = &structs.Multiregion{
		Strategy: &structs.MultiregionStrategy{
			MaxParallel: 1,
			OnFailure:   "fail_all",
		},
		Regions: []*structs.MultiregionRegion{
			{
				Name:        "west",
				Count:       2,
				Datacenters: []string{"west-1", "west-2"},
				Meta:        map[string]string{"region_code": "W"},
			},
			{
				Name:        "east",
				Count:       1,
				Datacenters: []string{"east-1"},
				Meta:        map[string]string{"region_code": "E"},
			},
		},
	}
	return job
}

func CSIPlugin() *structs.CSIPlugin {
	return &structs.CSIPlugin{
		ID:                 uuid.Generate(),
		Provider:           "com.hashicorp:mock",
		Version:            "0.1",
		ControllerRequired: true,
		Controllers:        map[string]*structs.CSIInfo{},
		Nodes:              map[string]*structs.CSIInfo{},
		Allocations:        []*structs.AllocListStub{},
		ControllersHealthy: 0,
		NodesHealthy:       0,
	}
}

func CSIVolume(plugin *structs.CSIPlugin) *structs.CSIVolume {
	return &structs.CSIVolume{
		ID:                  uuid.Generate(),
		Name:                "test-vol",
		ExternalID:          "vol-01",
		Namespace:           "default",
		Topologies:          []*structs.CSITopology{},
		AccessMode:          structs.CSIVolumeAccessModeUnknown,
		AttachmentMode:      structs.CSIVolumeAttachmentModeUnknown,
		MountOptions:        &structs.CSIMountOptions{},
		Secrets:             structs.CSISecrets{},
		Parameters:          map[string]string{},
		Context:             map[string]string{},
		ReadAllocs:          map[string]*structs.Allocation{},
		WriteAllocs:         map[string]*structs.Allocation{},
		ReadClaims:          map[string]*structs.CSIVolumeClaim{},
		WriteClaims:         map[string]*structs.CSIVolumeClaim{},
		PastClaims:          map[string]*structs.CSIVolumeClaim{},
		PluginID:            plugin.ID,
		Provider:            plugin.Provider,
		ProviderVersion:     plugin.Version,
		ControllerRequired:  plugin.ControllerRequired,
		ControllersHealthy:  plugin.ControllersHealthy,
		ControllersExpected: len(plugin.Controllers),
		NodesHealthy:        plugin.NodesHealthy,
		NodesExpected:       len(plugin.Nodes),
	}
}

func CSIPluginJob(pluginType structs.CSIPluginType, pluginID string) *structs.Job {

	job := new(structs.Job)

	switch pluginType {
	case structs.CSIPluginTypeController:
		job = Job()
		job.ID = fmt.Sprintf("mock-controller-%s", pluginID)
		job.Name = "job-plugin-controller"
		job.TaskGroups[0].Count = 2
	case structs.CSIPluginTypeNode:
		job = SystemJob()
		job.ID = fmt.Sprintf("mock-node-%s", pluginID)
		job.Name = "job-plugin-node"
	case structs.CSIPluginTypeMonolith:
		job = SystemJob()
		job.ID = fmt.Sprintf("mock-monolith-%s", pluginID)
		job.Name = "job-plugin-monolith"
	}

	job.TaskGroups[0].Name = "plugin"
	job.TaskGroups[0].Tasks[0].Name = "plugin"
	job.TaskGroups[0].Tasks[0].Driver = "docker"
	job.TaskGroups[0].Tasks[0].Services = nil
	job.TaskGroups[0].Tasks[0].CSIPluginConfig = &structs.TaskCSIPluginConfig{
		ID:       pluginID,
		Type:     pluginType,
		MountDir: "/csi",
	}
	job.Canonicalize()
	return job
}

func Events(index uint64) *structs.Events {
	return &structs.Events{
		Index: index,
		Events: []structs.Event{
			{
				Index:   index,
				Topic:   "Node",
				Type:    "update",
				Key:     uuid.Generate(),
				Payload: Node(),
			},
			{
				Index:   index,
				Topic:   "Eval",
				Type:    "update",
				Key:     uuid.Generate(),
				Payload: Eval(),
			},
		},
	}
}

func AllocNetworkStatus() *structs.AllocNetworkStatus {
	return &structs.AllocNetworkStatus{
		InterfaceName: "eth0",
		Address:       "192.168.0.100",
		DNS: &structs.DNSConfig{
			Servers:  []string{"1.1.1.1"},
			Searches: []string{"localdomain"},
			Options:  []string{"ndots:5"},
		},
	}
}

func Namespace() *structs.Namespace {
	uuid := uuid.Generate()
	ns := &structs.Namespace{
		Name:        fmt.Sprintf("team-%s", uuid),
		Meta:        map[string]string{"team": uuid},
		Description: "test namespace",
		CreateIndex: 100,
		ModifyIndex: 200,
	}
	ns.SetHash()
	return ns
}

// ServiceRegistrations generates an array containing two unique service
// registrations.
func ServiceRegistrations() []*structs.ServiceRegistration {
	return []*structs.ServiceRegistration{
		{
			ID:          "_nomad-task-2873cf75-42e5-7c45-ca1c-415f3e18be3d-group-cache-example-cache-db",
			ServiceName: "example-cache",
			Namespace:   "default",
			NodeID:      "17a6d1c0-811e-2ca9-ded0-3d5d6a54904c",
			Datacenter:  "dc1",
			JobID:       "example",
			AllocID:     "2873cf75-42e5-7c45-ca1c-415f3e18be3d",
			Tags:        []string{"foo"},
			Address:     "192.168.10.1",
			Port:        23000,
		},
		{
			ID:          "_nomad-task-ca60e901-675a-0ab2-2e57-2f3b05fdc540-group-api-countdash-api-http",
			ServiceName: "countdash-api",
			Namespace:   "platform",
			NodeID:      "ba991c17-7ce5-9c20-78b7-311e63578583",
			Datacenter:  "dc2",
			JobID:       "countdash-api",
			AllocID:     "ca60e901-675a-0ab2-2e57-2f3b05fdc540",
			Tags:        []string{"bar"},
			Address:     "192.168.200.200",
			Port:        29000,
		},
	}
}

type MockVariables map[string]*structs.VariableDecrypted

func Variable() *structs.VariableDecrypted {
	return &structs.VariableDecrypted{
		VariableMetadata: mockVariableMetadata(),
		Items: structs.VariableItems{
			"key1": "value1",
			"key2": "value2",
		},
	}
}

// Variables returns a random number of variables between min
// and max inclusive.
func Variables(minU, maxU uint8) MockVariables {
	// the unsignedness of the args is to prevent goofy parameters, they're
	// easier to work with as ints in this code.
	min := int(minU)
	max := int(maxU)
	vc := min
	// handle cases with irrational arguments. Max < Min = min
	if max > min {
		vc = rand.Intn(max-min) + min
	}
	svs := make([]*structs.VariableDecrypted, vc)
	paths := make(map[string]*structs.VariableDecrypted, vc)
	for i := 0; i < vc; i++ {
		nv := Variable()
		// There is an extremely rare chance of path collision because the mock
		// variables generate their paths randomly. This check will add
		// an extra component on conflict to (ideally) disambiguate them.
		if _, found := paths[nv.Path]; found {
			nv.Path = nv.Path + "/" + fmt.Sprint(time.Now().UnixNano())
		}
		paths[nv.Path] = nv
		svs[i] = nv
	}
	return paths
}

func (svs MockVariables) ListPaths() []string {
	out := make([]string, 0, len(svs))
	for _, sv := range svs {
		out = append(out, sv.Path)
	}
	sort.Strings(out)
	return out
}

func (svs MockVariables) List() []*structs.VariableDecrypted {
	out := make([]*structs.VariableDecrypted, 0, len(svs))
	for _, p := range svs.ListPaths() {
		pc := svs[p].Copy()
		out = append(out, &pc)
	}
	return out
}

type MockVariablesEncrypted map[string]*structs.VariableEncrypted

func VariableEncrypted() *structs.VariableEncrypted {
	return &structs.VariableEncrypted{
		VariableMetadata: mockVariableMetadata(),
		VariableData: structs.VariableData{
			KeyID: "foo",
			Data:  []byte("foo"),
		},
	}
}

// Variables returns a random number of variables between min
// and max inclusive.
func VariablesEncrypted(minU, maxU uint8) MockVariablesEncrypted {
	// the unsignedness of the args is to prevent goofy parameters, they're
	// easier to work with as ints in this code.
	min := int(minU)
	max := int(maxU)
	vc := min
	// handle cases with irrational arguments. Max < Min = min
	if max > min {
		vc = rand.Intn(max-min) + min
	}
	svs := make([]*structs.VariableEncrypted, vc)
	paths := make(map[string]*structs.VariableEncrypted, vc)
	for i := 0; i < vc; i++ {
		nv := VariableEncrypted()
		// There is an extremely rare chance of path collision because the mock
		// variables generate their paths randomly. This check will add
		// an extra component on conflict to (ideally) disambiguate them.
		if _, found := paths[nv.Path]; found {
			nv.Path = nv.Path + "/" + fmt.Sprint(time.Now().UnixNano())
		}
		paths[nv.Path] = nv
		svs[i] = nv
	}
	return paths
}

func (svs MockVariablesEncrypted) ListPaths() []string {
	out := make([]string, 0, len(svs))
	for _, sv := range svs {
		out = append(out, sv.Path)
	}
	sort.Strings(out)
	return out
}

func (svs MockVariablesEncrypted) List() []*structs.VariableEncrypted {
	out := make([]*structs.VariableEncrypted, 0, len(svs))
	for _, p := range svs.ListPaths() {
		pc := svs[p].Copy()
		out = append(out, &pc)
	}
	return out
}

func mockVariableMetadata() structs.VariableMetadata {
	envs := []string{"dev", "test", "prod"}
	envIdx := rand.Intn(3)
	env := envs[envIdx]
	domain := fake.DomainName()

	out := structs.VariableMetadata{
		Namespace:   "default",
		Path:        strings.ReplaceAll(env+"."+domain, ".", "/"),
		CreateIndex: uint64(rand.Intn(100) + 100),
		CreateTime:  fake.DateRange(time.Now().AddDate(0, -1, 0), time.Now()).UnixNano(),
	}
	out.ModifyIndex = out.CreateIndex
	out.ModifyTime = out.CreateTime

	// Flip a coin to see if we should return a "modified" object
	if fake.Bool() {
		out.ModifyTime = fake.DateRange(time.Unix(0, out.CreateTime), time.Now()).UnixNano()
		out.ModifyIndex = out.CreateIndex + uint64(rand.Intn(100))
	}
	return out
}

func ACLRole() *structs.ACLRole {
	role := structs.ACLRole{
		ID:          uuid.Generate(),
		Name:        fmt.Sprintf("acl-role-%s", uuid.Short()),
		Description: "mocked-test-acl-role",
		Policies: []*structs.ACLRolePolicyLink{
			{Name: "mocked-test-policy-1"},
			{Name: "mocked-test-policy-2"},
		},
		CreateIndex: 10,
		ModifyIndex: 10,
	}
	role.SetHash()
	return &role
}
