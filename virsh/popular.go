package main

// Popular marks the commands and flags people use most, so the TUI can
// highlight them. Curated from virsh cheat sheets and guides (InventiveHQ,
// Computing for Geeks, Red Hat's virtualization guide, the virsh man page);
// TestPopularExists checks every entry against the installed virsh.
type Popular struct {
	Flags   []string // option names, without --
	Example string   // a typical invocation
}

var popular = map[string]Popular{
	// Inventory and inspection
	"list":       {[]string{"all", "name", "inactive", "state-running"}, "virsh list --all"},
	"dominfo":    {nil, "virsh dominfo vm1"},
	"domstate":   {[]string{"reason"}, "virsh domstate vm1 --reason"},
	"dumpxml":    {[]string{"inactive", "security-info"}, "virsh dumpxml vm1 > vm1.xml"},
	"domifaddr":  {[]string{"source"}, "virsh domifaddr vm1 --source agent"},
	"domblklist": {[]string{"details"}, "virsh domblklist vm1 --details"},
	"vncdisplay": {nil, "virsh vncdisplay vm1"},
	"nodeinfo":   {nil, "virsh nodeinfo"},
	"version":    {nil, "virsh version"},

	// Lifecycle
	"start":    {[]string{"console"}, "virsh start vm1"},
	"shutdown": {[]string{"mode"}, "virsh shutdown vm1"},
	"reboot":   {nil, "virsh reboot vm1"},
	"destroy":  {[]string{"graceful"}, "virsh destroy vm1   # hard power-off"},
	"suspend":  {nil, "virsh suspend vm1"},
	"resume":   {nil, "virsh resume vm1"},
	"console":  {[]string{"force"}, "virsh console vm1   # leave with ctrl+]"},
	"save":     {nil, "virsh save vm1 vm1.state"},
	"restore":  {nil, "virsh restore vm1.state"},

	// Persistent configuration
	"define":    {nil, "virsh define vm1.xml"},
	"undefine":  {[]string{"remove-all-storage", "nvram", "snapshots-metadata"}, "virsh undefine vm1 --remove-all-storage --nvram"},
	"edit":      {nil, "virsh edit vm1"},
	"autostart": {[]string{"disable"}, "virsh autostart vm1"},
	"domrename": {nil, "virsh domrename vm1 vm2"},

	// Resources and devices
	"setmem":           {[]string{"config", "live"}, "virsh setmem vm1 4G --config"},
	"setmaxmem":        {[]string{"config"}, "virsh setmaxmem vm1 8G --config"},
	"setvcpus":         {[]string{"config", "live", "maximum"}, "virsh setvcpus vm1 4 --config"},
	"attach-disk":      {[]string{"target", "subdriver", "persistent", "config", "live"}, "virsh attach-disk vm1 /var/lib/libvirt/images/data.qcow2 vdb --subdriver qcow2 --persistent"},
	"detach-disk":      {[]string{"persistent"}, "virsh detach-disk vm1 vdb --persistent"},
	"attach-interface": {[]string{"type", "source", "model", "config"}, "virsh attach-interface vm1 network default --model virtio --config"},
	"change-media":     {[]string{"eject", "insert", "config"}, "virsh change-media vm1 sda /path/to.iso --insert"},
	"migrate":          {[]string{"live", "persistent", "undefinesource", "copy-storage-all", "verbose"}, "virsh migrate --live --persistent vm1 qemu+ssh://host2/system"},

	// Snapshots
	"snapshot-create-as": {[]string{"name", "description", "disk-only", "atomic"}, "virsh snapshot-create-as vm1 --name before-upgrade"},
	"snapshot-list":      {[]string{"tree"}, "virsh snapshot-list vm1 --tree"},
	"snapshot-revert":    {[]string{"running"}, "virsh snapshot-revert vm1 before-upgrade"},
	"snapshot-delete":    {nil, "virsh snapshot-delete vm1 before-upgrade"},

	// Networking
	"net-list":        {[]string{"all"}, "virsh net-list --all"},
	"net-start":       {nil, "virsh net-start default"},
	"net-autostart":   {nil, "virsh net-autostart default"},
	"net-dhcp-leases": {nil, "virsh net-dhcp-leases default"},
	"net-dumpxml":     {nil, "virsh net-dumpxml default"},

	// Storage
	"pool-list":     {[]string{"all", "details"}, "virsh pool-list --all --details"},
	"pool-refresh":  {nil, "virsh pool-refresh default"},
	"vol-list":      {[]string{"details"}, "virsh vol-list default --details"},
	"vol-create-as": {[]string{"format"}, "virsh vol-create-as default disk.qcow2 20G --format qcow2"},
	"vol-delete":    {[]string{"pool"}, "virsh vol-delete disk.qcow2 --pool default"},
}

func IsPopular(cmd string) bool {
	_, ok := popular[cmd]
	return ok
}

func IsPopularFlag(cmd, opt string) bool {
	for _, f := range popular[cmd].Flags {
		if f == opt {
			return true
		}
	}
	return false
}
